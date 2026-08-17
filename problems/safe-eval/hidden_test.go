package solution

import (
	"errors"
	"testing"
)

func TestHidden(t *testing.T) {
	validCases := []struct {
		expr string
		want int
	}{
		{"7-10", -3},
		{"  -20 / -6  ", 3},
		{"0 * 999", 0},
		{"+15 + -4", 11},
	}
	for index, testCase := range validCases {
		got, err := Eval(testCase.expr)
		if err != nil || got != testCase.want {
			t.Fatalf("valid case %d failed", index+1)
		}
	}

	errorCases := []struct {
		expr string
		want error
	}{
		{"", ErrInvalid},
		{"1 + 2 + 3", ErrInvalid},
		{"word + 1", ErrInvalid},
		{"42 / +0", ErrDivZero},
	}
	for index, testCase := range errorCases {
		_, err := Eval(testCase.expr)
		if !errors.Is(err, testCase.want) {
			t.Fatalf("error case %d failed", index+1)
		}
	}
}
