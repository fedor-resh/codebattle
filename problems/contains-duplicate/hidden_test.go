package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		nums []int
		want bool
	}{
		{[]int{}, false},
		{[]int{7}, false},
		{[]int{-1, 0, -1}, true},
		{[]int{5, 5, 5}, true},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.nums); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
