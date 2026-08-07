package judge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
