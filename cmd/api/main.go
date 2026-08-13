package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codebattle.local/codebattle/db/migrations"
	"codebattle.local/codebattle/internal/accounts"
	"codebattle.local/codebattle/internal/config"
	"codebattle.local/codebattle/internal/duels"
	"codebattle.local/codebattle/internal/health"
	"codebattle.local/codebattle/internal/httpapi"
	"codebattle.local/codebattle/internal/submissions"
	"github.com/jackc/pgx/v5/pgxpool"
)

const shutdownTimeout = 10 * time.Second

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.LoadAPI()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer startupCancel()
	pool, err := pgxpool.New(startupCtx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(startupCtx); err != nil {
		logger.Error("ping postgres", "error", err)
		os.Exit(1)
	}
	if err := migrations.Apply(startupCtx, pool); err != nil {
		logger.Error("apply migrations", "error", err)
		os.Exit(1)
	}

	accountService := accounts.NewService(accounts.NewPostgresStore(pool))
	duelRepository := duels.NewPostgresRepository(pool)
	duelService := duels.NewService(duelRepository)
	submissionRepository := submissions.NewRepository(pool)

	checker := health.NewChecker(
		health.NewTCPDependency("postgres", cfg.PostgresAddress),
	)

	server := &http.Server{
		Addr: cfg.HTTPAddress,
		Handler: httpapi.NewHandler(logger, checker, cfg.Environment, httpapi.Dependencies{
			Accounts:       accountService,
			Duels:          duelService,
			Submissions:    submissionRepository,
			SecureCookies:  cfg.Environment == "production",
			AllowedOrigins: cfg.PublicOrigins,
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go cleanExpiredRecords(ctx, logger, pool)

	go func() {
		logger.Info("api listening", "address", cfg.HTTPAddress, "environment", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("api shutdown started")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("api shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("api shutdown completed")
}

func cleanExpiredRecords(ctx context.Context, logger *slog.Logger, pool *pgxpool.Pool) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := pool.Exec(ctx, "DELETE FROM sessions WHERE expires_at <= now()"); err != nil {
				logger.Error("clean expired sessions", "error", err)
			}
			if _, err := pool.Exec(ctx, `
				UPDATE invitations SET status = 'expired', responded_at = now()
				WHERE status = 'pending' AND expires_at <= now()
			`); err != nil {
				logger.Error("expire invitations", "error", err)
			}
		}
	}
}
