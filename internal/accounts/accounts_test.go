package accounts

import (
	"context"
	"errors"
	"testing"
	"time"
)

type memoryStore struct {
	users    map[string]UserRecord
	sessions map[string]string
}

func newMemoryStore() *memoryStore {
	return &memoryStore{users: map[string]UserRecord{}, sessions: map[string]string{}}
}

func (s *memoryStore) CreateUser(_ context.Context, user UserRecord) error {
	if _, exists := s.users[user.UsernameKey]; exists {
		return ErrUsernameTaken
	}
	s.users[user.UsernameKey] = user
	return nil
}

func (s *memoryStore) UserByUsernameKey(_ context.Context, key string) (UserRecord, error) {
	user, exists := s.users[key]
	if !exists {
		return UserRecord{}, ErrNotFound
	}
	return user, nil
}

func (s *memoryStore) CreateSession(_ context.Context, session SessionRecord) error {
	s.sessions[string(session.TokenHash)] = session.UserID
	return nil
}

func (s *memoryStore) UserBySessionHash(_ context.Context, hash []byte, _ time.Time) (UserRecord, error) {
	userID, exists := s.sessions[string(hash)]
	if !exists {
		return UserRecord{}, ErrNotFound
	}
	for _, user := range s.users {
		if user.ID == userID {
			return user, nil
		}
	}
	return UserRecord{}, ErrNotFound
}

func (s *memoryStore) DeleteSession(_ context.Context, hash []byte) error {
	delete(s.sessions, string(hash))
	return nil
}

func (s *memoryStore) TouchUser(_ context.Context, _ string, _ time.Time) error { return nil }
func (s *memoryStore) ListUsers(context.Context, string, int, int) ([]User, bool, error) {
	return nil, false, nil
}

func TestPasswordHash(t *testing.T) {
	t.Parallel()
	hash, err := HashPassword("secret123")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword("secret123", hash) {
		t.Fatal("correct password was rejected")
	}
	if CheckPassword("wrong-password", hash) {
		t.Fatal("wrong password was accepted")
	}
}

func TestRegisterLoginAndAuthenticate(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store)
	ctx := context.Background()

	registered, err := service.Register(ctx, "Go_Player", "secret123")
	if err != nil {
		t.Fatal(err)
	}
	loggedIn, err := service.Login(ctx, "go_player", "secret123")
	if err != nil {
		t.Fatal(err)
	}
	if loggedIn.User.ID != registered.User.ID {
		t.Fatalf("login user ID = %q, want %q", loggedIn.User.ID, registered.User.ID)
	}
	authenticated, err := service.Authenticate(ctx, loggedIn.Token)
	if err != nil {
		t.Fatal(err)
	}
	if authenticated.Username != "Go_Player" {
		t.Fatalf("username = %q, want Go_Player", authenticated.Username)
	}
}

func TestRegisterValidationAndUniqueness(t *testing.T) {
	service := NewService(newMemoryStore())
	ctx := context.Background()

	if _, err := service.Register(ctx, "bad name", "secret123"); !errors.Is(err, ErrInvalidUsername) {
		t.Fatalf("invalid username error = %v", err)
	}
	if _, err := service.Register(ctx, "valid_name", "short"); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("invalid password error = %v", err)
	}
	if _, err := service.Register(ctx, "Player", "secret123"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Register(ctx, "player", "secret123"); !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("duplicate username error = %v", err)
	}
}
