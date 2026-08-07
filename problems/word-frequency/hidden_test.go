package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"x", "x:1"},
		{"A a A", "A:2,a:1"},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.input); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
