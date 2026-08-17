package solution

import (
	"errors"
	"reflect"
	"testing"
)

func TestHidden(t *testing.T) {
	validCases := []struct {
		text string
		want map[string]int
	}{
		{"", map[string]int{}},
		{"answer = -42\nzero=0", map[string]int{"answer": -42, "zero": 0}},
		{" # comment\r\nsize = +12\r\n", map[string]int{"size": 12}},
	}
	for index, testCase := range validCases {
		got, err := Parse(testCase.text)
		if err != nil || !reflect.DeepEqual(got, testCase.want) {
			t.Fatalf("valid case %d failed", index+1)
		}
	}

	errorCases := []struct {
		text string
		want error
	}{
		{"missing", ErrBadLine},
		{"=10", ErrBadLine},
		{"a=1=2", ErrBadLine},
		{"a=1\na=2", ErrDuplicate},
		{"a=1.5", ErrInvalidValue},
	}
	for index, testCase := range errorCases {
		_, err := Parse(testCase.text)
		if !errors.Is(err, testCase.want) {
			t.Fatalf("error case %d failed", index+1)
		}
	}
}
