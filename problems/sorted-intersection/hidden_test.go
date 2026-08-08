package solution

import (
	"slices"
	"testing"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		left  []int
		right []int
		want  []int
	}{
		{[]int{}, []int{1, 2}, []int{}},
		{[]int{-3, -1, 0}, []int{-3, 0, 4}, []int{-3, 0}},
		{[]int{1, 1, 1}, []int{1}, []int{1}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.left, testCase.right); !slices.Equal(got, testCase.want) {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
