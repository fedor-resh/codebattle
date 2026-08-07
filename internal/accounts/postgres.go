package accounts

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) CreateUser(ctx context.Context, user UserRecord) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO users (id, username, username_key, password_hash, last_seen_at)
		VALUES ($1, $2, $3, $4, $5)
	`, user.ID, user.Username, user.UsernameKey, user.PasswordHash, user.LastSeenAt)
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return ErrUsernameTaken
	}
	return err
}

func (s *PostgresStore) UserByUsernameKey(ctx context.Context, usernameKey string) (UserRecord, error) {
	var user UserRecord
	err := s.pool.QueryRow(ctx, `
		SELECT id, username, username_key, password_hash, last_seen_at
		FROM users WHERE username_key = $1
	`, usernameKey).Scan(&user.ID, &user.Username, &user.UsernameKey, &user.PasswordHash, &user.LastSeenAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserRecord{}, ErrNotFound
	}
	return user, err
}

func (s *PostgresStore) CreateSession(ctx context.Context, session SessionRecord) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)
	`, session.TokenHash, session.UserID, session.ExpiresAt)
	return err
}

func (s *PostgresStore) UserBySessionHash(ctx context.Context, tokenHash []byte, now time.Time) (UserRecord, error) {
	var user UserRecord
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.username, u.username_key, u.password_hash, u.last_seen_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > $2
	`, tokenHash, now).Scan(&user.ID, &user.Username, &user.UsernameKey, &user.PasswordHash, &user.LastSeenAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserRecord{}, ErrNotFound
	}
	return user, err
}

func (s *PostgresStore) DeleteSession(ctx context.Context, tokenHash []byte) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM sessions WHERE token_hash = $1", tokenHash)
	return err
}

func (s *PostgresStore) TouchUser(ctx context.Context, userID string, at time.Time) error {
	_, err := s.pool.Exec(ctx, "UPDATE users SET last_seen_at = $2 WHERE id = $1", userID, at)
	return err
}

func (s *PostgresStore) ListUsers(ctx context.Context, query string, offset, limit int) ([]User, bool, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT u.id, u.username,
			CASE
				WHEN EXISTS (
					SELECT 1 FROM matches m WHERE m.state <> 'ended'
						AND (m.player_one_id = u.id OR m.player_two_id = u.id)
				) THEN 'online_busy'
				WHEN u.last_seen_at >= now() - interval '45 seconds' THEN 'online_available'
				ELSE 'offline'
			END AS status
		FROM users u
		WHERE ($1 = '' OR u.username_key LIKE '%' || lower($1) || '%')
		ORDER BY
			CASE
				WHEN EXISTS (
					SELECT 1 FROM matches m WHERE m.state <> 'ended'
						AND (m.player_one_id = u.id OR m.player_two_id = u.id)
				) THEN 1
				WHEN u.last_seen_at >= now() - interval '45 seconds' THEN 0
				ELSE 2
			END,
			u.username_key,
			u.id
		LIMIT $2 OFFSET $3
	`, query, limit+1, offset)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	users := make([]User, 0, limit+1)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Status); err != nil {
			return nil, false, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}
	hasMore := len(users) > limit
	if hasMore {
		users = users[:limit]
	}
	return users, hasMore, nil
}
