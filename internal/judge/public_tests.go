package judge

import (
	"encoding/json"
	"fmt"
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
func publicTestSource(tests []problems.PublicTest, functionSignature string) (string, error) {
	schema, err := problems.ParseSignature(functionSignature)
	if err != nil {
		return "", err
	}
	if err := problems.ValidatePublicTests(schema, tests); err != nil {
		return "", err
	}
	var cases strings.Builder
	for _, testCase := range tests {
		arguments := make([]string, len(testCase.Arguments))
		for index, argument := range testCase.Arguments {
			arguments[index], _ = problems.GoLiteral(schema.Params[index].Type, argument)
		}
		expected, _ := problems.GoLiteral(schema.Result, testCase.Expected)
		fmt.Fprintf(
			&cases,
			"\t\t{run: func() %s { return codebattleSolution.Solve(%s) }, expected: %s},\n",
			schema.Result,
			strings.Join(arguments, ", "),
			expected,
		)
	}
	optionalImports := ""
	if strings.Contains(cases.String(), "context.") {
		optionalImports += "\t\"context\"\n"
	}
	if strings.Contains(cases.String(), "time.") {
		optionalImports += "\t\"time\"\n"
	}

	return fmt.Sprintf(`package solution_test

import (
%s
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
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

func codebattleFormatValue(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "не удалось отобразить значение"
	}
	return string(encoded)
}

func codebattleRunSolution[T any](run func() T) (
	actual T,
	actualAvailable bool,
	console string,
	consoleTruncated bool,
	runtimeError string,
) {
	read, write, err := os.Pipe()
	if err != nil {
		var zero T
		return zero, false, "", false, "не удалось захватить вывод программы"
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

	actual = run()
	actualAvailable = true
	return
}

func TestCodebattlePublic(t *testing.T) {
	tests := []struct {
		run      func() %s
		expected %s
	}{
%s	}

	for index, testCase := range tests {
		actual, actualAvailable, console, consoleTruncated, runtimeError := codebattleRunSolution(testCase.run)
		displayActual, truncated := "", false
		if actualAvailable {
			displayActual, truncated = codebattleDisplayValue(codebattleFormatValue(actual))
		}
		passed := actualAvailable && reflect.DeepEqual(actual, testCase.expected)
		record, _ := json.Marshal(codebattlePublicResult{
			Index: index + 1, Actual: displayActual, ActualAvailable: actualAvailable,
			Passed: passed, Truncated: truncated,
			Console: console, ConsoleTruncated: consoleTruncated, RuntimeError: runtimeError,
		})
		fmt.Printf("%s%%s\n", record)
		if runtimeError != "" {
			t.Errorf("public case %%d panicked", index+1)
		} else if !passed {
			t.Errorf("public case %%d failed", index+1)
		}
	}
}
`, optionalImports, schema.Result, schema.Result, cases.String(), publicResultMarker), nil
}

func collectPublicTestReport(
	output string,
	publicTests []problems.PublicTest,
	functionSignature string,
	executionPassed bool,
) publicTestReport {
	results := make([]TestCaseResult, len(publicTests), len(publicTests)+1)
	schema, _ := problems.ParseSignature(functionSignature)
	consoleParts := make([]string, 0, len(publicTests))
	consoleTruncated := false
	hadRuntimeError := false
	for index, testCase := range publicTests {
		results[index] = TestCaseResult{
			Kind:     "public",
			Index:    index + 1,
			Status:   "not_run",
			Input:    problems.DisplayArguments(schema, testCase),
			Expected: problems.DisplayJSON(testCase.Expected),
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
