package practice

import (
	"context"
	"errors"
	"testing"
	"time"
)

type repositoryStub struct{}

func (repositoryStub) Problems(context.Context, string) ([]ProblemSummary, error) {
	return nil, nil
}

func (repositoryStub) StartSession(context.Context, string, string, string) (string, error) {
	return "", nil
}

func (repositoryStub) Session(context.Context, string, string) (Session, error) {
	return Session{}, nil
}

func (repositoryStub) UpdateCode(context.Context, string, string, string, int64, time.Time) error {
	return nil
}

func TestUpdateCodeRejectsOversizedSource(t *testing.T) {
	service := NewService(repositoryStub{})
	err := service.UpdateCode(context.Background(), "user", "session", string(make([]byte, 64*1024+1)), 1)
	if !errors.Is(err, ErrSourceTooLarge) {
		t.Fatalf("error = %v, want ErrSourceTooLarge", err)
	}
}

func TestStartSessionRejectsEmptySlug(t *testing.T) {
	service := NewService(repositoryStub{})
	_, err := service.StartSession(context.Background(), "user", "")
	if !errors.Is(err, ErrProblemNotFound) {
		t.Fatalf("error = %v, want ErrProblemNotFound", err)
	}
}
