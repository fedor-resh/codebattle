ALTER TABLE matches
    ADD COLUMN IF NOT EXISTS round_number integer NOT NULL DEFAULT 1 CHECK (round_number > 0),
    ADD COLUMN IF NOT EXISTS round_winner_id text REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS player_one_ready boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS player_two_ready boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS winning_source_code text,
    ADD COLUMN IF NOT EXISTS problem_history jsonb NOT NULL DEFAULT '[]'::jsonb;

UPDATE matches
SET problem_history = jsonb_build_array(problem_version_id)
WHERE problem_version_id IS NOT NULL AND problem_history = '[]'::jsonb;

CREATE TABLE IF NOT EXISTS match_code_snapshots (
    match_id text NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    problem_version_id text NOT NULL REFERENCES problem_versions(id),
    source_code text NOT NULL CHECK (octet_length(source_code) <= 65536),
    revision bigint NOT NULL CHECK (revision >= 0),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (match_id, user_id)
);
