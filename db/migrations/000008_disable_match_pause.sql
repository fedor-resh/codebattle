UPDATE matches
SET state = COALESCE(paused_from_state, 'active'),
    paused_from_state = NULL,
    paused_at = NULL
WHERE state = 'paused';
