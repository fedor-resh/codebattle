package duels

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"codebattle.local/codebattle/internal/problems"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateInvitation(
	ctx context.Context,
	id, senderID, receiverID string,
	problemClass problems.Class,
	now, expiresAt time.Time,
) (Invitation, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Invitation{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	players, err := lockPlayers(ctx, tx, senderID, receiverID)
	if err != nil {
		return Invitation{}, err
	}
	if players[receiverID].LastSeenAt.Before(now.Add(-45 * time.Second)) {
		return Invitation{}, ErrUserUnavailable
	}

	if _, err := tx.Exec(ctx, `
		UPDATE invitations SET status = 'expired', responded_at = $1
		WHERE status = 'pending' AND expires_at <= $1
			AND (sender_id IN ($2, $3) OR receiver_id IN ($2, $3))
	`, now, senderID, receiverID); err != nil {
		return Invitation{}, err
	}

	var unavailable bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM matches
			WHERE state <> 'ended'
				AND (player_one_id IN ($1, $2) OR player_two_id IN ($1, $2))
		)
	`, senderID, receiverID).Scan(&unavailable); err != nil {
		return Invitation{}, err
	}
	if unavailable {
		return Invitation{}, ErrUserUnavailable
	}

	var hasInvitation bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM invitations
			WHERE status = 'pending' AND expires_at > $3
				AND (sender_id IN ($1, $2) OR receiver_id IN ($1, $2))
		)
	`, senderID, receiverID, now).Scan(&hasInvitation); err != nil {
		return Invitation{}, err
	}
	if hasInvitation {
		return Invitation{}, ErrInvitationBusy
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO invitations (id, sender_id, receiver_id, problem_class, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, senderID, receiverID, problemClass, expiresAt, now); err != nil {
		return Invitation{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Invitation{}, err
	}
	return Invitation{
		ID:           id,
		Sender:       players[senderID].Player,
		Receiver:     players[receiverID].Player,
		Status:       "pending",
		ProblemClass: problemClass,
		ExpiresAt:    expiresAt,
	}, nil
}

func (r *PostgresRepository) State(ctx context.Context, userID string, now time.Time) (State, error) {
	if _, err := r.pool.Exec(ctx, `
		UPDATE invitations SET status = 'expired', responded_at = $1
		WHERE status = 'pending' AND expires_at <= $1
	`, now); err != nil {
		return State{}, err
	}

	if match, err := r.activeMatch(ctx, userID); err == nil {
		return State{Match: &match}, nil
	} else if !errors.Is(err, ErrMatchNotFound) {
		return State{}, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT i.id, i.sender_id, sender.username, i.receiver_id, receiver.username,
			i.status, i.problem_class, i.expires_at
		FROM invitations i
		JOIN users sender ON sender.id = i.sender_id
		JOIN users receiver ON receiver.id = i.receiver_id
		WHERE i.status = 'pending' AND i.expires_at > $2
			AND (i.sender_id = $1 OR i.receiver_id = $1)
		ORDER BY i.created_at DESC
	`, userID, now)
	if err != nil {
		return State{}, err
	}
	defer rows.Close()

	var state State
	for rows.Next() {
		invitation, err := scanInvitation(rows)
		if err != nil {
			return State{}, err
		}
		if invitation.Receiver.ID == userID {
			state.Incoming = &invitation
		} else {
			state.Outgoing = &invitation
		}
	}
	return state, rows.Err()
}

func (r *PostgresRepository) AcceptInvitation(
	ctx context.Context,
	matchID, invitationID, userID string,
	now time.Time,
) (Match, error) {
	var senderID, receiverID string
	var problemClass problems.Class
	err := r.pool.QueryRow(ctx,
		"SELECT sender_id, receiver_id, problem_class FROM invitations WHERE id = $1",
		invitationID,
	).Scan(&senderID, &receiverID, &problemClass)
	if errors.Is(err, pgx.ErrNoRows) {
		return Match{}, ErrInvitationGone
	}
	if err != nil {
		return Match{}, err
	}
	if receiverID != userID {
		return Match{}, ErrForbidden
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Match{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	players, err := lockPlayers(ctx, tx, senderID, receiverID)
	if err != nil {
		return Match{}, err
	}

	var status string
	var expiresAt time.Time
	if err := tx.QueryRow(ctx, `
		SELECT status, problem_class, expires_at FROM invitations WHERE id = $1 FOR UPDATE
	`, invitationID).Scan(&status, &problemClass, &expiresAt); err != nil {
		return Match{}, err
	}
	if status != "pending" || !expiresAt.After(now) {
		if status == "pending" {
			_, _ = tx.Exec(ctx, "UPDATE invitations SET status = 'expired', responded_at = $2 WHERE id = $1", invitationID, now)
			_ = tx.Commit(ctx)
		}
		return Match{}, ErrInvitationGone
	}
	if players[senderID].LastSeenAt.Before(now.Add(-45*time.Second)) ||
		players[receiverID].LastSeenAt.Before(now.Add(-45*time.Second)) {
		return Match{}, ErrUserUnavailable
	}

	var busy bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM matches WHERE state <> 'ended'
			AND (player_one_id IN ($1, $2) OR player_two_id IN ($1, $2))
		)
	`, senderID, receiverID).Scan(&busy); err != nil {
		return Match{}, err
	}
	if busy {
		return Match{}, ErrUserUnavailable
	}
	var problemVersionID string
	if err := tx.QueryRow(ctx, `
		SELECT id FROM (
			SELECT DISTINCT ON (slug) id, slug, problem_class
			FROM problem_versions
			ORDER BY slug, version DESC
		) latest
		WHERE problem_class = $1
		ORDER BY random() LIMIT 1
	`, problemClass).Scan(&problemVersionID); errors.Is(err, pgx.ErrNoRows) {
		return Match{}, ErrProblemsMissing
	} else if err != nil {
		return Match{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO matches (
			id, player_one_id, player_two_id, problem_class, problem_version_id,
			problem_history, state, created_at
		)
		VALUES ($1, $2, $3, $4, $5, jsonb_build_array($5::text), 'active', $6)
	`, matchID, senderID, receiverID, problemClass, problemVersionID, now); err != nil {
		return Match{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE invitations
		SET status = 'accepted', match_id = $2, responded_at = $3
		WHERE id = $1
	`, invitationID, matchID, now); err != nil {
		return Match{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Match{}, err
	}
	return r.Match(ctx, matchID, userID)
}

func (r *PostgresRepository) DeclineInvitation(
	ctx context.Context,
	invitationID, userID string,
	now time.Time,
) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE invitations SET status = 'declined', responded_at = $3
		WHERE id = $1 AND receiver_id = $2 AND status = 'pending' AND expires_at > $3
	`, invitationID, userID, now)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrInvitationGone
	}
	return nil
}

func (r *PostgresRepository) Match(ctx context.Context, matchID, userID string) (Match, error) {
	return scanMatch(r.pool.QueryRow(ctx, matchQuery+`
		WHERE m.id = $1 AND (m.player_one_id = $2 OR m.player_two_id = $2)
	`, matchID, userID))
}

func (r *PostgresRepository) LeaveMatch(ctx context.Context, matchID, userID string, now time.Time) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE matches SET state = 'ended', ended_at = $3, ended_by = $2
		WHERE id = $1 AND state <> 'ended'
			AND (player_one_id = $2 OR player_two_id = $2)
	`, matchID, userID, now)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrMatchNotFound
	}
	return nil
}

