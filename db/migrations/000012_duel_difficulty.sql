ALTER TABLE invitations
    ADD COLUMN IF NOT EXISTS difficulty text,
    DROP CONSTRAINT IF EXISTS invitations_difficulty_check,
    ADD CONSTRAINT invitations_difficulty_check
        CHECK (difficulty IS NULL OR difficulty IN ('easy', 'medium', 'hard'));

ALTER TABLE matches
    ADD COLUMN IF NOT EXISTS difficulty text,
    DROP CONSTRAINT IF EXISTS matches_difficulty_check,
    ADD CONSTRAINT matches_difficulty_check
        CHECK (difficulty IS NULL OR difficulty IN ('easy', 'medium', 'hard'));
