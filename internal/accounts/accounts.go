package accounts

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUsernameTaken      = errors.New("username is already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidCursor      = errors.New("invalid cursor")
	ErrNotFound           = errors.New("not found")
)

const sessionLifetime = 30 * 24 * time.Hour

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Status   string `json:"status"`
}

type UserRecord struct {
	ID           string
	Username     string
	UsernameKey  string
	PasswordHash string
	LastSeenAt   time.Time
}

type SessionRecord struct {
	TokenHash []byte
	UserID    string
	ExpiresAt time.Time
}

type Store interface {
	CreateUser(context.Context, UserRecord) error
	UserByUsernameKey(context.Context, string) (UserRecord, error)
	CreateSession(context.Context, SessionRecord) error
	UserBySessionHash(context.Context, []byte, time.Time) (UserRecord, error)
	DeleteSession(context.Context, []byte) error
	TouchUser(context.Context, string, time.Time) error
	ListUsers(context.Context, string, int, int) ([]User, bool, error)
}

type Service struct {
	store Store
	now   func() time.Time
}

type Session struct {
	Token     string
	ExpiresAt time.Time
	User      User
}

type UserPage struct {
	Items      []User `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func NewService(store Store) *Service {
	return &Service{store: store, now: time.Now}
}

func (s *Service) Register(ctx context.Context, username, password string) (Session, error) {
	if err := ValidateUsername(username); err != nil {
		return Session{}, fmt.Errorf("%w: %v", ErrInvalidUsername, err)
	}
	if err := ValidatePassword(password); err != nil {
		return Session{}, fmt.Errorf("%w: %v", ErrInvalidPassword, err)
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return Session{}, fmt.Errorf("hash password: %w", err)
	}

	now := s.now().UTC()
	record := UserRecord{
		ID:           randomString(16),
		Username:     username,
		UsernameKey:  strings.ToLower(username),
		PasswordHash: passwordHash,
		LastSeenAt:   now,
	}
	if err := s.store.CreateUser(ctx, record); err != nil {
		return Session{}, err
	}

	return s.newSession(ctx, record, now)
}

func (s *Service) Login(ctx context.Context, username, password string) (Session, error) {
	record, err := s.store.UserByUsernameKey(ctx, strings.ToLower(username))
	if errors.Is(err, ErrNotFound) {
		return Session{}, ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, err
	}
	if !CheckPassword(password, record.PasswordHash) {
		return Session{}, ErrInvalidCredentials
	}

	now := s.now().UTC()
	if err := s.store.TouchUser(ctx, record.ID, now); err != nil {
		return Session{}, err
	}
	return s.newSession(ctx, record, now)
}

func (s *Service) Authenticate(ctx context.Context, token string) (User, error) {
	if token == "" {
		return User{}, ErrUnauthorized
	}
	now := s.now().UTC()
	record, err := s.store.UserBySessionHash(ctx, tokenHash(token), now)
	if errors.Is(err, ErrNotFound) {
		return User{}, ErrUnauthorized
	}
	if err != nil {
		return User{}, err
	}
	if err := s.store.TouchUser(ctx, record.ID, now); err != nil {
		return User{}, err
	}
	return User{ID: record.ID, Username: record.Username, Status: "online_available"}, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.store.DeleteSession(ctx, tokenHash(token))
}

func (s *Service) Heartbeat(ctx context.Context, userID string) error {
	return s.store.TouchUser(ctx, userID, s.now().UTC())
}

func (s *Service) ListUsers(ctx context.Context, query, cursor string, limit int) (UserPage, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	offset, err := decodeCursor(cursor)
	if err != nil {
		return UserPage{}, ErrInvalidCursor
	}
	items, hasMore, err := s.store.ListUsers(ctx, strings.TrimSpace(query), offset, limit)
	if err != nil {
		return UserPage{}, err
	}
	page := UserPage{Items: items}
	if hasMore {
		page.NextCursor = encodeCursor(offset + limit)
	}
	return page, nil
}

func (s *Service) newSession(ctx context.Context, record UserRecord, now time.Time) (Session, error) {
	token := randomString(32)
	expiresAt := now.Add(sessionLifetime)
	if err := s.store.CreateSession(ctx, SessionRecord{
		TokenHash: tokenHash(token),
		UserID:    record.ID,
		ExpiresAt: expiresAt,
	}); err != nil {
		return Session{}, err
	}
	return Session{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      User{ID: record.ID, Username: record.Username, Status: "online_available"},
	}, nil
}

func randomString(size int) string {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		panic(fmt.Sprintf("secure random source failed: %v", err))
	}
	return base64.RawURLEncoding.EncodeToString(buffer)
}

func tokenHash(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

func encodeCursor(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func decodeCursor(cursor string) (int, error) {
	if cursor == "" {
		return 0, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}
	offset, err := strconv.Atoi(string(decoded))
	if err != nil || offset < 0 {
		return 0, ErrInvalidCursor
	}
	return offset, nil
}
