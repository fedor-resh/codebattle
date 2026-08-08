package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		n    int
		want uint64
	}{
		{1, 1},
		{20, 6765},
		{50, 12586269025},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.n); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
