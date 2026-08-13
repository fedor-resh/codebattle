package problems

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestConcurrencyHiddenTestsRejectSequentialWork(t *testing.T) {
	tests := []struct {
		slug   string
		marker string
		source string
	}{
		{
			slug:   "parallel-sum",
			marker: "work was not performed in parallel",
			source: `package solution

func Solve(nums []int, workers int, work func(int) int) int {
	result := 0
	for _, value := range nums {
		result += work(value)
	}
	return result
}
`,
		},
		{
			slug:   "ordered-worker-pool",
			marker: "work was not performed in parallel",
			source: `package solution

func Solve(nums []int, workers int, work func(int) int) []int {
	result := make([]int, len(nums))
	for index, value := range nums {
		result[index] = work(value)
	}
	return result
}
`,
		},
		{
			slug:   "concurrent-word-frequency",
			marker: "work was not performed in parallel",
			source: `package solution

func Solve(words []string, workers int, work func(string) string) map[string]int {
	result := make(map[string]int)
	for _, word := range words {
		result[work(word)]++
	}
	return result
}
`,
		},
		{
			slug:   "goroutine-relay",
			marker: "ping and pong were not processed in parallel",
			source: `package solution

import "strings"

func Solve(n int, work func(string) string) string {
	var result strings.Builder
	for range max(n, 0) {
		result.WriteString(work("ping"))
		result.WriteString(work("pong"))
	}
	return result.String()
}
`,
		},
		{
			slug:   "fastest-handler",
			marker: "handlers were not started in parallel",
			source: `package solution

import "context"

func Solve(delays []int, timeoutMS int, work func(context.Context, int) bool) int {
	if len(delays) == 0 || timeoutMS <= 0 {
		return -1
	}
	fastest := 0
	for index := 1; index < len(delays); index++ {
		if delays[index] < delays[fastest] {
			fastest = index
		}
	}
	if delays[fastest] >= timeoutMS {
		return -1
	}
	for _, delay := range delays {
		work(context.Background(), delay)
	}
	return fastest
}
`,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.slug, func(t *testing.T) {
			directory := t.TempDir()
			hidden, err := os.ReadFile(filepath.Join("..", "..", "problems", testCase.slug, "hidden_test.go"))
			if err != nil {
				t.Fatal(err)
			}
			for name, content := range map[string][]byte{
				"go.mod":         []byte("module solution\n\ngo 1.26.0\n"),
				"solution.go":    []byte(testCase.source),
				"hidden_test.go": hidden,
			} {
				if err := os.WriteFile(filepath.Join(directory, name), content, 0o600); err != nil {
					t.Fatal(err)
				}
			}

			command := exec.Command("go", "test", "-run", "^TestHidden$", ".")
			command.Dir = directory
			command.Env = append(os.Environ(), "GOWORK=off")
			output, err := command.CombinedOutput()
			if err == nil {
				t.Fatal("sequential solution passed the hidden timing test")
			}
			if !strings.Contains(string(output), testCase.marker) {
				t.Fatalf("sequential solution failed for an unexpected reason:\n%s", output)
			}
		})
	}
}
