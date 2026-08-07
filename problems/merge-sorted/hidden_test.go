package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"|1,2", "1,2"},
		{"-3,0|0,4", "-3,0,0,4"},
		{"1,2|", "1,2"},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.input); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
