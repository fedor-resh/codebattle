package solution

import (
	"slices"
	"testing"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		nums []int
		want []int
	}{
		{[]int{}, []int{}},
		{[]int{5}, []int{5}},
		{[]int{-2, -3, 10}, []int{-2, -5, 5}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.nums); !slices.Equal(got, testCase.want) {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
