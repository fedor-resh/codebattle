package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"5", "5"},
		{"0 -1 1", "1"},
		{"-100 -200", "-100"},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.input); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
