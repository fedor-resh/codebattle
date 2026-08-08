package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		n    int
		want uint64
	}{
		{1, 1},
		{10, 3628800},
		{20, 2432902008176640000},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.n); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
