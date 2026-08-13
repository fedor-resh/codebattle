package httpapi

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"codebattle.local/codebattle/internal/duels"
	"codebattle.local/codebattle/internal/health"
)

type healthyDependency struct{ name string }

func (d healthyDependency) Name() string                { return d.name }
func (d healthyDependency) Check(context.Context) error { return nil }

func TestInvalidProblemClassErrorContract(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPost, "/api/v1/invitations", nil)
	response := httptest.NewRecorder()
	duelHandlers{}.writeDuelError(response, request, duels.ErrInvalidProblemClass)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnprocessableEntity)
	}
	if body := response.Body.String(); !bytes.Contains([]byte(body), []byte(`"code":"INVALID_PROBLEM_CLASS"`)) {
		t.Fatalf("response body = %s", body)
	}
}

func TestLiveHealth(t *testing.T) {
	t.Parallel()

	handler := NewHandler(
		slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		health.NewChecker(healthyDependency{name: "postgres"}),
		"test",
	)

	request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", got)
	}
}

func TestReadyHealth(t *testing.T) {
	t.Parallel()

	handler := NewHandler(
		slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)),
		health.NewChecker(healthyDependency{name: "postgres"}),
		"test",
	)

	request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestOriginGuardRejectsUnknownOriginForMutation(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := requestID(originGuard([]string{"http://localhost:8088"}, next))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	request.Header.Set("Origin", "https://attacker.example")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
	if response.Header().Get("X-Request-ID") == "" {
		t.Fatal("response has no request id")
	}
}

func TestOriginGuardAllowsConfiguredOrigin(t *testing.T) {
	t.Parallel()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := originGuard([]string{"http://localhost:8088"}, next)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	request.Header.Set("Origin", "http://localhost:8088")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestRateLimiterUsesFixedWindow(t *testing.T) {
	t.Parallel()

	limiter := newRateLimiter()
	now := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }
	handler := limiter.wrap(2, time.Minute, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for attempt, want := range []int{http.StatusNoContent, http.StatusNoContent, http.StatusTooManyRequests} {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		request.RemoteAddr = "192.0.2.1:1234"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("attempt %d: status = %d, want %d", attempt+1, response.Code, want)
		}
	}

	now = now.Add(time.Minute)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	request.RemoteAddr = "192.0.2.1:1234"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status after reset = %d, want %d", response.Code, http.StatusNoContent)
	}
}
