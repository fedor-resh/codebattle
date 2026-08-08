package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		nums []int
		want int
	}{
		{[]int{5}, 5},
		{[]int{0, -1, 1}, 1},
		{[]int{-100, -200}, -100},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.nums); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
