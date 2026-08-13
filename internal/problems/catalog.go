package problems

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var requiredFiles = []string{
	"problem.yaml",
	"statement.md",
	"starter.go",
	"public_tests.json",
	"hidden_test.go",
	"reference.go",
}

type Metadata struct {
	Slug          string       `yaml:"slug"`
	Title         string       `yaml:"title"`
	Difficulty    string       `yaml:"difficulty"`
	Class         Class        `yaml:"class,omitempty"`
	Requirements  Requirements `yaml:"requirements,omitempty"`
	Function      string       `yaml:"function"`
	Signature     string       `yaml:"signature"`
	Version       int          `yaml:"version"`
	TimeLimitMS   int          `yaml:"time_limit_ms"`
	MemoryLimitMB int          `yaml:"memory_limit_mb"`
}

type Class string

const (
	ClassAlgorithms  Class = "algorithms"
	ClassConcurrency Class = "concurrency"
)

func IsValidClass(value Class) bool {
	return value == ClassAlgorithms || value == ClassConcurrency
}

type Requirements struct {
	Goroutine     bool `yaml:"goroutine,omitempty" json:"goroutine,omitempty"`
	Channel       bool `yaml:"channel,omitempty" json:"channel,omitempty"`
	WaitGroup     bool `yaml:"wait_group,omitempty" json:"wait_group,omitempty"`
	Mutex         bool `yaml:"mutex,omitempty" json:"mutex,omitempty"`
	Select        bool `yaml:"select,omitempty" json:"select,omitempty"`
	ContextCancel bool `yaml:"context_cancel,omitempty" json:"context_cancel,omitempty"`
}

func (requirements Requirements) Empty() bool {
	return !requirements.Goroutine && !requirements.Channel && !requirements.WaitGroup &&
		!requirements.Mutex && !requirements.Select && !requirements.ContextCancel
}

type PublicTest struct {
	Arguments []json.RawMessage `json:"arguments"`
	Expected  json.RawMessage   `json:"expected"`
}

