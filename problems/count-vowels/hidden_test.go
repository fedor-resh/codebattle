package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"AEIOU", 5},
		{"", 0},
		{"Go is fun", 3},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.input); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
