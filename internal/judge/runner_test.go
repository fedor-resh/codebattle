package judge

import (
	"strings"
	"testing"

	"codebattle.local/codebattle/internal/submissions"
)

func TestCompileAndRuntimeContainersAreIsolated(t *testing.T) {
	runner := NewRunner(Config{
		Image:        "judge-image",
		SourceVolume: "source-volume",
		BinaryVolume: "binary-volume",
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
