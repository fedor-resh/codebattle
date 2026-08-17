package practice

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"codebattle.local/codebattle/internal/problems"
)

var (
	ErrProblemNotFound = errors.New("problem not found")
	ErrSessionNotFound = errors.New("practice session not found")
	ErrStaleRevision   = errors.New("stale editor revision")
	ErrSourceTooLarge  = errors.New("source is too large")
)

type ProblemSummary struct {
	Slug         string                `json:"slug"`
	Title        string                `json:"title"`
	Difficulty   string                `json:"difficulty"`
	ProblemClass problems.Class        `json:"problem_class"`
	Requirements problems.Requirements `json:"requirements"`
	Solved       bool                  `json:"solved"`
}

type Problem struct {
	ID                string                `json:"id"`
	Slug              string                `json:"slug"`
	Title             string                `json:"title"`
	Difficulty        string                `json:"difficulty"`
	ProblemClass      problems.Class        `json:"problem_class"`
	Requirements      problems.Requirements `json:"requirements"`
	Statement         string                `json:"statement_markdown"`
	FunctionSignature string                `json:"function_signature"`
	StarterCode       string                `json:"starter_code"`
	PublicTests       json.RawMessage       `json:"public_tests"`
	TimeLimitMS       int                   `json:"time_limit_ms"`
	MemoryLimitMB     int                   `json:"memory_limit_mb"`
}

type Session struct {
	ID         string     `json:"id"`
	Problem    Problem    `json:"problem"`
	SourceCode string     `json:"source_code"`
	Revision   int64      `json:"revision"`
	SolvedAt   *time.Time `json:"solved_at,omitempty"`
}

type Repository interface {
	Problems(context.Context, string) ([]ProblemSummary, error)
	StartSession(context.Context, string, string, string) (string, error)
	Session(context.Context, string, string) (Session, error)
	UpdateCode(context.Context, string, string, string, int64, time.Time) error
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Problems(ctx context.Context, userID string) ([]ProblemSummary, error) {
	return s.repository.Problems(ctx, userID)
}

func (s *Service) StartSession(ctx context.Context, userID, slug string) (Session, error) {
	if slug == "" {
		return Session{}, ErrProblemNotFound
	}
	sessionID, err := s.repository.StartSession(ctx, randomID(), userID, slug)
	if err != nil {
		return Session{}, err
	}
	return s.repository.Session(ctx, userID, sessionID)
}

func (s *Service) Session(ctx context.Context, userID, sessionID string) (Session, error) {
	return s.repository.Session(ctx, userID, sessionID)
}

func (s *Service) UpdateCode(ctx context.Context, userID, sessionID, source string, revision int64) error {
	if len([]byte(source)) > 64*1024 {
		return ErrSourceTooLarge
	}
	return s.repository.UpdateCode(ctx, sessionID, userID, source, revision, time.Now().UTC())
}

func randomID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer)
}
