package solution

import (
	"slices"
	"testing"
)

var (
	_ Filter = Add{}
	_ Filter = Mul{}
	_ Filter = Clamp{}
)

func TestHidden(t *testing.T) {
	methodCases := []struct {
		filter Filter
		input  int
		want   int
	}{
		{Add{Value: -3}, 10, 7},
		{Mul{Value: -2}, 4, -8},
		{Clamp{Min: -5, Max: 5}, -10, -5},
		{Clamp{Min: -5, Max: 5}, 2, 2},
		{Clamp{Min: -5, Max: 5}, 12, 5},
	}
	for index, testCase := range methodCases {
		if testCase.filter.Apply(testCase.input) != testCase.want {
			t.Fatalf("method case %d failed", index+1)
		}
	}

	solveCases := []struct {
		nums []int
		ops  []string
		want []int
	}{
		{nil, nil, []int{}},
		{[]int{1, 2}, []string{"unknown 5", "add nope", "add 3"}, []int{4, 5}},
		{[]int{-10, 0, 10}, []string{"add 2", "clamp -3 7", "mul 2"}, []int{-6, 4, 14}},
	}
	for index, testCase := range solveCases {
		if got := Solve(testCase.nums, testCase.ops); !slices.Equal(got, testCase.want) {
			t.Fatalf("solve case %d failed", index+1)
		}
	}
}
