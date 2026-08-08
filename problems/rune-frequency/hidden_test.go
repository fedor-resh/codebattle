package solution

import (
	"maps"
	"testing"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		text string
		want map[string]int
	}{
		{"", map[string]int{}},
		{"аа!", map[string]int{"а": 2, "!": 1}},
		{"A a", map[string]int{"A": 1, " ": 1, "a": 1}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.text); !maps.Equal(got, testCase.want) {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
