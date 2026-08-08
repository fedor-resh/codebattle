package judge

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codebattle.local/codebattle/internal/problems"
	"codebattle.local/codebattle/internal/submissions"
)

func TestCompileAndRuntimeContainersAreIsolated(t *testing.T) {
	runner := NewRunner(Config{
		Image:          "judge-image",
		SourceVolume:   "source-volume",
		BinaryVolume:   "binary-volume",
		CacheVolume:    "cache-volume",
		CacheDirectory: "/cache",
	})
	job := submissions.Job{Submission: submissions.Submission{ID: "safe-job"}, MemoryLimitMB: 256}

	compile := strings.Join(runner.compileArgs(job), " ")
	for _, required := range []string{
		"--network none",
		"--read-only",
		"--pids-limit 64",
		"--cap-drop ALL",
		"no-new-privileges",
		"source-volume:/src:ro",
		"binary-volume:/out",
		"cache-volume:/cache",
		"GOCACHE=/cache/go-build",
		"--cpus 2",
	} {
		if !strings.Contains(compile, required) {
			t.Fatalf("compile args do not contain %q: %s", required, compile)
		}
	}

	runtime := strings.Join(runner.runtimeArgs(job), " ")
	if strings.Contains(runtime, "source-volume") {
		t.Fatalf("runtime container can access sources: %s", runtime)
	}
	if !strings.Contains(runtime, "binary-volume:/out:ro") {
		t.Fatalf("runtime binary volume is not read-only: %s", runtime)
	}
	if !strings.Contains(runtime, "-test.timeout=2s") {
		t.Fatalf("runtime binary does not enforce the solution timeout: %s", runtime)
	}
}

func TestMemoryFlagsCanBeDisabledForHostsWithoutCgroups(t *testing.T) {
	runner := NewRunner(Config{
		Image:              "judge-image",
		SourceVolume:       "source-volume",
		BinaryVolume:       "binary-volume",
		CacheVolume:        "cache-volume",
		DisableMemoryLimit: true,
	})
	job := submissions.Job{Submission: submissions.Submission{ID: "safe-job"}, MemoryLimitMB: 256}

	for name, args := range map[string][]string{
		"compile": runner.compileArgs(job),
		"runtime": runner.runtimeArgs(job),
	} {
		joined := strings.Join(args, " ")
		if strings.Contains(joined, "--memory") || strings.Contains(joined, "--memory-swap") {
			t.Fatalf("%s args contain unsupported memory flags: %s", name, joined)
		}
		for _, required := range []string{"--cpus", "--pids-limit 64", "--network none"} {
			if !strings.Contains(joined, required) {
				t.Fatalf("%s args lost %q: %s", name, required, joined)
			}
		}
	}
}

func TestCleanStaleArtifactsPreservesCache(t *testing.T) {
	root := t.TempDir()
	sourceDirectory := filepath.Join(root, "source")
	binaryDirectory := filepath.Join(root, "binary")
	cacheDirectory := filepath.Join(root, "cache")
	for _, directory := range []string{sourceDirectory, binaryDirectory, cacheDirectory} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "artifact"), []byte("data"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runner := NewRunner(Config{
		SourceDirectory: sourceDirectory,
		BinaryDirectory: binaryDirectory,
		CacheDirectory:  cacheDirectory,
	})

	if err := runner.CleanStaleArtifacts(); err != nil {
		t.Fatal(err)
	}
	for _, directory := range []string{sourceDirectory, binaryDirectory} {
		entries, err := os.ReadDir(directory)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Fatalf("directory %s was not cleaned", directory)
		}
	}
	if _, err := os.Stat(filepath.Join(cacheDirectory, "artifact")); err != nil {
		t.Fatalf("cache artifact was removed: %v", err)
	}
}

