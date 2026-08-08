package duels

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrSelfInvitation  = errors.New("cannot invite yourself")
	ErrUserNotFound    = errors.New("user not found")
	ErrUserUnavailable = errors.New("user is unavailable")
	ErrInvitationBusy  = errors.New("user already has an invitation")
	ErrInvitationGone  = errors.New("invitation is no longer pending")
	ErrForbidden       = errors.New("forbidden")
	ErrMatchNotFound   = errors.New("match not found")
	ErrProblemsMissing = errors.New("problem catalog is empty")
	ErrRoundNotActive  = errors.New("round is not active")
	ErrStaleRevision   = errors.New("stale editor revision")
	ErrSourceTooLarge  = errors.New("source is too large")
	ErrInvalidCursor   = errors.New("invalid editor cursor")
)

const invitationTTL = 30 * time.Second

type Player struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Invitation struct {
	ID        string    `json:"id"`
	Sender    Player    `json:"sender"`
	Receiver  Player    `json:"receiver"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Match struct {
	ID             string         `json:"id"`
	PlayerOne      Player         `json:"player_one"`
	PlayerTwo      Player         `json:"player_two"`
	PlayerOneScore int            `json:"player_one_score"`
	PlayerTwoScore int            `json:"player_two_score"`
	State          string         `json:"state"`
	Problem        *Problem       `json:"problem,omitempty"`
	RoundNumber    int            `json:"round_number"`
	RoundWinnerID  string         `json:"round_winner_id,omitempty"`
	PlayerOneReady bool           `json:"player_one_ready"`
	PlayerTwoReady bool           `json:"player_two_ready"`
	WinningSource  string         `json:"winning_source_code,omitempty"`
	CodeSnapshots  []CodeSnapshot `json:"code_snapshots"`
	PausedAt       *time.Time     `json:"paused_at,omitempty"`
}

type CodeSnapshot struct {
	UserID           string `json:"user_id"`
	ProblemVersionID string `json:"problem_version_id"`
	SourceCode       string `json:"source_code"`
	Revision         int64  `json:"revision"`
	CursorLine       int    `json:"cursor_line"`
	CursorColumn     int    `json:"cursor_column"`
}

type Problem struct {
	ID                string          `json:"id"`
	Slug              string          `json:"slug"`
	Title             string          `json:"title"`
	Difficulty        string          `json:"difficulty"`
	Statement         string          `json:"statement_markdown"`
	FunctionSignature string          `json:"function_signature"`
	StarterCode       string          `json:"starter_code"`
	PublicTests       json.RawMessage `json:"public_tests"`
	TimeLimitMS       int             `json:"time_limit_ms"`
	MemoryLimitMB     int             `json:"memory_limit_mb"`
}

type State struct {
	Incoming *Invitation `json:"incoming,omitempty"`
	Outgoing *Invitation `json:"outgoing,omitempty"`
	Match    *Match      `json:"match,omitempty"`
}

type Repository interface {
	CreateInvitation(context.Context, string, string, string, time.Time, time.Time) (Invitation, error)
	State(context.Context, string, time.Time) (State, error)
	AcceptInvitation(context.Context, string, string, string, time.Time) (Match, error)
	DeclineInvitation(context.Context, string, string, time.Time) error
	Match(context.Context, string, string) (Match, error)
	LeaveMatch(context.Context, string, string, time.Time) error
	UpdateCode(context.Context, string, string, string, int64, int, int, time.Time) error
	Ready(context.Context, string, string, time.Time) (Match, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (s *Service) CreateInvitation(ctx context.Context, senderID, receiverID string) (Invitation, error) {
	if senderID == receiverID {
		return Invitation{}, ErrSelfInvitation
	}
	now := s.now().UTC()
	return s.repository.CreateInvitation(ctx, randomID(), senderID, receiverID, now, now.Add(invitationTTL))
}

func (s *Service) State(ctx context.Context, userID string) (State, error) {
	return s.repository.State(ctx, userID, s.now().UTC())
}

func (s *Service) AcceptInvitation(ctx context.Context, userID, invitationID string) (Match, error) {
	return s.repository.AcceptInvitation(ctx, randomID(), invitationID, userID, s.now().UTC())
}

func (s *Service) DeclineInvitation(ctx context.Context, userID, invitationID string) error {
	return s.repository.DeclineInvitation(ctx, invitationID, userID, s.now().UTC())
}

func (s *Service) Match(ctx context.Context, userID, matchID string) (Match, error) {
	return s.repository.Match(ctx, matchID, userID)
}

func (s *Service) LeaveMatch(ctx context.Context, userID, matchID string) error {
	return s.repository.LeaveMatch(ctx, matchID, userID, s.now().UTC())
}

func (s *Service) UpdateCode(
	ctx context.Context,
	userID, matchID, source string,
	revision int64,
	cursorLine, cursorColumn int,
) error {
	if len([]byte(source)) > 64*1024 {
		return ErrSourceTooLarge
	}
	if cursorLine < 1 || cursorColumn < 1 || cursorLine > 1_000_000 || cursorColumn > 1_000_000 {
		return ErrInvalidCursor
	}
	return s.repository.UpdateCode(
		ctx, matchID, userID, source, revision, cursorLine, cursorColumn, s.now().UTC(),
	)
}

func (s *Service) Ready(ctx context.Context, userID, matchID string) (Match, error) {
	return s.repository.Ready(ctx, matchID, userID, s.now().UTC())
}

func randomID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer)
}
