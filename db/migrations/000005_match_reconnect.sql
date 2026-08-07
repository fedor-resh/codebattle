ALTER TABLE matches
    ADD COLUMN IF NOT EXISTS paused_at timestamptz,
    ADD COLUMN IF NOT EXISTS paused_from_state text
        CHECK (paused_from_state IN ('active', 'waiting_ready'));
