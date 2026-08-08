package judge

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"codebattle.local/codebattle/internal/problems"
	"codebattle.local/codebattle/internal/submissions"
)

const maxOutputBytes = 64 * 1024

var jobIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type Config struct {
	DockerBinary       string
	Image              string
	SourceDirectory    string
	BinaryDirectory    string
	CacheDirectory     string
	SourceVolume       string
	BinaryVolume       string
	CacheVolume        string
	DisableMemoryLimit bool
}

type Result struct {
	Status      string           `json:"status"`
	Message     string           `json:"message"`
	DurationMS  int64            `json:"duration_ms"`
	PassedTests int              `json:"passed_tests,omitempty"`
	TotalTests  int              `json:"total_tests,omitempty"`
	TestCases   []TestCaseResult `json:"test_cases,omitempty"`
}

type TestCaseResult struct {
	Kind            string `json:"kind"`
	Index           int    `json:"index,omitempty"`
	Status          string `json:"status"`
	Input           string `json:"input,omitempty"`
	Expected        string `json:"expected,omitempty"`
	Actual          string `json:"actual,omitempty"`
	ActualAvailable bool   `json:"actual_available,omitempty"`
	ActualTruncated bool   `json:"actual_truncated,omitempty"`
}

type Runner struct {
	config Config
}

func NewRunner(config Config) *Runner {
	return &Runner{config: config}
}

func (r *Runner) CleanStaleArtifacts() error {
	for _, directory := range []string{r.config.SourceDirectory, r.config.BinaryDirectory} {
		cleaned := filepath.Clean(directory)
		if cleaned == "." || cleaned == string(filepath.Separator) {
			return fmt.Errorf("refusing to clean unsafe judge directory %q", directory)
		}
		entries, err := os.ReadDir(cleaned)
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(cleaned, 0o777); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := os.RemoveAll(filepath.Join(cleaned, entry.Name())); err != nil {
				return err
			}
		}
		_ = os.Chmod(cleaned, 0o777)
	}
	return nil
}

func (r *Runner) WarmCache(ctx context.Context) error {
	if r.config.CacheVolume == "" {
		return nil
	}
	job := submissions.Job{
		Submission: submissions.Submission{ID: "warm-cache"},
		SourceCode: "package solution\n\nfunc Solve(input string) string { return input }\n",
		PublicTests: []problems.PublicTest{
			{Input: "ready", Expected: "ready"},
		},
		HiddenTestSource: `package solution

import "testing"

func TestWarmCache(t *testing.T) {
	if Solve("ready") != "ready" {
		t.Fatal("unexpected result")
	}
}
`,
		MemoryLimitMB: 256,
	}
	if err := r.prepare(job); err != nil {
		return fmt.Errorf("prepare cache warmup: %w", err)
	}
	defer r.cleanup(job.ID)

	warmupCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	output, err, timedOut := r.dockerCommand(
		warmupCtx,
		"codebattle-compile-"+job.ID,
		r.compileArgs(job),
	)
	if timedOut {
		return errors.New("cache warmup exceeded 60 seconds")
	}
	if err != nil {
		return fmt.Errorf("cache warmup: %s", r.sanitizeOutput(output, job.ID))
	}
	return nil
}

