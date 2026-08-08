package problems

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestCatalogContainsTwentyFiveValidProblems(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	catalog, err := LoadCatalog(ctx, filepath.Join("..", "..", "problems"), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog) != 25 {
		t.Fatalf("catalog size = %d, want 25", len(catalog))
	}
	seen := map[string]bool{}
	for _, problem := range catalog {
		if seen[problem.Slug] {
			t.Fatalf("duplicate slug %q", problem.Slug)
		}
		seen[problem.Slug] = true
		if len(problem.ContentHash) != 64 {
			t.Fatalf("problem %s hash length = %d", problem.Slug, len(problem.ContentHash))
		}
	}
}

func TestValidateSolutionRejectsWrongSignature(t *testing.T) {
	err := ValidateSolution(
		"package solution\nfunc Solve(input int) string { return \"\" }",
		"func Solve(input string) string",
	)
	if err == nil {
		t.Fatal("wrong signature was accepted")
	}
}

func TestValidateSolutionAcceptsPreparedArguments(t *testing.T) {
	err := ValidateSolution(
		"package solution\nfunc Solve(values []int, target int) []int { return nil }",
		"func Solve(nums []int, target int) []int",
	)
	if err != nil {
		t.Fatal(err)
	}
}
