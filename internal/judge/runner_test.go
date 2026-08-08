package judge

import (
	"encoding/json"
	"go/format"
	"os"
	"os/exec"
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

	runtime := strings.Join(runner.runtimeArgs(job, "^TestHidden$"), " ")
	if strings.Contains(runtime, "source-volume") {
		t.Fatalf("runtime container can access sources: %s", runtime)
	}
	if !strings.Contains(runtime, "binary-volume:/out:ro") {
		t.Fatalf("runtime binary volume is not read-only: %s", runtime)
	}
	if !strings.Contains(runtime, "-test.timeout=2s") {
		t.Fatalf("runtime binary does not enforce the solution timeout: %s", runtime)
	}
	if !strings.Contains(runtime, "-test.run=^TestHidden$") {
		t.Fatalf("runtime binary does not select the requested test phase: %s", runtime)
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
		"runtime": runner.runtimeArgs(job, "^TestHidden$"),
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
	source, err := publicTestSource([]problems.PublicTest{
		testPublicCase([]string{`"line one\n\"quoted\""`}, `"ответ"`),
	}, "func Solve(text string) string")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := format.Source([]byte(source)); err != nil {
		t.Fatalf("generated public test is not valid Go: %v\n%s", err, source)
	}
	if !strings.Contains(source, publicResultMarker) {
		t.Fatal("generated public test does not contain the result marker")
	}
}

func TestGeneratedPublicTestCapturesActualConsoleAndPanic(t *testing.T) {
	directory := t.TempDir()
	publicTests := []problems.PublicTest{
		testPublicCase([]string{`"hello"`}, `"HELLO"`),
		testPublicCase([]string{`"panic"`}, `"PANIC"`),
	}
	publicSource, err := publicTestSource(publicTests, "func Solve(input string) string")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"go.mod": "module solution\n\ngo 1.26.0\n",
		"solution.go": `package solution

import (
	"fmt"
	"strings"
)

func Solve(input string) string {
	fmt.Println("processing", input)
	if input == "panic" {
		panic("boom")
	}
	return strings.ToUpper(input)
}
`,
		"public_test.go": publicSource,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	command := exec.Command("go", "test", "-v", ".")
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("generated tests unexpectedly passed despite the panic case")
	}
	report := collectPublicTestReport(
		string(output),
		publicTests,
		"func Solve(input string) string",
		false,
	)

	if report.TestCases[0].Actual != `"HELLO"` || !report.TestCases[0].ActualAvailable {
		t.Fatalf("successful public result = %+v", report.TestCases[0])
	}
	if report.TestCases[1].ActualAvailable || report.TestCases[1].Error != "panic: boom" {
		t.Fatalf("panicking public result = %+v", report.TestCases[1])
	}
	if !strings.Contains(report.ConsoleOutput, "processing hello") ||
		!strings.Contains(report.ConsoleOutput, "processing panic") {
		t.Fatalf("console output = %q", report.ConsoleOutput)
	}
}

func TestCollectPublicTestReportIncludesValuesAndConsole(t *testing.T) {
	publicTests := []problems.PublicTest{
		testPublicCase([]string{`"level"`}, `true`),
		testPublicCase([]string{`"Code"`}, `false`),
	}
	output := strings.Join([]string{
		`__CODEBATTLE_PUBLIC_RESULT__{"Index":1,"Actual":"true","ActualAvailable":true,"Passed":true,"Console":"checking level\n"}`,
		`__CODEBATTLE_PUBLIC_RESULT__{"Index":2,"Actual":"true","ActualAvailable":true,"Passed":false,"Console":"checking Code\n"}`,
		"--- PASS: TestHidden (0.00s)",
	}, "\n")

	report := collectPublicTestReport(output, publicTests, "func Solve(text string) bool", false)
	results := report.TestCases
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Status != "passed" || results[0].Actual != "true" {
		t.Fatalf("first public result = %+v", results[0])
	}
	if results[1].Status != "failed" || results[1].Actual != "true" {
		t.Fatalf("second public result = %+v", results[1])
	}
	if report.ConsoleOutput != "Пример 1:\nchecking level\n\n\nПример 2:\nchecking Code\n" {
		t.Fatalf("console output = %q", report.ConsoleOutput)
	}
	if got := failedTestMessage(results); got != "Публичный пример 2 не пройден" {
		t.Fatalf("message = %q", got)
	}
}

func TestCollectPublicTestReportUsesSuccessfulExitWhenMarkerIsMissing(t *testing.T) {
	report := collectPublicTestReport("", []problems.PublicTest{
		testPublicCase([]string{`"hello"`}, `"world"`),
	}, "func Solve(text string) string", true)
	results := report.TestCases

	if results[0].Status != "passed" || results[0].ActualAvailable {
		t.Fatalf("public result = %+v", results[0])
	}
}

func TestCollectTestReportIncludesPanicAndCapturedConsole(t *testing.T) {
	report := collectPublicTestReport(
		`__CODEBATTLE_PUBLIC_RESULT__{"Index":1,"Actual":"","ActualAvailable":false,"Passed":false,"Console":"before panic\n","RuntimeError":"panic: index out of range"}`,
		[]problems.PublicTest{testPublicCase([]string{`"input"`}, `"output"`)},
		"func Solve(text string) string",
		false,
	)

	if !report.HadRuntimeError {
		t.Fatal("runtime error was not reported")
	}
	if report.ConsoleOutput != "Пример 1:\nbefore panic\n" {
		t.Fatalf("console output = %q", report.ConsoleOutput)
	}
	if report.TestCases[0].ActualAvailable || report.TestCases[0].Error != "panic: index out of range" {
		t.Fatalf("public result = %+v", report.TestCases[0])
	}
}

func testPublicCase(arguments []string, expected string) problems.PublicTest {
	rawArguments := make([]json.RawMessage, len(arguments))
	for index, argument := range arguments {
		rawArguments[index] = json.RawMessage(argument)
	}
	return problems.PublicTest{
		Arguments: rawArguments,
		Expected:  json.RawMessage(expected),
	}
}

func TestHiddenTestResultDoesNotExposeOutput(t *testing.T) {
	failed := hiddenTestResult("private value\n--- FAIL: TestHidden (0.00s)", false)
	if failed.Kind != "hidden" || failed.Status != "failed" {
		t.Fatalf("hidden result = %+v", failed)
	}
	passed := hiddenTestResult("", true)
	if passed.Status != "passed" {
		t.Fatalf("hidden result = %+v", passed)
	}
}
