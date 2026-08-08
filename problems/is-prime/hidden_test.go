package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		n    int
		want bool
	}{
		{-7, false},
		{0, false},
		{1, false},
		{97, true},
		{100, false},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.n); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