func (r *PostgresRepository) UpdateCode(
	ctx context.Context,
	matchID, userID, source string,
	revision int64,
	cursorLine, cursorColumn int,
	now time.Time,
) error {
	result, err := r.pool.Exec(ctx, `
		INSERT INTO match_code_snapshots (
			match_id, user_id, problem_version_id, source_code, revision,
			cursor_line, cursor_column, updated_at
		)
		SELECT m.id, $2, m.problem_version_id, $3, $4, $5, $6, $7
		FROM matches m
		WHERE m.id = $1 AND m.state IN ('active', 'waiting_ready')
			AND (m.player_one_id = $2 OR m.player_two_id = $2)
		ON CONFLICT (match_id, user_id) DO UPDATE SET
			problem_version_id = EXCLUDED.problem_version_id,
			source_code = EXCLUDED.source_code,
			revision = EXCLUDED.revision,
			cursor_line = EXCLUDED.cursor_line,
			cursor_column = EXCLUDED.cursor_column,
			updated_at = EXCLUDED.updated_at
		WHERE match_code_snapshots.problem_version_id <> EXCLUDED.problem_version_id
			OR match_code_snapshots.revision < EXCLUDED.revision
	`, matchID, userID, source, revision, cursorLine, cursorColumn, now)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		var editable bool
		if err := r.pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM matches WHERE id = $1 AND state IN ('active', 'waiting_ready')
					AND (player_one_id = $2 OR player_two_id = $2)
			)
		`, matchID, userID).Scan(&editable); err != nil {
			return err
		}
		if editable {
			return ErrStaleRevision
		}
		return ErrRoundNotActive
	}
	return nil
}

func (r *PostgresRepository) Ready(ctx context.Context, matchID, userID string, now time.Time) (Match, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Match{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var playerOneID, playerTwoID, state, currentProblemID string
	var problemClass problems.Class
	var playerOneReady, playerTwoReady bool
	var history []byte
	err = tx.QueryRow(ctx, `
		SELECT player_one_id, player_two_id, state, problem_class, problem_version_id,
			player_one_ready, player_two_ready, problem_history
		FROM matches WHERE id = $1 FOR UPDATE
	`, matchID).Scan(
		&playerOneID, &playerTwoID, &state, &problemClass, &currentProblemID,
		&playerOneReady, &playerTwoReady, &history,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Match{}, ErrMatchNotFound
	}
	if err != nil {
		return Match{}, err
	}
	if userID != playerOneID && userID != playerTwoID {
		return Match{}, ErrForbidden
	}
	if state != "waiting_ready" {
		return Match{}, ErrRoundNotActive
	}
	if userID == playerOneID {
		playerOneReady = true
	} else {
		playerTwoReady = true
	}

	if !playerOneReady || !playerTwoReady {
		_, err = tx.Exec(ctx, `
			UPDATE matches SET player_one_ready = $2, player_two_ready = $3 WHERE id = $1
		`, matchID, playerOneReady, playerTwoReady)
	} else {
		var nextProblemID string
		err = tx.QueryRow(ctx, `
			WITH latest AS (
				SELECT DISTINCT ON (slug) id, slug, problem_class
				FROM problem_versions
				ORDER BY slug, version DESC
			), seen_slugs AS (
				SELECT slug FROM problem_versions
				WHERE id = ANY (SELECT jsonb_array_elements_text($1::jsonb))
			)
			SELECT id FROM latest
			WHERE problem_class = $2
			  AND slug NOT IN (SELECT slug FROM seen_slugs)
			ORDER BY random() LIMIT 1
		`, history, problemClass).Scan(&nextProblemID)
		resetHistory := false
		if errors.Is(err, pgx.ErrNoRows) {
			resetHistory = true
			err = tx.QueryRow(ctx, `
				SELECT id FROM (
					SELECT DISTINCT ON (slug) id, slug, problem_class
					FROM problem_versions
					ORDER BY slug, version DESC
				) latest
				WHERE problem_class = $2
				  AND slug <> (SELECT slug FROM problem_versions WHERE id = $1)
				ORDER BY random() LIMIT 1
			`, currentProblemID, problemClass).Scan(&nextProblemID)
		}
		if err != nil {
			return Match{}, ErrProblemsMissing
		}
		_, err = tx.Exec(ctx, `
			UPDATE matches SET
				state = 'active', round_number = round_number + 1,
				problem_version_id = $2, round_winner_id = NULL,
				player_one_ready = false, player_two_ready = false,
				winning_source_code = NULL,
				problem_history = CASE WHEN $3
					THEN jsonb_build_array($2::text)
					ELSE problem_history || jsonb_build_array($2::text) END
			WHERE id = $1
		`, matchID, nextProblemID, resetHistory)
		if err == nil {
			_, err = tx.Exec(ctx, "DELETE FROM match_code_snapshots WHERE match_id = $1", matchID)
		}
	}
	if err != nil {
		return Match{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Match{}, err
	}
	return r.Match(ctx, matchID, userID)
}

func (r *PostgresRepository) activeMatch(ctx context.Context, userID string) (Match, error) {
	return scanMatch(r.pool.QueryRow(ctx, matchQuery+`
		WHERE m.state <> 'ended' AND (m.player_one_id = $1 OR m.player_two_id = $1)
		ORDER BY m.created_at DESC LIMIT 1
	`, userID))
}

type lockedPlayer struct {
	Player
	LastSeenAt time.Time
}

func lockPlayers(ctx context.Context, tx pgx.Tx, firstID, secondID string) (map[string]lockedPlayer, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, username, last_seen_at FROM users
		WHERE id IN ($1, $2) ORDER BY id FOR UPDATE
	`, firstID, secondID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make(map[string]lockedPlayer, 2)
	for rows.Next() {
		var player lockedPlayer
		if err := rows.Scan(&player.ID, &player.Username, &player.LastSeenAt); err != nil {
			return nil, err
		}
		players[player.ID] = player
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(players) != 2 {
		return nil, ErrUserNotFound
	}
	return players, nil
}