func (r *Runner) Run(
	ctx context.Context,
	job submissions.Job,
	onRunning func() error,
) (result Result) {
	started := time.Now()
	result = Result{Status: "internal_error", Message: "Judge временно недоступен"}
	defer func() { result.DurationMS = time.Since(started).Milliseconds() }()

	if !jobIDPattern.MatchString(job.ID) {
		return result
	}
	if err := problems.ValidateSolution(job.SourceCode, "Solve"); err != nil {
		result.Status = "compile_error"
		result.Message = err.Error()
		return result
	}
	if err := r.prepare(job); err != nil {
		return result
	}
	defer r.cleanup(job.ID)

	compileCtx, cancelCompile := context.WithTimeout(ctx, 30*time.Second)
	compileOutput, compileErr, compileTimedOut := r.dockerCommand(
		compileCtx,
		"codebattle-compile-"+job.ID,
		r.compileArgs(job),
	)
	cancelCompile()
	if compileTimedOut {
		result.Status = "compile_error"
		result.Message = "Компиляция превысила лимит 30 секунд"
		return result
	}
	if compileErr != nil {
		if message, infrastructureFailure := infrastructureFailureMessage(compileOutput); infrastructureFailure {
			result.Status = "internal_error"
			result.Message = message
			return result
		}
		result.Status = "compile_error"
		result.Message = r.sanitizeOutput(compileOutput, job.ID)
		if result.Message == "" {
			result.Message = "Код не компилируется: проверьте синтаксис, импорты и сигнатуру Solve"
		}
		return result
	}
	if err := onRunning(); err != nil {
		return result
	}

	testLimit := solutionTimeLimit(job)
	// The solution timeout is enforced by the Go test binary. The outer limit
	// additionally covers Docker startup and shutdown, which can be noticeably
	// slower on a cold host and must not count as the player's execution time.
	testCtx, cancelTest := context.WithTimeout(ctx, testLimit+5*time.Second)
	testOutput, testErr, testTimedOut := r.dockerCommand(
		testCtx,
		"codebattle-run-"+job.ID,
		r.runtimeArgs(job),
	)
	cancelTest()
	result.TestCases = collectTestResults(testOutput, job.PublicTests, testErr == nil)
	result.TotalTests = len(result.TestCases)
	for _, testCase := range result.TestCases {
		if testCase.Status == "passed" {
			result.PassedTests++
		}
	}
	if testTimedOut || strings.Contains(testOutput, "panic: test timed out") {
		result.Status = "time_limit"
		result.Message = "Превышен лимит времени"
		return result
	}
	if testErr != nil {
		if strings.Contains(testOutput, "panic:") || strings.Contains(testOutput, "fatal error:") {
			result.Status = "runtime_error"
			result.Message = "Ошибка выполнения"
		} else {
			result.Status = "wrong_answer"
			result.Message = failedTestMessage(result.TestCases)
		}
		return result
	}
	result.Status = "accepted"
	result.Message = "Все тесты пройдены"
	return result
}

