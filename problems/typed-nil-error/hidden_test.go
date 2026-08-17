package solution

import (
	"errors"
	"testing"
)

func TestHidden(t *testing.T) {
	if err := Validate(nil); err != nil {
		t.Fatal("empty input must return a nil error interface")
	}
	if err := Validate([]int{0, 4, 9}); err != nil {
		t.Fatal("valid input must return a nil error interface")
	}

	err := Validate([]int{1, -5, -1})
	var problem *ValidationError
	if !errors.As(err, &problem) || problem == nil {
		t.Fatal("invalid input must return *ValidationError")
	}
	if problem.Index != 1 || problem.Value != -5 {
		t.Fatal("ValidationError fields are wrong")
	}
	if err.Error() != "invalid value at index 1" {
		t.Fatal("Error() text is wrong")
	}

	if got := Solve([]int{-3}); got != "invalid value at index 0" {
		t.Fatal("solve failed for the first negative value")
	}
	if got := Solve([]int{8, 0}); got != "ok" {
		t.Fatal("solve failed for a valid slice")
	}
}
