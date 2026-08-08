package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		text  string
		shift int
		want  string
	}{
		{"Go!", 0, "Go!"},
		{"abc", 26, "abc"},
		{"b", 25, "a"},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.text, testCase.shift); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
