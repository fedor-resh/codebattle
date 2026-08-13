package solution

import (
	"slices"
	"testing"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		nums    []int
		workers int
		want    []int
	}{
		{nil, -1, []int{}},
		{[]int{7}, 12, []int{49}},
		{[]int{9, -8, 7, -6, 5, -4}, 1, []int{81, 64, 49, 36, 25, 16}},
		{[]int{1, 10, 2, 9, 3, 8}, 4, []int{1, 100, 4, 81, 9, 64}},
	}
	for attempt := range 20 {
		for _, testCase := range cases {
			if got := Solve(testCase.nums, testCase.workers); !slices.Equal(got, testCase.want) {
				t.Fatalf("attempt %d: got %v, want %v", attempt, got, testCase.want)
			}
		}
	}
}
