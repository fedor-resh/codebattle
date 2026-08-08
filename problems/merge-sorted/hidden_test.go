package solution

import (
	"reflect"
	"testing"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		left  []int
		right []int
		want  []int
	}{
		{[]int{}, []int{1, 2}, []int{1, 2}},
		{[]int{-3, 0}, []int{0, 4}, []int{-3, 0, 0, 4}},
		{[]int{1, 2}, []int{}, []int{1, 2}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.left, testCase.right); !reflect.DeepEqual(got, testCase.want) {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
