package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		nums    []int
		workers int
		want    int
	}{
		{nil, -2, 0},
		{[]int{42}, 20, 42},
		{[]int{-10, -20, 5, 7, 18}, 1, 0},
		{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9}, 4, 45},
	}
	for attempt := range 20 {
		for _, testCase := range cases {
			if got := Solve(testCase.nums, testCase.workers); got != testCase.want {
				t.Fatalf("attempt %d: got %d, want %d", attempt, got, testCase.want)
			}
		}
	}
}