const matchQuery = `
	SELECT m.id,
		m.player_one_id, player_one.username,
		m.player_two_id, player_two.username,
		m.player_one_score, m.player_two_score, m.state, m.problem_class,
		COALESCE(problem.id, ''), COALESCE(problem.slug, ''), COALESCE(problem.title, ''),
		COALESCE(problem.difficulty, ''), COALESCE(problem.problem_class, 'algorithms'),
		COALESCE(problem.requirements, '{}'::jsonb)::text,
		COALESCE(problem.statement_markdown, ''),
		COALESCE(problem.function_signature, ''), COALESCE(problem.starter_code, ''),
		COALESCE(problem.public_tests, '[]'::jsonb)::text,
		COALESCE(problem.time_limit_ms, 0), COALESCE(problem.memory_limit_mb, 0),
		m.round_number, COALESCE(m.round_winner_id, ''),
		m.player_one_ready, m.player_two_ready, COALESCE(m.winning_source_code, ''),
		COALESCE((
			SELECT jsonb_agg(jsonb_build_object(
				'user_id', snapshot.user_id,
				'problem_version_id', snapshot.problem_version_id,
				'source_code', snapshot.source_code,
				'revision', snapshot.revision,
				'cursor_line', snapshot.cursor_line,
				'cursor_column', snapshot.cursor_column
			) ORDER BY snapshot.user_id)
			FROM match_code_snapshots snapshot WHERE snapshot.match_id = m.id
		), '[]'::jsonb)::text,
		m.paused_at
	FROM matches m
	JOIN users player_one ON player_one.id = m.player_one_id
	JOIN users player_two ON player_two.id = m.player_two_id
	LEFT JOIN problem_versions problem ON problem.id = m.problem_version_id
`