func TestLimitedBufferTruncatesOutput(t *testing.T) {
	buffer := &limitedBuffer{limit: 4}
	written, err := buffer.Write([]byte("123456"))
	if err != nil {
		t.Fatal(err)
	}
	if written != 6 || buffer.String() != "1234" {
		t.Fatalf("written=%d value=%q", written, buffer.String())
	}
}

func TestSanitizeOutputHidesInternalPathsAndHiddenTestName(t *testing.T) {
	runner := NewRunner(Config{
		SourceDirectory: "/judge-source",
		BinaryDirectory: "/judge-bin",
	})
	got := runner.sanitizeOutput(
		"/judge-source/job-1/hidden_test.go:12: secret compile detail",
		"job-1",
	)
	if strings.Contains(got, "/judge-source") || strings.Contains(got, "hidden_test.go") {
		t.Fatalf("sanitized output leaks internals: %q", got)
	}
}

func TestSanitizeOutputRemovesDockerWarningAndPackageBanner(t *testing.T) {
	runner := NewRunner(Config{})
	got := runner.sanitizeOutput(strings.Join([]string{
		"WARNING: Your kernel does not support memory limit capabilities or the cgroup is not mounted. Limitation discarded.",
		"# solution [solution.test]",
		"./solution.go:7:2: undefined: result",
	}, "\n"), "safe-job")

	if got != "./solution.go:7:2: undefined: result" {
		t.Fatalf("sanitized output = %q", got)
	}
}

func TestInfrastructureFailureMessage(t *testing.T) {
	message, ok := infrastructureFailureMessage("link: mapping output file failed: no space left on device")
	if !ok || !strings.Contains(message, "закончилось место") {
		t.Fatalf("message=%q ok=%v", message, ok)
	}
	if _, ok := infrastructureFailureMessage("./solution.go:2: undefined: value"); ok {
		t.Fatal("user compile error was classified as infrastructure failure")
	}
}

func TestPublicTestSourceIsValidGoAndQuotesValues(t *testing.T) {
	source := publicTestSource([]problems.PublicTest{
		{Input: "line one\n\"quoted\"", Expected: "ответ"},
	})
	if _, err := format.Source([]byte(source)); err != nil {
		t.Fatalf("generated public test is not valid Go: %v\n%s", err, source)
	}
	if !strings.Contains(source, publicResultMarker) {
		t.Fatal("generated public test does not contain the result marker")
	}
}

func TestCollectTestResultsIncludesPublicValuesAndHiddenSummary(t *testing.T) {
	publicTests := []problems.PublicTest{
		{Input: "level", Expected: "true"},
		{Input: "Code", Expected: "false"},
	}
	output := strings.Join([]string{
		`__CODEBATTLE_PUBLIC_RESULT__{"Index":1,"Actual":"true","Passed":true}`,
		`__CODEBATTLE_PUBLIC_RESULT__{"Index":2,"Actual":"true","Passed":false}`,
		"--- PASS: TestHidden (0.00s)",
	}, "\n")

	results := collectTestResults(output, publicTests, false)
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	if results[0].Status != "passed" || results[0].Actual != "true" {
		t.Fatalf("first public result = %+v", results[0])
	}
	if results[1].Status != "failed" || results[1].Actual != "true" {
		t.Fatalf("second public result = %+v", results[1])
	}
	if results[2].Kind != "hidden" || results[2].Status != "passed" {
		t.Fatalf("hidden result = %+v", results[2])
	}
	if got := failedTestMessage(results); got != "Публичный пример 2 не пройден" {
		t.Fatalf("message = %q", got)
	}
}

func TestCollectTestResultsUsesSuccessfulExitWhenOutputIsTruncated(t *testing.T) {
	results := collectTestResults("--- PASS: TestHidden", []problems.PublicTest{
		{Input: "hello", Expected: "world"},
	}, true)

	if results[0].Status != "passed" || results[0].ActualAvailable {
		t.Fatalf("public result = %+v", results[0])
	}
	if results[1].Status != "passed" {
		t.Fatalf("hidden result = %+v", results[1])
	}
}
