package problems

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SeedResult struct {
	Inserted int
	Skipped  int
}

func Seed(ctx context.Context, pool *pgxpool.Pool, catalog []Problem) (SeedResult, error) {
	var result SeedResult
	for _, problem := range catalog {
		var existingHash string
		err := pool.QueryRow(ctx, `
			SELECT content_hash FROM problem_versions WHERE slug = $1 AND version = $2
		`, problem.Slug, problem.Version).Scan(&existingHash)
		if err == nil {
			if existingHash != problem.ContentHash {
				return result, fmt.Errorf("problem %s version %d is immutable and has a different hash", problem.Slug, problem.Version)
			}
			result.Skipped++
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return result, err
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return result, err
		}
		requirements, err := json.Marshal(problem.Requirements)
		if err != nil {
			_ = tx.Rollback(ctx)
			return result, fmt.Errorf("encode requirements for problem %s: %w", problem.Slug, err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO problem_versions (
				id, slug, version, title, difficulty, problem_class, requirements,
				function_name, function_signature,
				statement_markdown, starter_code, public_tests, time_limit_ms,
				memory_limit_mb, content_hash
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		`,
			problem.ID, problem.Slug, problem.Version, problem.Title, problem.Difficulty,
			problem.Class, requirements, problem.Function, problem.Signature, problem.Statement,
			problem.Starter, problem.PublicTestRaw, problem.TimeLimitMS, problem.MemoryLimitMB,
			problem.ContentHash,
		)
		if err == nil {
			_, err = tx.Exec(ctx, `
				INSERT INTO judge.problem_bundles (problem_version_id, hidden_test_source, reference_source)
				VALUES ($1, $2, $3)
			`, problem.ID, problem.HiddenTest, problem.Reference)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return result, fmt.Errorf("seed problem %s: %w", problem.Slug, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return result, err
		}
		result.Inserted++
	}
	return result, nil
}
