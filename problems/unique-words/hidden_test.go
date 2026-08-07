package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"one", "one"},
		{"A a A", "A,a"},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.input); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
