package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		nums   []int
		target int
		want   int
	}{
		{[]int{1}, 1, 0},
		{[]int{1, 3, 7, 9}, 9, 3},
		{[]int{-5, -1, 2}, 0, -1},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.nums, testCase.target); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
