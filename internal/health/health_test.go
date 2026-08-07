package health

import (
	"context"
	"errors"
	"testing"
)

type fakeDependency struct {
	name string
	err  error
}

func (d fakeDependency) Name() string                { return d.name }
func (d fakeDependency) Check(context.Context) error { return d.err }

func TestCheckerReadiness(t *testing.T) {
	t.Parallel()

	checker := NewChecker(
		fakeDependency{name: "postgres"},
		fakeDependency{name: "redis", err: errors.New("unavailable")},
	)

	got := checker.Readiness(context.Background())
	if got.OK {
		t.Fatal("Readiness().OK = true, want false")
	}
	if !got.Dependencies["postgres"] {
		t.Fatal("postgres should be ready")
	}
	if got.Dependencies["redis"] {
		t.Fatal("redis should not be ready")
	}
}
