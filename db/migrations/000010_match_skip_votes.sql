ALTER TABLE matches
    ADD COLUMN IF NOT EXISTS player_one_skip boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS player_two_skip boolean NOT NULL DEFAULT false;
