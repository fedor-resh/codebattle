package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"codebattle.local/codebattle/internal/config"
	"codebattle.local/codebattle/internal/judge"
	"codebattle.local/codebattle/internal/submissions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.LoadWorker()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		logger.Error("ping postgres", "error", err)
		os.Exit(1)
	}

	repository := submissions.NewRepository(pool)
	runner := judge.NewRunner(judge.Config{
		DockerBinary:    cfg.DockerBinary,
		Image:           cfg.JudgeImage,
		SourceDirectory: cfg.SourceDirectory,
		BinaryDirectory: cfg.BinaryDirectory,
		SourceVolume:    cfg.SourceVolume,
		BinaryVolume:    cfg.BinaryVolume,
	})

	server := healthServer(cfg.HealthAddress, pool, cfg.DockerBinary)
	go func() {
		logger.Info("judge worker listening", "health_address", cfg.HealthAddress)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("worker health server failed", "error", err)
			stop()
		}
	}()
	go work(ctx, logger, repository, runner)

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("worker shutdown failed", "error", err)
	}
}

func work(ctx context.Context, logger *slog.Logger, repository *submissions.Repository, runner *judge.Runner) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		job, err := repository.Claim(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				continue
			}
		}
		if err != nil {
			logger.Error("claim submission", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				continue
			}
		}

		logger.Info("judge submission", "submission_id", job.ID, "match_id", job.MatchID)
		result := runner.Run(ctx, job, func() error {
			return repository.SetStatus(ctx, job.ID, "running", map[string]string{"message": "Тесты выполняются"})
		})
		if err := repository.Finish(ctx, job, result.Status, result); err != nil {
			logger.Error("finish submission", "submission_id", job.ID, "error", err)
			continue
		}
		logger.Info("submission finished", "submission_id", job.ID, "status", result.Status, "duration_ms", result.DurationMS)
	}
}

func healthServer(address string, pool *pgxpool.Pool, dockerBinary string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"status":"ok","service":"judge-worker"}`))
	})
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, request *http.Request) {
		ctx, cancel := context.WithTimeout(request.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, `{"status":"error","dependency":"postgres"}`, http.StatusServiceUnavailable)
			return
		}
		if err := exec.CommandContext(ctx, dockerBinary, "version", "--format", "{{.Server.Version}}").Run(); err != nil {
			http.Error(w, `{"status":"error","dependency":"docker"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"status":"ok","service":"judge-worker"}`))
	})
	return &http.Server{Addr: address, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
}