func (test *PublicTest) UnmarshalJSON(data []byte) error {
	var value struct {
		Arguments []json.RawMessage `json:"arguments"`
		Input     *string           `json:"input"`
		Expected  json.RawMessage   `json:"expected"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value.Arguments == nil && value.Input != nil {
		encoded, _ := json.Marshal(*value.Input)
		value.Arguments = []json.RawMessage{encoded}
	}
	if value.Arguments == nil {
		return errors.New("arguments are required")
	}
	if len(value.Expected) == 0 {
		return errors.New("expected is required")
	}
	test.Arguments = value.Arguments
	test.Expected = value.Expected
	return nil
}

type Problem struct {
	Metadata
	ID            string
	Statement     string
	Starter       string
	PublicTests   []PublicTest
	HiddenTest    string
	Reference     string
	ContentHash   string
	PublicTestRaw json.RawMessage
}

func LoadCatalog(ctx context.Context, root string, validateReferences bool) ([]Problem, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read problem directory: %w", err)
	}

	problems := make([]Problem, 0, len(entries))
	seenSlugs := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		problem, err := loadProblem(filepath.Join(root, entry.Name()), entry.Name())
		if err != nil {
			return nil, err
		}
		if seenSlugs[problem.Slug] {
			return nil, fmt.Errorf("duplicate problem slug %q", problem.Slug)
		}
		seenSlugs[problem.Slug] = true
		if validateReferences {
			if err := validateReference(ctx, problem); err != nil {
				return nil, fmt.Errorf("problem %s: %w", problem.Slug, err)
			}
		}
		problems = append(problems, problem)
	}
	if len(problems) == 0 {
		return nil, errors.New("problem catalog is empty")
	}
	sort.Slice(problems, func(i, j int) bool { return problems[i].Slug < problems[j].Slug })
	return problems, nil
}

func loadProblem(directory, directoryName string) (Problem, error) {
	content := make(map[string][]byte, len(requiredFiles))
	hasher := sha256.New()
	for _, name := range requiredFiles {
		value, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			return Problem{}, fmt.Errorf("problem %s: read %s: %w", directoryName, name, err)
		}
		content[name] = value
		_, _ = hasher.Write([]byte(name))
		_, _ = hasher.Write([]byte{0})
		_, _ = hasher.Write(value)
		_, _ = hasher.Write([]byte{0})
	}

	var metadata Metadata
	decoder := yaml.NewDecoder(bytes.NewReader(content["problem.yaml"]))
	decoder.KnownFields(true)
	if err := decoder.Decode(&metadata); err != nil {
		return Problem{}, fmt.Errorf("problem %s: decode metadata: %w", directoryName, err)
	}
	if metadata.Class == "" {
		metadata.Class = ClassAlgorithms
	}
	if err := validateMetadata(metadata, directoryName); err != nil {
		return Problem{}, err
	}

	var publicTests []PublicTest
	if err := json.Unmarshal(content["public_tests.json"], &publicTests); err != nil {
		return Problem{}, fmt.Errorf("problem %s: decode public tests: %w", metadata.Slug, err)
	}
	if len(publicTests) == 0 {
		return Problem{}, fmt.Errorf("problem %s: public tests are empty", metadata.Slug)
	}
	if strings.TrimSpace(string(content["statement.md"])) == "" {
		return Problem{}, fmt.Errorf("problem %s: statement is empty", metadata.Slug)
	}
	if err := ValidateSolution(string(content["starter.go"]), metadata.Signature); err != nil {
		return Problem{}, fmt.Errorf("problem %s starter: %w", metadata.Slug, err)
	}
	if err := ValidateSolution(string(content["reference.go"]), metadata.Signature); err != nil {
		return Problem{}, fmt.Errorf("problem %s reference: %w", metadata.Slug, err)
	}
	if err := ValidateRequirements(string(content["reference.go"]), metadata.Requirements); err != nil {
		return Problem{}, fmt.Errorf("problem %s reference: %w", metadata.Slug, err)
	}
	schema, _ := ParseSignature(metadata.Signature)
	if err := ValidatePublicTests(schema, publicTests); err != nil {
		return Problem{}, fmt.Errorf("problem %s: %w", metadata.Slug, err)
	}

	contentHash := hex.EncodeToString(hasher.Sum(nil))
	publicRaw, _ := json.Marshal(publicTests)
	return Problem{
		Metadata:      metadata,
		ID:            contentHash[:32],
		Statement:     string(content["statement.md"]),
		Starter:       string(content["starter.go"]),
		PublicTests:   publicTests,
		HiddenTest:    string(content["hidden_test.go"]),
		Reference:     string(content["reference.go"]),
		ContentHash:   contentHash,
		PublicTestRaw: publicRaw,
	}, nil
}

func validateMetadata(metadata Metadata, directoryName string) error {
	if !slugPattern.MatchString(metadata.Slug) || metadata.Slug != directoryName {
		return fmt.Errorf("problem %s: invalid or mismatched slug %q", directoryName, metadata.Slug)
	}
	if strings.TrimSpace(metadata.Title) == "" {
		return fmt.Errorf("problem %s: title is empty", metadata.Slug)
	}
	if metadata.Difficulty != "easy" && metadata.Difficulty != "medium" && metadata.Difficulty != "hard" {
		return fmt.Errorf("problem %s: invalid difficulty", metadata.Slug)
	}
	if !IsValidClass(metadata.Class) {
		return fmt.Errorf("problem %s: invalid class", metadata.Slug)
	}
	if metadata.Class == ClassAlgorithms && !metadata.Requirements.Empty() {
		return fmt.Errorf("problem %s: algorithm tasks cannot declare concurrency requirements", metadata.Slug)
	}
	if metadata.Class == ClassConcurrency && !metadata.Requirements.Goroutine {
		return fmt.Errorf("problem %s: concurrency tasks must require a goroutine", metadata.Slug)
	}
	if metadata.Function != "Solve" {
		return fmt.Errorf("problem %s: function must be Solve", metadata.Slug)
	}
	if _, err := ParseSignature(metadata.Signature); err != nil {
		return fmt.Errorf("problem %s: %w", metadata.Slug, err)
	}
	if metadata.Version <= 0 || metadata.TimeLimitMS <= 0 || metadata.MemoryLimitMB <= 0 {
		return fmt.Errorf("problem %s: version and limits must be positive", metadata.Slug)
	}
	return nil
}

func validateReference(ctx context.Context, problem Problem) error {
	directory, err := os.MkdirTemp("", "codebattle-problem-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(directory)

	schema, err := ParseSignature(problem.Signature)
	if err != nil {
		return err
	}
	publicTest, err := publicValidationTestSource(schema, problem.PublicTests)
	if err != nil {
		return err
	}
	files := map[string]string{
		"go.mod":         "module solution\n\ngo 1.26.0\n",
		"solution.go":    problem.Reference,
		"public_test.go": publicTest,
		"hidden_test.go": problem.HiddenTest,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600); err != nil {
			return err
		}
	}
	command := exec.CommandContext(ctx, "go", "test", "./...")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("reference tests failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
