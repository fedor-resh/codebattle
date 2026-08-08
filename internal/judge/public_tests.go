package judge

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"codebattle.local/codebattle/internal/problems"
)

const publicResultMarker = "__CODEBATTLE_PUBLIC_RESULT__"

type publicResultRecord struct {
	Index            int
	Actual           string
	ActualAvailable  bool
	Passed           bool
	Truncated        bool
	Console          string
	ConsoleTruncated bool
	RuntimeError     string
}

type publicTestReport struct {
	TestCases              []TestCaseResult
	ConsoleOutput          string
	ConsoleOutputTruncated bool
	HadRuntimeError        bool
}

// publicTestSource creates a separate external test package. This keeps the
// reporting helpers isolated from identifiers declared by a player's solution.
func publicTestSource(tests []problems.PublicTest) string {
	var cases strings.Builder
	for _, testCase := range tests {
		fmt.Fprintf(
			&cases,
			"\t\t{input: %s, expected: %s},\n",
			strconv.Quote(testCase.Input),
			strconv.Quote(testCase.Expected),
		)
	}

	return fmt.Sprintf(`package solution_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	codebattleSolution "solution"
	"testing"
)

const codebattleActualLimit = 2048
const codebattleConsoleLimit = 4096

type codebattlePublicResult struct {
	Index            int
	Actual           string
	ActualAvailable  bool
	Passed           bool
	Truncated        bool
	Console          string
	ConsoleTruncated bool
	RuntimeError     string
}

type codebattleLimitedWriter struct {
	value     []byte
	truncated bool
}

func (writer *codebattleLimitedWriter) Write(value []byte) (int, error) {
	originalLength := len(value)
	remaining := codebattleConsoleLimit - len(writer.value)
	if remaining > 0 {
		if len(value) > remaining {
			writer.value = append(writer.value, value[:remaining]...)
			writer.truncated = true
		} else {
			writer.value = append(writer.value, value...)
		}
	} else if len(value) > 0 {
		writer.truncated = true
	}
	return originalLength, nil
}

func codebattleDisplayValue(value string) (string, bool) {
	if len(value) <= codebattleActualLimit {
		return value, false
	}
	return value[:codebattleActualLimit], true
}

func codebattleRunSolution(input string) (
	actual string,
	actualAvailable bool,
	console string,
	consoleTruncated bool,
	runtimeError string,
) {
	read, write, err := os.Pipe()
	if err != nil {
		return "", false, "", false, "не удалось захватить вывод программы"
	}

	previousStdout, previousStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = write, write
	captured := &codebattleLimitedWriter{}
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(captured, read)
		close(done)
	}()

	defer func() {
		_ = write.Close()
		os.Stdout, os.Stderr = previousStdout, previousStderr
		<-done
		_ = read.Close()
		console = string(captured.value)
		consoleTruncated = captured.truncated
		if recovered := recover(); recovered != nil {
			actualAvailable = false
			runtimeError, _ = codebattleDisplayValue(fmt.Sprintf("panic: %%v", recovered))
		}
	}()

	actual = codebattleSolution.Solve(input)
	actualAvailable = true
	return
}

func TestCodebattlePublic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
%s	}

	for index, testCase := range tests {
		actual, actualAvailable, console, consoleTruncated, runtimeError := codebattleRunSolution(testCase.input)
		displayActual, truncated := codebattleDisplayValue(actual)
		record, _ := json.Marshal(codebattlePublicResult{
			Index: index + 1, Actual: displayActual, ActualAvailable: actualAvailable,
			Passed: actualAvailable && actual == testCase.expected, Truncated: truncated,
			Console: console, ConsoleTruncated: consoleTruncated, RuntimeError: runtimeError,
		})
		fmt.Printf("%s%%s\n", record)
		if runtimeError != "" {
			t.Errorf("public case %%d panicked", index+1)
		} else if actual != testCase.expected {
			t.Errorf("public case %%d failed", index+1)
		}
	}
}
`, cases.String(), publicResultMarker)
}

func collectPublicTestReport(
	output string,
	publicTests []problems.PublicTest,
	executionPassed bool,
) publicTestReport {
	results := make([]TestCaseResult, len(publicTests), len(publicTests)+1)
	consoleParts := make([]string, 0, len(publicTests))
	consoleTruncated := false
	hadRuntimeError := false
	for index, testCase := range publicTests {
		results[index] = TestCaseResult{
			Kind:     "public",
			Index:    index + 1,
			Status:   "not_run",
			Input:    testCase.Input,
			Expected: testCase.Expected,
		}
	}

	for _, line := range strings.Split(output, "\n") {
		markerIndex := strings.Index(line, publicResultMarker)
		if markerIndex < 0 {
			continue
		}
		var record publicResultRecord
		if err := json.Unmarshal([]byte(line[markerIndex+len(publicResultMarker):]), &record); err != nil {
			continue
		}
		if record.Index < 1 || record.Index > len(results) {
			continue
		}
		result := &results[record.Index-1]
		result.Actual = record.Actual
		result.ActualAvailable = record.ActualAvailable
		result.ActualTruncated = record.Truncated
		result.Error = record.RuntimeError
		if record.Console != "" {
			consoleParts = append(consoleParts, fmt.Sprintf("Пример %d:\n%s", record.Index, record.Console))
		}
		consoleTruncated = consoleTruncated || record.ConsoleTruncated
		hadRuntimeError = hadRuntimeError || record.RuntimeError != ""
		passed := record.Passed
		if record.ActualAvailable && !record.Truncated {
			passed = record.Actual == result.Expected
		}
		if passed {
			result.Status = "passed"
		} else {
			result.Status = "failed"
		}
	}

	if executionPassed {
		for index := range results {
			// The binary could fill the output limit before the reporter writes its
			// marker. A successful test process still proves that the case passed.
			if results[index].Status == "not_run" {
				results[index].Status = "passed"
			}
		}
	}

	return publicTestReport{
		TestCases:              results,
		ConsoleOutput:          strings.Join(consoleParts, "\n\n"),
		ConsoleOutputTruncated: consoleTruncated,
		HadRuntimeError:        hadRuntimeError,
	}
}

func hiddenTestResult(output string, executionPassed bool) TestCaseResult {
	status := "not_run"
	if executionPassed || strings.Contains(output, "--- PASS: TestHidden") {
		status = "passed"
	} else if strings.Contains(output, "--- FAIL: TestHidden") {
		status = "failed"
	}
	return TestCaseResult{Kind: "hidden", Status: status}
}

func failedTestMessage(testCases []TestCaseResult) string {
	for _, testCase := range testCases {
		if testCase.Kind == "public" && testCase.Status == "failed" {
			return fmt.Sprintf("Публичный пример %d не пройден", testCase.Index)
		}
	}
	return "Скрытый тест не пройден"
}
