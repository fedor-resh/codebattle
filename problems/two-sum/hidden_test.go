package solution

import (
	"reflect"
	"testing"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		nums   []int
		target int
		want   []int
	}{
		{[]int{-1, 1}, 0, []int{0, 1}},
		{[]int{5, 5}, 10, []int{0, 1}},
		{[]int{1, 2, 3, 5}, 8, []int{2, 3}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.nums, testCase.target); !reflect.DeepEqual(got, testCase.want) {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
