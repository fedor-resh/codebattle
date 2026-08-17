package problems

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestCatalogContainsFortyTwoValidProblems(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	catalog, err := LoadCatalog(ctx, filepath.Join("..", "..", "problems"), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog) != 42 {
		t.Fatalf("catalog size = %d, want 42", len(catalog))
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
	if classCounts[ClassAlgorithms] != 25 || classCounts[ClassConcurrency] != 7 || classCounts[ClassOOP] != 10 {
		t.Fatalf("class counts = %v", classCounts)
	}
	for _, problem := range catalog {
		if problem.Slug == "fizz-buzz" && problem.ContentHash != "5f9bf75c09db37f8c4c5f705e9dc45c7f1e685ca6fdd40670b79dcda3797d227" {
			t.Fatalf("legacy fizz-buzz hash changed: %s", problem.ContentHash)
		}
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

	oopTask := base
	oopTask.Class = ClassOOP
	if err := validateMetadata(oopTask, "sample"); err != nil {
		t.Fatalf("oop task should be accepted: %v", err)
	}

	concurrencyWithoutRequirements := base
	concurrencyWithoutRequirements.Class = ClassConcurrency
	if err := validateMetadata(concurrencyWithoutRequirements, "sample"); err != nil {
		t.Fatalf("concurrency task should be validated by behavior, not source requirements: %v", err)
	}
}

func TestParseSignatureSupportsLongRunningCallbacks(t *testing.T) {
	for _, signature := range []string{
		"func Solve(values []int, workers int, work func(int) int) []int",
		"func Solve(values []string, workers int, work func(string) string) map[string]int",
		"func Solve(delays []int, timeoutMS int, work func(context.Context, int) bool) int",
	} {
		if _, err := ParseSignature(signature); err != nil {
			t.Fatalf("ParseSignature(%q): %v", signature, err)
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
