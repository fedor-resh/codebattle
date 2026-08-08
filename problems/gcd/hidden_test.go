package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		a    int
		b    int
		want int
	}{
		{0, 5, 5},
		{270, 192, 6},
		{13, 13, 13},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.a, testCase.b); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
