package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		nums []int
		want int
	}{
		{[]int{}, 0},
		{[]int{42}, 42},
		{[]int{-1, -2, -3}, -6},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.nums); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
