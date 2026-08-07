package judge

import (
	"bytes"
	"context"
	"errors"
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
	DockerBinary    string
	Image           string
	SourceDirectory string
	BinaryDirectory string
	SourceVolume    string
	BinaryVolume    string
}

type Result struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	DurationMS int64  `json:"duration_ms"`
}

type Runner struct {
	config Config
}

func NewRunner(config Config) *Runner {
	return &Runner{config: config}
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
		result.Status = "compile_error"
		result.Message = r.sanitizeOutput(compileOutput, job.ID)
		if result.Message == "" {
			result.Message = "Код не компилируется"
		}
		return result
	}
	if err := onRunning(); err != nil {
		return result
	}

	testLimit := time.Duration(job.TimeLimitMS) * time.Millisecond
	if testLimit <= 0 {
		testLimit = 2 * time.Second
	}
	testCtx, cancelTest := context.WithTimeout(ctx, testLimit)
	testOutput, testErr, testTimedOut := r.dockerCommand(
		testCtx,
		"codebattle-run-"+job.ID,
		r.runtimeArgs(job),
	)
	cancelTest()
	if testTimedOut {
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
			result.Message = "Скрытый тест не пройден"
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
	_ = os.Chmod(sourceDirectory, 0o777)
	_ = os.Chmod(binaryDirectory, 0o777)
	files := map[string]string{
		"go.mod":         "module solution\n\ngo 1.26.0\n",
		"solution.go":    job.SourceCode,
		"hidden_test.go": job.HiddenTestSource,
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
	return []string{
		"run", "--rm", "--network", "none", "--read-only",
		"--memory", memory, "--memory-swap", memory, "--cpus", "1",
		"--pids-limit", "64", "--cap-drop", "ALL",
		"--security-opt", "no-new-privileges", "--user", "65532:65532",
		"--tmpfs", "/tmp:rw,nosuid,nodev,size=128m",
		"-e", "GOCACHE=/tmp/go-cache", "-e", "CGO_ENABLED=0",
		"-v", r.config.SourceVolume + ":/src:ro",
		"-v", r.config.BinaryVolume + ":/out",
		"-w", "/src/" + job.ID,
		r.config.Image,
		"go", "test", "-c", "-trimpath", "-o", "/out/" + job.ID + "/tests", ".",
	}
}

func (r *Runner) runtimeArgs(job submissions.Job) []string {
	memory := strconv.Itoa(max(job.MemoryLimitMB, 32)) + "m"
	return []string{
		"run", "--rm", "--network", "none", "--read-only",
		"--memory", memory, "--memory-swap", memory, "--cpus", "1",
		"--pids-limit", "64", "--cap-drop", "ALL",
		"--security-opt", "no-new-privileges", "--user", "65532:65532",
		"--tmpfs", "/tmp:rw,nosuid,nodev,size=16m",
		"-v", r.config.BinaryVolume + ":/out:ro",
		"-w", "/out/" + job.ID,
		r.config.Image,
		"./tests", "-test.v",
	}
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
	cleaned := strings.TrimSpace(replacer.Replace(output))
	if len(cleaned) > maxOutputBytes {
		cleaned = cleaned[:maxOutputBytes]
	}
	return cleaned
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
