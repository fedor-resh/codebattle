package solution

import (
	"errors"
	"slices"
	"testing"
)

func TestHidden(t *testing.T) {
	defer func() {
		if recover() != nil {
			t.Fatal("panic leaked from SafeDivide")
		}
	}()

	value, err := SafeDivide(9, 4)
	if err != nil || value != 2 {
		t.Fatal("successful division failed")
	}

	value, err = SafeDivide(-10, 3)
	if err != nil || value != -3 {
		t.Fatal("negative division failed")
	}

	value, err = SafeDivide(8, 0)
	if value != 0 || !errors.Is(err, ErrPanic) {
		t.Fatal("division by zero must return ErrPanic and a zero result")
	}

	value, err = SafeDivide(0, 0)
	if value != 0 || !errors.Is(err, ErrPanic) {
		t.Fatal("0/0 must be recovered")
	}

	solveCases := []struct {
		nums     []int
		divisors []int
		want     []string
	}{
		{nil, nil, []string{}},
		{[]int{15, 4}, []int{5}, []string{"3"}},
		{[]int{1, -8, 0}, []int{0, -3, 2}, []string{"panic recovered", "2", "0"}},
	}
	for index, testCase := range solveCases {
		if got := Solve(testCase.nums, testCase.divisors); !slices.Equal(got, testCase.want) {
			t.Fatalf("solve case %d failed", index+1)
		}
	}
}
