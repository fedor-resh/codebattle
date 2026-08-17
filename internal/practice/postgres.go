package practice

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Problems(ctx context.Context, userID string) ([]ProblemSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT latest.slug, latest.title, latest.difficulty, latest.problem_class,
			COALESCE(latest.requirements, '{}'::jsonb)::text,
			session.solved_at IS NOT NULL
		FROM (
			SELECT DISTINCT ON (slug) slug, title, difficulty, problem_class, requirements, id
			FROM problem_versions
			ORDER BY slug, version DESC
		) latest
		LEFT JOIN practice_sessions session
			ON session.user_id = $1 AND session.problem_version_id = latest.id
		ORDER BY latest.title
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	problems := make([]ProblemSummary, 0)
	for rows.Next() {
		var item ProblemSummary
		var requirements string
		if err := rows.Scan(
			&item.Slug, &item.Title, &item.Difficulty, &item.ProblemClass,
			&requirements, &item.Solved,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(requirements), &item.Requirements); err != nil {
			return nil, err
		}
		problems = append(problems, item)
	}
	return problems, rows.Err()
}

func (r *PostgresRepository) StartSession(ctx context.Context, id, userID, slug string) (string, error) {
	var sessionID string
	err := r.pool.QueryRow(ctx, `
		WITH latest AS (
			SELECT DISTINCT ON (slug) id, starter_code
			FROM problem_versions
			WHERE slug = $3
			ORDER BY slug, version DESC
		)
		INSERT INTO practice_sessions (
			id, user_id, problem_version_id, source_code, revision, created_at, updated_at
		)
		SELECT $1, $2, latest.id, latest.starter_code, 0, now(), now()
		FROM latest
		ON CONFLICT (user_id, problem_version_id) DO UPDATE SET updated_at = now()
		RETURNING id
	`, id, userID, slug).Scan(&sessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrProblemNotFound
	}
	if err != nil {
		return "", err
	}
	return sessionID, nil
}

func (r *PostgresRepository) Session(ctx context.Context, userID, sessionID string) (Session, error) {
	var session Session
	var requirements string
	var publicTests string
	err := r.pool.QueryRow(ctx, `
		SELECT session.id, session.source_code, session.revision, session.solved_at,
			problem.id, problem.slug, problem.title, problem.difficulty, problem.problem_class,
			COALESCE(problem.requirements, '{}'::jsonb)::text,
			problem.statement_markdown, problem.function_signature, problem.starter_code,
			COALESCE(problem.public_tests, '[]'::jsonb)::text,
			problem.time_limit_ms, problem.memory_limit_mb
		FROM practice_sessions session
		JOIN problem_versions problem ON problem.id = session.problem_version_id
		WHERE session.id = $1 AND session.user_id = $2
	`, sessionID, userID).Scan(
		&session.ID, &session.SourceCode, &session.Revision, &session.SolvedAt,
		&session.Problem.ID, &session.Problem.Slug, &session.Problem.Title,
		&session.Problem.Difficulty, &session.Problem.ProblemClass, &requirements,
		&session.Problem.Statement, &session.Problem.FunctionSignature, &session.Problem.StarterCode,
		&publicTests, &session.Problem.TimeLimitMS, &session.Problem.MemoryLimitMB,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, err
	}
	session.Problem.PublicTests = []byte(publicTests)
	if err := json.Unmarshal([]byte(requirements), &session.Problem.Requirements); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (r *PostgresRepository) UpdateCode(
	ctx context.Context,
	sessionID, userID, source string,
	revision int64,
	now time.Time,
) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE practice_sessions
		SET source_code = $3, revision = $4, updated_at = $5
		WHERE id = $1 AND user_id = $2 AND revision < $4
	`, sessionID, userID, source, revision, now)
	if err != nil {
		return err
	}
	if result.RowsAffected() > 0 {
		return nil
	}
	var exists bool
	if err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM practice_sessions WHERE id = $1 AND user_id = $2
		)
	`, sessionID, userID).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return ErrStaleRevision
	}
	return ErrSessionNotFound
}
