CREATE TABLE IF NOT EXISTS problem_versions (
    id text PRIMARY KEY,
    slug text NOT NULL,
    version integer NOT NULL CHECK (version > 0),
    title text NOT NULL,
    difficulty text NOT NULL CHECK (difficulty IN ('easy', 'medium', 'hard')),
    function_name text NOT NULL,
    function_signature text NOT NULL,
    statement_markdown text NOT NULL,
    starter_code text NOT NULL,
    public_tests jsonb NOT NULL,
    time_limit_ms integer NOT NULL CHECK (time_limit_ms > 0),
    memory_limit_mb integer NOT NULL CHECK (memory_limit_mb > 0),
    content_hash text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (slug, version)
);

CREATE SCHEMA IF NOT EXISTS judge;

CREATE TABLE IF NOT EXISTS judge.problem_bundles (
    problem_version_id text PRIMARY KEY REFERENCES problem_versions(id) ON DELETE CASCADE,
    hidden_test_source text NOT NULL,
    reference_source text NOT NULL
);

ALTER TABLE matches
    ADD COLUMN IF NOT EXISTS problem_version_id text REFERENCES problem_versions(id);

CREATE TABLE IF NOT EXISTS submissions (
    id text PRIMARY KEY,
    match_id text NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_version_id text NOT NULL REFERENCES problem_versions(id),
    source_code text NOT NULL CHECK (octet_length(source_code) <= 65536),
    status text NOT NULL DEFAULT 'queued'
        CHECK (status IN (
            'queued', 'compiling', 'running', 'accepted', 'wrong_answer',
            'compile_error', 'runtime_error', 'time_limit', 'memory_limit',
            'internal_error'
        )),
    result jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    claimed_at timestamptz,
    finished_at timestamptz
);

CREATE INDEX IF NOT EXISTS submissions_queue_idx
    ON submissions (created_at) WHERE status = 'queued';
CREATE INDEX IF NOT EXISTS submissions_match_user_idx
    ON submissions (match_id, user_id, created_at DESC);