type rowScanner interface {
	Scan(...any) error
}

func scanMatch(row rowScanner) (Match, error) {
	var match Match
	var problem Problem
	var publicTests string
	var requirements string
	var codeSnapshots string
	err := row.Scan(
		&match.ID,
		&match.PlayerOne.ID, &match.PlayerOne.Username,
		&match.PlayerTwo.ID, &match.PlayerTwo.Username,
		&match.PlayerOneScore, &match.PlayerTwoScore, &match.State, &match.ProblemClass,
		&problem.ID, &problem.Slug, &problem.Title, &problem.Difficulty,
		&problem.ProblemClass, &requirements,
		&problem.Statement, &problem.FunctionSignature, &problem.StarterCode,
		&publicTests, &problem.TimeLimitMS, &problem.MemoryLimitMB,
		&match.RoundNumber, &match.RoundWinnerID,
		&match.PlayerOneReady, &match.PlayerTwoReady, &match.WinningSource,
		&codeSnapshots,
		&match.PausedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Match{}, ErrMatchNotFound
	}
	if err != nil {
		return Match{}, err
	}
	if problem.ID != "" {
		problem.PublicTests = []byte(publicTests)
		if err := json.Unmarshal([]byte(requirements), &problem.Requirements); err != nil {
			return Match{}, err
		}
		match.Problem = &problem
	}
	if err := json.Unmarshal([]byte(codeSnapshots), &match.CodeSnapshots); err != nil {
		return Match{}, err
	}
	return match, nil
}

func scanInvitation(row rowScanner) (Invitation, error) {
	var invitation Invitation
	err := row.Scan(
		&invitation.ID,
		&invitation.Sender.ID, &invitation.Sender.Username,
		&invitation.Receiver.ID, &invitation.Receiver.Username,
		&invitation.Status, &invitation.ProblemClass, &invitation.ExpiresAt,
	)
	if err != nil {
		return Invitation{}, fmt.Errorf("scan invitation: %w", err)
	}
	return invitation, nil
}
