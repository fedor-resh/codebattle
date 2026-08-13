package problems

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestCatalogContainsThirtyValidProblems(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	catalog, err := LoadCatalog(ctx, filepath.Join("..", "..", "problems"), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog) != 30 {
		t.Fatalf("catalog size = %d, want 30", len(catalog))
	}
	seen := map[string]bool{}
	classCounts := map[Class]int{}
	for _, problem := range catalog {
		if seen[problem.Slug] {
			t.Fatalf("duplicate slug %q", problem.Slug)
		}
		seen[problem.Slug] = true
		classCounts[problem.Class]++
		if len(problem.ContentHash) != 64 {
			t.Fatalf("problem %s hash length = %d", problem.Slug, len(problem.ContentHash))
		}
	}
	if classCounts[ClassAlgorithms] != 25 || classCounts[ClassConcurrency] != 5 {
		t.Fatalf("class counts = %v", classCounts)
	}
	for _, problem := range catalog {
		if problem.Slug == "fizz-buzz" && problem.ContentHash != "5f9bf75c09db37f8c4c5f705e9dc45c7f1e685ca6fdd40670b79dcda3797d227" {
			t.Fatalf("legacy fizz-buzz hash changed: %s", problem.ContentHash)
		}
	}
}

func TestValidateRequirementsAcceptsAliasedImports(t *testing.T) {
	source := `package solution

import (
	c "context"
	s "sync"
)

func Solve(input int) int {
	ctx, cancel := c.WithCancel(c.Background())
	defer cancel()
	values := make(chan int)
	var wait s.WaitGroup
	wait.Add(1)
	go func() { defer wait.Done(); values <- input }()
	var result int
	select {
	case result = <-values:
	case <-ctx.Done():
	}
	wait.Wait()
	return result
}`
	requirements := Requirements{
		Goroutine: true, Channel: true, WaitGroup: true,
		Mutex: false, Select: true, ContextCancel: true,
	}
	if err := ValidateRequirements(source, requirements); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRequirementsListsMissingConstructs(t *testing.T) {
	err := ValidateRequirements(
		"package solution\nfunc Solve(input int) int { return input }",
		Requirements{Goroutine: true, Channel: true, Mutex: true},
	)
	if err == nil {
		t.Fatal("sequential source passed concurrency requirements")
	}
}

func TestValidateMetadataRejectsInvalidClassAndRequirements(t *testing.T) {
	base := Metadata{
		Slug: "sample", Title: "Sample", Difficulty: "easy", Class: ClassAlgorithms,
		Function: "Solve", Signature: "func Solve(input int) int", Version: 1,
		TimeLimitMS: 1000, MemoryLimitMB: 64,
	}

	invalidClass := base
	invalidClass.Class = "unknown"
	if err := validateMetadata(invalidClass, "sample"); err == nil {
		t.Fatal("unknown class was accepted")
	}

	algorithmWithRequirements := base
	algorithmWithRequirements.Requirements.Goroutine = true
	if err := validateMetadata(algorithmWithRequirements, "sample"); err == nil {
		t.Fatal("algorithm task with concurrency requirements was accepted")
	}

	concurrencyWithoutGoroutine := base
	concurrencyWithoutGoroutine.Class = ClassConcurrency
	concurrencyWithoutGoroutine.Requirements.Channel = true
	if err := validateMetadata(concurrencyWithoutGoroutine, "sample"); err == nil {
		t.Fatal("concurrency task without a goroutine requirement was accepted")
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
