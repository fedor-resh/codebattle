package solution

import (
	"reflect"
	"testing"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		nums []int
		k    int
		want []int
	}{
		{[]int{1, 2}, 0, []int{1, 2}},
		{[]int{1, 2, 3}, 5, []int{2, 3, 1}},
		{[]int{9}, 3, []int{9}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.nums, testCase.k); !reflect.DeepEqual(got, testCase.want) {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
