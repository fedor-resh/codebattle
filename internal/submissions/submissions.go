package submissions

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"codebattle.local/codebattle/internal/problems"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidSource      = errors.New("invalid source")
	ErrSourceTooLarge     = errors.New("source is too large")
	ErrMatchNotFound      = errors.New("match not found")
	ErrRateLimited        = errors.New("submission rate limited")
	ErrTooManyPending     = errors.New("too many pending submissions")
	ErrSubmissionNotFound = errors.New("submission not found")
)

type Submission struct {
	ID        string          `json:"id"`
	MatchID   string          `json:"match_id"`
	UserID    string          `json:"user_id"`
	Status    string          `json:"status"`
	Result    json.RawMessage `json:"result,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type Job struct {
	Submission
	ProblemVersionID string
	SourceCode       string
	PublicTests      []problems.PublicTest
	HiddenTestSource string
	TimeLimitMS      int
	MemoryLimitMB    int
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, userID, matchID, source string) (Submission, error) {
	if len([]byte(source)) > 64*1024 {
		return Submission{}, ErrSourceTooLarge
	}
	if err := problems.ValidateSolution(source, "Solve"); err != nil {
		return Submission{}, fmt.Errorf("%w: %v", ErrInvalidSource, err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Submission{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var problemVersionID string
	if err := tx.QueryRow(ctx, `
		SELECT problem_version_id FROM matches
		WHERE id = $1 AND state = 'active'
			AND (player_one_id = $2 OR player_two_id = $2)
		FOR UPDATE
	`, matchID, userID).Scan(&problemVersionID); errors.Is(err, pgx.ErrNoRows) {
		return Submission{}, ErrMatchNotFound
	} else if err != nil {
		return Submission{}, err
	}

	var pending int
	if err := tx.QueryRow(ctx, `
		SELECT count(*) FROM submissions
		WHERE match_id = $1 AND user_id = $2
			AND status IN ('queued', 'compiling', 'running')
	`, matchID, userID).Scan(&pending); err != nil {
		return Submission{}, err
	}
	if pending >= 3 {
		return Submission{}, ErrTooManyPending
	}

	var recent bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM submissions WHERE match_id = $1 AND user_id = $2
				AND created_at > now() - interval '2 seconds'
		)
	`, matchID, userID).Scan(&recent); err != nil {
		return Submission{}, err
	}
	if recent {
		return Submission{}, ErrRateLimited
	}

	submission := Submission{
		ID:        randomID(),
		MatchID:   matchID,
		UserID:    userID,
		Status:    "queued",
		CreatedAt: time.Now().UTC(),
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO submissions (id, match_id, user_id, problem_version_id, source_code, status, created_at)
		VALUES ($1, $2, $3, $4, $5, 'queued', $6)
	`, submission.ID, matchID, userID, problemVersionID, source, submission.CreatedAt); err != nil {
		return Submission{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Submission{}, err
	}
	return submission, nil
}

func (r *Repository) Get(ctx context.Context, submissionID, userID string) (Submission, error) {
	var submission Submission
	var result []byte
	err := r.pool.QueryRow(ctx, `
		SELECT s.id, s.match_id, s.user_id, s.status, COALESCE(s.result, '{}'::jsonb), s.created_at
		FROM submissions s
		JOIN matches m ON m.id = s.match_id
		WHERE s.id = $1 AND (m.player_one_id = $2 OR m.player_two_id = $2)
	`, submissionID, userID).Scan(
		&submission.ID, &submission.MatchID, &submission.UserID,
		&submission.Status, &result, &submission.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Submission{}, ErrSubmissionNotFound
	}
	if err != nil {
		return Submission{}, err
	}
	if string(result) != "{}" {
		submission.Result = result
	}
	return submission, nil
}

func (r *Repository) Claim(ctx context.Context) (Job, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Job{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var job Job
	var publicTests string
	err = tx.QueryRow(ctx, `
		SELECT s.id, s.match_id, s.user_id, s.status, s.created_at,
			s.problem_version_id, s.source_code, problem.public_tests::text,
			bundle.hidden_test_source,
			problem.time_limit_ms, problem.memory_limit_mb
		FROM submissions s
		JOIN problem_versions problem ON problem.id = s.problem_version_id
		JOIN judge.problem_bundles bundle ON bundle.problem_version_id = s.problem_version_id
		WHERE s.status = 'queued'
		ORDER BY s.created_at
		FOR UPDATE OF s SKIP LOCKED
		LIMIT 1
	`).Scan(
		&job.ID, &job.MatchID, &job.UserID, &job.Status, &job.CreatedAt,
		&job.ProblemVersionID, &job.SourceCode, &publicTests, &job.HiddenTestSource,
		&job.TimeLimitMS, &job.MemoryLimitMB,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Job{}, pgx.ErrNoRows
	}
	if err != nil {
		return Job{}, err
	}
	if err := json.Unmarshal([]byte(publicTests), &job.PublicTests); err != nil {
		return Job{}, fmt.Errorf("decode public tests for submission %s: %w", job.ID, err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE submissions SET status = 'compiling', claimed_at = now() WHERE id = $1
	`, job.ID); err != nil {
		return Job{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Job{}, err
	}
	job.Status = "compiling"
	return job, nil
}

func (r *Repository) SetStatus(ctx context.Context, submissionID, status string, result any) error {
	encoded, err := json.Marshal(result)
	if err != nil {
		return err
	}
	terminal := status != "compiling" && status != "running" && status != "queued"
	_, err = r.pool.Exec(ctx, `
		UPDATE submissions SET status = $2, result = $3,
			finished_at = CASE WHEN $4 THEN now() ELSE finished_at END
		WHERE id = $1
	`, submissionID, status, encoded, terminal)
	return err
}

func (r *Repository) Finish(ctx context.Context, job Job, status string, result any) error {
	encoded, err := json.Marshal(result)
	if err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
		UPDATE submissions SET status = $2, result = $3, finished_at = now() WHERE id = $1
	`, job.ID, status, encoded); err != nil {
		return err
	}
	if status == "accepted" {
		if _, err := tx.Exec(ctx, `
			UPDATE matches SET
				state = CASE WHEN state = 'paused' THEN 'paused' ELSE 'waiting_ready' END,
				paused_from_state = CASE WHEN state = 'paused' THEN 'waiting_ready' ELSE paused_from_state END,
				round_winner_id = $2,
				winning_source_code = $4,
				player_one_score = player_one_score + CASE WHEN player_one_id = $2 THEN 1 ELSE 0 END,
				player_two_score = player_two_score + CASE WHEN player_two_id = $2 THEN 1 ELSE 0 END
			WHERE id = $1
				AND (state = 'active' OR (state = 'paused' AND paused_from_state = 'active'))
				AND problem_version_id = $3
		`, job.MatchID, job.UserID, job.ProblemVersionID, job.SourceCode); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func randomID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer)
}
