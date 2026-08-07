package duels

import (
	"context"
	"errors"
	"testing"
	"time"
)

type repositoryStub struct {
	createdID         string
	createdSenderID   string
	createdReceiverID string
	createdAt         time.Time
	expiresAt         time.Time
}

func (r *repositoryStub) CreateInvitation(
	_ context.Context,
	id, senderID, receiverID string,
	createdAt, expiresAt time.Time,
) (Invitation, error) {
	r.createdID = id
	r.createdSenderID = senderID
	r.createdReceiverID = receiverID
	r.createdAt = createdAt
	r.expiresAt = expiresAt
	return Invitation{ID: id, ExpiresAt: expiresAt}, nil
}

func (*repositoryStub) State(context.Context, string, time.Time) (State, error) {
	return State{}, nil
}

func (*repositoryStub) AcceptInvitation(context.Context, string, string, string, time.Time) (Match, error) {
	return Match{}, nil
}

func (*repositoryStub) DeclineInvitation(context.Context, string, string, time.Time) error {
	return nil
}

func (*repositoryStub) Match(context.Context, string, string) (Match, error) {
	return Match{}, nil
}

func (*repositoryStub) LeaveMatch(context.Context, string, string, time.Time) error {
	return nil
}

func (*repositoryStub) UpdateCode(context.Context, string, string, string, int64, time.Time) error {
	return nil
}

func (*repositoryStub) Ready(context.Context, string, string, time.Time) (Match, error) {
	return Match{}, nil
}

func TestCreateInvitationUsesThirtySecondTTL(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository)
	now := time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	invitation, err := service.CreateInvitation(context.Background(), "sender", "receiver")
	if err != nil {
		t.Fatal(err)
	}
	if repository.createdID == "" || invitation.ID != repository.createdID {
		t.Fatal("invitation ID was not generated")
	}
	if repository.createdSenderID != "sender" || repository.createdReceiverID != "receiver" {
		t.Fatalf("players = %q -> %q", repository.createdSenderID, repository.createdReceiverID)
	}
	if got := repository.expiresAt.Sub(repository.createdAt); got != 30*time.Second {
		t.Fatalf("TTL = %s, want 30s", got)
	}
}

func TestCreateInvitationRejectsSelf(t *testing.T) {
	service := NewService(&repositoryStub{})
	_, err := service.CreateInvitation(context.Background(), "same-user", "same-user")
	if !errors.Is(err, ErrSelfInvitation) {
		t.Fatalf("error = %v, want ErrSelfInvitation", err)
	}
}
