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
	Index     int    `json:"index"`
	Actual    string `json:"actual"`
	Passed    bool   `json:"passed"`
	Truncated bool   `json:"truncated,omitempty"`
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
	codebattleSolution "solution"
	"testing"
)

const codebattleActualLimit = 2048

type codebattlePublicResult struct {
	Index     int
	Actual    string
	Passed    bool
	Truncated bool
}

func codebattleDisplayValue(value string) (string, bool) {
	if len(value) <= codebattleActualLimit {
		return value, false
	}
	return value[:codebattleActualLimit], true
}

func TestCodebattlePublic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
%s	}

	for index, testCase := range tests {
		actual := codebattleSolution.Solve(testCase.input)
		displayActual, truncated := codebattleDisplayValue(actual)
		record, _ := json.Marshal(codebattlePublicResult{
			Index: index + 1, Actual: displayActual,
			Passed: actual == testCase.expected, Truncated: truncated,
		})
		fmt.Printf("%s%%s\n", record)
		if actual != testCase.expected {
			t.Errorf("public case %%d failed", index+1)
		}
	}
}
`, cases.String(), publicResultMarker)
}

func collectTestResults(
	output string,
	publicTests []problems.PublicTest,
	executionPassed bool,
) []TestCaseResult {
	results := make([]TestCaseResult, len(publicTests), len(publicTests)+1)
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
		result.ActualAvailable = true
		result.ActualTruncated = record.Truncated
		passed := record.Passed
		if !record.Truncated {
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

	hiddenStatus := "not_run"
	if executionPassed || strings.Contains(output, "--- PASS: TestHidden") {
		hiddenStatus = "passed"
	} else if strings.Contains(output, "--- FAIL: TestHidden") {
		hiddenStatus = "failed"
	}
	results = append(results, TestCaseResult{Kind: "hidden", Status: hiddenStatus})
	return results
}

func failedTestMessage(testCases []TestCaseResult) string {
	for _, testCase := range testCases {
		if testCase.Kind == "public" && testCase.Status == "failed" {
			return fmt.Sprintf("Публичный пример %d не пройден", testCase.Index)
		}
	}
	return "Скрытый тест не пройден"
}
