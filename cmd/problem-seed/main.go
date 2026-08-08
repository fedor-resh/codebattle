package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"codebattle.local/codebattle/db/migrations"
	"codebattle.local/codebattle/internal/config"
	"codebattle.local/codebattle/internal/problems"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	directory := flag.String("dir", "problems", "path to problem catalog")
	validateOnly := flag.Bool("validate-only", false, "validate files without writing to PostgreSQL")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	catalog, err := problems.LoadCatalog(ctx, *directory, true)
	if err != nil {
		logger.Error("validate problem catalog", "error", err)
		os.Exit(1)
	}
	if len(catalog) != 25 {
		logger.Error("problem catalog must contain exactly 25 tasks", "count", len(catalog))
		os.Exit(1)
	}
	if *validateOnly {
		fmt.Printf("validated %d problems\n", len(catalog))
		return
	}

	cfg, err := config.LoadAPI()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := migrations.Apply(ctx, pool); err != nil {
		logger.Error("apply migrations", "error", err)
		os.Exit(1)
	}
	result, err := problems.Seed(ctx, pool, catalog)
	if err != nil {
		logger.Error("seed problems", "error", err)
		os.Exit(1)
	}
	logger.Info("problem catalog ready", "inserted", result.Inserted, "skipped", result.Skipped)
}
