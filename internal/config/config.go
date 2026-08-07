package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
)

type API struct {
	Environment     string
	HTTPAddress     string
	DatabaseURL     string
	PostgresAddress string
	PublicOrigins   []string
}

type Worker struct {
	Environment     string
	HealthAddress   string
	DatabaseURL     string
	DockerBinary    string
	JudgeImage      string
	SourceDirectory string
	BinaryDirectory string
	CacheDirectory  string
	SourceVolume    string
	BinaryVolume    string
	CacheVolume     string
}

func LoadAPI() (API, error) {
	databaseURL := env("DATABASE_URL", "postgres://codebattle:codebattle@localhost:5432/codebattle?sslmode=disable")
	postgresAddress, err := addressFromURL(databaseURL, "5432")
	if err != nil {
		return API{}, fmt.Errorf("DATABASE_URL: %w", err)
	}

	return API{
		Environment:     env("APP_ENV", "development"),
		HTTPAddress:     env("HTTP_ADDR", ":8080"),
		DatabaseURL:     databaseURL,
		PostgresAddress: postgresAddress,
		PublicOrigins: splitCSV(env(
			"PUBLIC_ORIGINS",
			"http://127.0.0.1:8088,http://localhost:8088,http://127.0.0.1:5173,http://localhost:5173",
		)),
	}, nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func LoadWorker() (Worker, error) {
	return Worker{
		Environment:     env("APP_ENV", "development"),
		HealthAddress:   env("WORKER_HEALTH_ADDR", ":8081"),
		DatabaseURL:     env("DATABASE_URL", "postgres://codebattle:codebattle@localhost:5432/codebattle?sslmode=disable"),
		DockerBinary:    env("DOCKER_BINARY", "docker"),
		JudgeImage:      env("JUDGE_IMAGE", "golang:1.26.5-alpine"),
		SourceDirectory: env("JUDGE_SOURCE_DIR", "/judge-source"),
		BinaryDirectory: env("JUDGE_BINARY_DIR", "/judge-bin"),
		CacheDirectory:  env("JUDGE_CACHE_DIR", "/judge-cache"),
		SourceVolume:    env("JUDGE_SOURCE_VOLUME", "codebattle-judge-source"),
		BinaryVolume:    env("JUDGE_BINARY_VOLUME", "codebattle-judge-bin"),
		CacheVolume:     env("JUDGE_CACHE_VOLUME", "codebattle-judge-cache"),
	}, nil
}

func addressFromURL(rawURL, defaultPort string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsed.Hostname() == "" {
		return "", fmt.Errorf("missing hostname")
	}
	port := parsed.Port()
	if port == "" {
		port = defaultPort
	}
	return net.JoinHostPort(parsed.Hostname(), port), nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
