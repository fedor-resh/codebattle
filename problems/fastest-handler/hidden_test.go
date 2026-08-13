package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		delays  []int
		timeout int
		want    int
	}{
		{nil, 10, -1},
		{[]int{80, 50, 1, 30}, 250, 2},
		{[]int{25, 15, 35}, 3, -1},
		{[]int{12}, 100, 0},
		{[]int{8, 2, 5}, -1, -1},
	}
	for attempt := range 10 {
		for _, testCase := range cases {
			if got := Solve(testCase.delays, testCase.timeout); got != testCase.want {
				t.Fatalf("attempt %d: got %d, want %d", attempt, got, testCase.want)
			}
		}
	}
}
