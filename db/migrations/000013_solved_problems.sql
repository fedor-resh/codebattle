CREATE TABLE IF NOT EXISTS solved_problems (
    user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_slug text NOT NULL,
    first_solved_at timestamptz NOT NULL DEFAULT now(),
    last_solved_at timestamptz NOT NULL DEFAULT now(),
    solved_count integer NOT NULL DEFAULT 1 CHECK (solved_count > 0),
    PRIMARY KEY (user_id, problem_slug)
);

CREATE INDEX IF NOT EXISTS solved_problems_slug_idx ON solved_problems (problem_slug);

INSERT INTO solved_problems (user_id, problem_slug, first_solved_at, last_solved_at, solved_count)
SELECT session.user_id, problem.slug, MIN(session.solved_at), MAX(session.solved_at), 1
FROM practice_sessions session
JOIN problem_versions problem ON problem.id = session.problem_version_id
WHERE session.solved_at IS NOT NULL
GROUP BY session.user_id, problem.slug
ON CONFLICT (user_id, problem_slug) DO NOTHING;
