CREATE TABLE IF NOT EXISTS practice_sessions (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_version_id text NOT NULL REFERENCES problem_versions(id),
    source_code text NOT NULL DEFAULT '' CHECK (octet_length(source_code) <= 65536),
    revision bigint NOT NULL DEFAULT 0 CHECK (revision >= 0),
    solved_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, problem_version_id)
);

ALTER TABLE submissions
    ALTER COLUMN match_id DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS practice_session_id text REFERENCES practice_sessions(id) ON DELETE CASCADE;

ALTER TABLE submissions
    DROP CONSTRAINT IF EXISTS submissions_one_parent_check,
    ADD CONSTRAINT submissions_one_parent_check
        CHECK (num_nonnulls(match_id, practice_session_id) = 1);

CREATE INDEX IF NOT EXISTS submissions_practice_session_user_idx
    ON submissions (practice_session_id, user_id, created_at DESC);