func (r *Runner) prepare(job submissions.Job) error {
	sourceDirectory := filepath.Join(r.config.SourceDirectory, job.ID)
	binaryDirectory := filepath.Join(r.config.BinaryDirectory, job.ID)
	if err := os.RemoveAll(sourceDirectory); err != nil {
		return err
	}
	if err := os.RemoveAll(binaryDirectory); err != nil {
		return err
	}
	if err := os.MkdirAll(sourceDirectory, 0o777); err != nil {
		return err
	}
	if err := os.MkdirAll(binaryDirectory, 0o777); err != nil {
		return err
	}
	if r.config.CacheDirectory != "" {
		if err := os.MkdirAll(r.config.CacheDirectory, 0o777); err != nil {
			return err
		}
		_ = os.Chmod(r.config.CacheDirectory, 0o777)
	}
	_ = os.Chmod(sourceDirectory, 0o777)
	_ = os.Chmod(binaryDirectory, 0o777)
	files := map[string]string{
		"go.mod":         "module solution\n\ngo 1.26.0\n",
		"solution.go":    job.SourceCode,
		"hidden_test.go": job.HiddenTestSource,
	}
	if len(job.PublicTests) > 0 {
		files["public_test.go"] = publicTestSource(job.PublicTests)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(sourceDirectory, name), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) cleanup(jobID string) {
	if !jobIDPattern.MatchString(jobID) {
		return
	}
	_ = os.RemoveAll(filepath.Join(r.config.SourceDirectory, jobID))
	_ = os.RemoveAll(filepath.Join(r.config.BinaryDirectory, jobID))
}

func (r *Runner) compileArgs(job submissions.Job) []string {
	memory := strconv.Itoa(max(job.MemoryLimitMB*2, 512)) + "m"
	args := []string{
		"run", "--rm", "--network", "none", "--read-only",
	}
	if !r.config.DisableMemoryLimit {
		args = append(args, "--memory", memory, "--memory-swap", memory)
	}
	return append(args,
		"--cpus", "2",
		"--pids-limit", "64", "--cap-drop", "ALL",
		"--security-opt", "no-new-privileges", "--user", "65532:65532",
		"--tmpfs", "/tmp:rw,nosuid,nodev,size=128m",
		"-e", "GOCACHE=/cache/go-build", "-e", "GOMAXPROCS=2", "-e", "CGO_ENABLED=0",
		"-v", r.config.SourceVolume+":/src:ro",
		"-v", r.config.BinaryVolume+":/out",
		"-v", r.config.CacheVolume+":/cache",
		"-w", "/src/"+job.ID,
		r.config.Image,
		"go", "test", "-c", "-trimpath", "-o", "/out/"+job.ID+"/tests", ".",
	)
}

func (r *Runner) runtimeArgs(job submissions.Job) []string {
	memory := strconv.Itoa(max(job.MemoryLimitMB, 32)) + "m"
	args := []string{
		"run", "--rm", "--network", "none", "--read-only",
	}
	if !r.config.DisableMemoryLimit {
		args = append(args, "--memory", memory, "--memory-swap", memory)
	}
	return append(args,
		"--cpus", "1",
		"--pids-limit", "64", "--cap-drop", "ALL",
		"--security-opt", "no-new-privileges", "--user", "65532:65532",
		"--tmpfs", "/tmp:rw,nosuid,nodev,size=16m",
		"-v", r.config.BinaryVolume+":/out:ro",
		"-w", "/out/"+job.ID,
		r.config.Image,
		"./tests", "-test.v", "-test.timeout="+solutionTimeLimit(job).String(),
	)
}

func solutionTimeLimit(job submissions.Job) time.Duration {
	limit := time.Duration(job.TimeLimitMS) * time.Millisecond
	if limit <= 0 {
		return 2 * time.Second
	}
	return limit
}

func (r *Runner) dockerCommand(ctx context.Context, name string, args []string) (string, error, bool) {
	args = append(args[:1], append([]string{"--name", name}, args[1:]...)...)
	command := exec.CommandContext(ctx, r.config.DockerBinary, args...)
	buffer := &limitedBuffer{limit: maxOutputBytes}
	command.Stdout = buffer
	command.Stderr = buffer
	err := command.Run()
	timedOut := errors.Is(ctx.Err(), context.DeadlineExceeded)
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = exec.CommandContext(cleanupCtx, r.config.DockerBinary, "rm", "-f", name).Run()
	return buffer.String(), err, timedOut
}

func (r *Runner) sanitizeOutput(output, jobID string) string {
	replacer := strings.NewReplacer(
		r.config.SourceDirectory, "",
		r.config.BinaryDirectory, "",
		"/src/"+jobID, "",
		"/out/"+jobID, "",
		"hidden_test.go", "tests.go",
	)
	cleanedLines := make([]string, 0)
	for _, line := range strings.Split(replacer.Replace(output), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "# solution [solution.test]" ||
			strings.Contains(trimmed, "kernel does not support memory limit capabilities") {
			continue
		}
		cleanedLines = append(cleanedLines, line)
	}
	cleaned := strings.TrimSpace(strings.Join(cleanedLines, "\n"))
	if len(cleaned) > maxOutputBytes {
		cleaned = cleaned[:maxOutputBytes]
	}
	return cleaned
}

func DetectMemoryLimitSupport(ctx context.Context, dockerBinary string) (bool, error) {
	output, err := exec.CommandContext(ctx, dockerBinary, "info", "--format", "{{.MemoryLimit}}").Output()
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(string(output)), "true"), nil
}

func infrastructureFailureMessage(output string) (string, bool) {
	lower := strings.ToLower(output)
	switch {
	case strings.Contains(lower, "no space left on device"):
		return "Judge временно недоступен: на сервере закончилось место", true
	case strings.Contains(lower, "cannot connect to the docker daemon"),
		strings.Contains(lower, "error during connect"),
		strings.Contains(lower, "error response from daemon"):
		return "Judge временно недоступен: Docker не смог запустить проверку", true
	default:
		return "", false
	}
}

type limitedBuffer struct {
	buffer bytes.Buffer
	limit  int
}

func (b *limitedBuffer) Write(value []byte) (int, error) {
	originalLength := len(value)
	remaining := b.limit - b.buffer.Len()
	if remaining > 0 {
		if len(value) > remaining {
			value = value[:remaining]
		}
		_, _ = b.buffer.Write(value)
	}
	return originalLength, nil
}

func (b *limitedBuffer) String() string {
	return b.buffer.String()
}
