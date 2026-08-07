CREATE TABLE IF NOT EXISTS matches (
    id text PRIMARY KEY,
    player_one_id text NOT NULL REFERENCES users(id),
    player_two_id text NOT NULL REFERENCES users(id),
    player_one_score integer NOT NULL DEFAULT 0 CHECK (player_one_score >= 0),
    player_two_score integer NOT NULL DEFAULT 0 CHECK (player_two_score >= 0),
    state text NOT NULL DEFAULT 'active'
        CHECK (state IN ('active', 'waiting_ready', 'paused', 'ended')),
    created_at timestamptz NOT NULL DEFAULT now(),
    ended_at timestamptz,
    ended_by text REFERENCES users(id),
    CONSTRAINT matches_players_differ CHECK (player_one_id <> player_two_id)
);

CREATE INDEX IF NOT EXISTS matches_player_one_state_idx ON matches (player_one_id, state);
CREATE INDEX IF NOT EXISTS matches_player_two_state_idx ON matches (player_two_id, state);

CREATE TABLE IF NOT EXISTS invitations (
    id text PRIMARY KEY,
    sender_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status text NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'declined', 'expired', 'cancelled')),
    expires_at timestamptz NOT NULL,
    match_id text REFERENCES matches(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    responded_at timestamptz,
    CONSTRAINT invitations_users_differ CHECK (sender_id <> receiver_id)
);

CREATE INDEX IF NOT EXISTS invitations_pending_sender_idx
    ON invitations (sender_id, expires_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS invitations_pending_receiver_idx
    ON invitations (receiver_id, expires_at) WHERE status = 'pending';
