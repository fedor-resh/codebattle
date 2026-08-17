package duels

import (
	"context"
	"errors"
	"testing"
	"time"

	"codebattle.local/codebattle/internal/problems"
)

type repositoryStub struct {
	createdID           string
	createdSenderID     string
	createdReceiverID   string
	createdProblemClass problems.Class
	createdAt           time.Time
	expiresAt           time.Time
	updatedRevision     int64
	updatedCursorLine   int
	updatedCursorCol    int
}

func (r *repositoryStub) CreateInvitation(
	_ context.Context,
	id, senderID, receiverID string,
	problemClass problems.Class,
	createdAt, expiresAt time.Time,
) (Invitation, error) {
	r.createdID = id
	r.createdSenderID = senderID
	r.createdReceiverID = receiverID
	r.createdProblemClass = problemClass
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

func (r *repositoryStub) UpdateCode(
	_ context.Context,
	_, _, _ string,
	revision int64,
	cursorLine, cursorColumn int,
	_ time.Time,
) error {
	r.updatedRevision = revision
	r.updatedCursorLine = cursorLine
	r.updatedCursorCol = cursorColumn
	return nil
}

func (*repositoryStub) Ready(context.Context, string, string, time.Time) (Match, error) {
	return Match{}, nil
}

func (*repositoryStub) Skip(context.Context, string, string, time.Time) (Match, error) {
	return Match{}, nil
}

func TestCreateInvitationUsesThirtySecondTTL(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository)
	now := time.Date(2026, 8, 8, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	invitation, err := service.CreateInvitation(
		context.Background(), "sender", "receiver", problems.ClassConcurrency,
	)
	if err != nil {
		t.Fatal(err)
	}
	if repository.createdID == "" || invitation.ID != repository.createdID {
		t.Fatal("invitation ID was not generated")
	}
	if repository.createdSenderID != "sender" || repository.createdReceiverID != "receiver" {
		t.Fatalf("players = %q -> %q", repository.createdSenderID, repository.createdReceiverID)
	}
	if repository.createdProblemClass != problems.ClassConcurrency {
		t.Fatalf("problem class = %q", repository.createdProblemClass)
	}
	if got := repository.expiresAt.Sub(repository.createdAt); got != 30*time.Second {
		t.Fatalf("TTL = %s, want 30s", got)
	}
}

func TestCreateInvitationRejectsSelf(t *testing.T) {
	service := NewService(&repositoryStub{})
	_, err := service.CreateInvitation(
		context.Background(), "same-user", "same-user", problems.ClassAlgorithms,
	)
	if !errors.Is(err, ErrSelfInvitation) {
		t.Fatalf("error = %v, want ErrSelfInvitation", err)
	}
}

func TestCreateInvitationDefaultsToAlgorithms(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository)

	if _, err := service.CreateInvitation(context.Background(), "sender", "receiver", ""); err != nil {
		t.Fatal(err)
	}
	if repository.createdProblemClass != problems.ClassAlgorithms {
		t.Fatalf("problem class = %q", repository.createdProblemClass)
	}
}

func TestCreateInvitationRejectsInvalidProblemClass(t *testing.T) {
	service := NewService(&repositoryStub{})
	_, err := service.CreateInvitation(context.Background(), "sender", "receiver", "unknown")
	if !errors.Is(err, ErrInvalidProblemClass) {
		t.Fatalf("error = %v, want ErrInvalidProblemClass", err)
	}
}

func TestUpdateCodeForwardsCursorPosition(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository)

	if err := service.UpdateCode(context.Background(), "user", "match", "package solution", 7, 12, 4); err != nil {
		t.Fatal(err)
	}
	if repository.updatedRevision != 7 || repository.updatedCursorLine != 12 || repository.updatedCursorCol != 4 {
		t.Fatalf(
			"revision/cursor = %d/%d:%d",
			repository.updatedRevision,
			repository.updatedCursorLine,
			repository.updatedCursorCol,
		)
	}
}

func TestUpdateCodeRejectsInvalidCursor(t *testing.T) {
	service := NewService(&repositoryStub{})
	err := service.UpdateCode(context.Background(), "user", "match", "package solution", 1, 0, 1)
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("error = %v, want ErrInvalidCursor", err)
	}
}
