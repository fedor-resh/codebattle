ALTER TABLE match_code_snapshots
    ADD COLUMN IF NOT EXISTS cursor_line integer NOT NULL DEFAULT 1
        CHECK (cursor_line > 0),
    ADD COLUMN IF NOT EXISTS cursor_column integer NOT NULL DEFAULT 1
        CHECK (cursor_column > 0);
