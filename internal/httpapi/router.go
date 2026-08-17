package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"codebattle.local/codebattle/internal/accounts"
	"codebattle.local/codebattle/internal/duels"
	"codebattle.local/codebattle/internal/health"
	"codebattle.local/codebattle/internal/submissions"
)

const appVersion = "0.3.0"

type Dependencies struct {
	Accounts       *accounts.Service
	Duels          *duels.Service
	Submissions    *submissions.Repository
	SecureCookies  bool
	AllowedOrigins []string
}

func NewHandler(logger *slog.Logger, checker *health.Checker, environment string, dependencies ...Dependencies) http.Handler {
	mux := http.NewServeMux()
	limiter := newRateLimiter()
	var deps Dependencies
	if len(dependencies) > 0 {
		deps = dependencies[0]
	}

	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "ok",
			"service": "api",
			"version": appVersion,
		})
	})

	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		status := checker.Readiness(r.Context())
		code := http.StatusOK
		if !status.OK {
			code = http.StatusServiceUnavailable
		}
		writeJSON(w, code, status)
	})

	mux.HandleFunc("GET /api/v1/meta", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"name":        "CodeBattle",
			"version":     appVersion,
			"environment": environment,
			"stage":       "mvp",
		})
	})

	if deps.Accounts != nil {
		handlers := accountHandlers{service: deps.Accounts, secureCookies: deps.SecureCookies}
		mux.Handle("POST /api/v1/auth/register", limiter.wrap(5, time.Minute, http.HandlerFunc(handlers.register)))
		mux.Handle("POST /api/v1/auth/login", limiter.wrap(10, time.Minute, http.HandlerFunc(handlers.login)))
		mux.HandleFunc("POST /api/v1/auth/logout", handlers.logout)
		mux.HandleFunc("GET /api/v1/me", handlers.me)
		mux.HandleFunc("POST /api/v1/presence/heartbeat", handlers.heartbeat)
		mux.HandleFunc("GET /api/v1/users", handlers.users)

		if deps.Duels != nil {
			duelRoutes := duelHandlers{accounts: handlers, service: deps.Duels}
			mux.Handle("POST /api/v1/invitations", limiter.wrap(20, time.Minute, http.HandlerFunc(duelRoutes.createInvitation)))
			mux.HandleFunc("GET /api/v1/invitations", duelRoutes.invitationState)
			mux.HandleFunc("POST /api/v1/invitations/{id}/accept", duelRoutes.acceptInvitation)
			mux.HandleFunc("POST /api/v1/invitations/{id}/decline", duelRoutes.declineInvitation)
			mux.HandleFunc("GET /api/v1/matches/{id}", duelRoutes.match)
			mux.HandleFunc("POST /api/v1/matches/{id}/leave", duelRoutes.leaveMatch)
			mux.HandleFunc("PUT /api/v1/matches/{id}/code", duelRoutes.updateCode)
			mux.HandleFunc("POST /api/v1/matches/{id}/ready", duelRoutes.ready)
			mux.HandleFunc("POST /api/v1/matches/{id}/skip", duelRoutes.skip)
		}
		if deps.Submissions != nil {
			submissionRoutes := submissionHandlers{accounts: handlers, repository: deps.Submissions}
			mux.HandleFunc("POST /api/v1/matches/{id}/submissions", submissionRoutes.create)
			mux.HandleFunc("GET /api/v1/submissions/{id}", submissionRoutes.get)
		}
	}

	return requestID(requestLogger(logger, recoverer(logger, originGuard(deps.AllowedOrigins, securityHeaders(mux)))))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type requestIDKey struct{}

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buffer := make([]byte, 8)
		_, _ = rand.Read(buffer)
		id := hex.EncodeToString(buffer)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string, details any) {
	payload := map[string]any{
		"code":       code,
		"message":    message,
		"request_id": r.Context().Value(requestIDKey{}),
	}
	if details != nil {
		payload["details"] = details
	}
	writeJSON(w, status, payload)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

func originGuard(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutating := r.Method == http.MethodPost || r.Method == http.MethodPut ||
			r.Method == http.MethodPatch || r.Method == http.MethodDelete
		origin := r.Header.Get("Origin")
		if mutating && origin != "" && !allowed[origin] {
			writeError(w, r, http.StatusForbidden, "ORIGIN_NOT_ALLOWED", "Недопустимый источник запроса", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("http request",
			"request_id", r.Context().Value(requestIDKey{}),
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(started).Milliseconds(),
		)
	})
}

func recoverer(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("http panic", "error", recovered, "stack", strings.TrimSpace(string(debug.Stack())))
				writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
