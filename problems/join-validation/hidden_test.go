package solution

import (
	"errors"
	"slices"
	"testing"
)

func TestHidden(t *testing.T) {
	if err := Validate(nil); err != nil {
		t.Fatal("empty input must be valid")
	}
	if err := Validate([]string{"a", "b", "abcdefgh"}); err != nil {
		t.Fatal("valid fields must return nil")
	}

	err := Validate([]string{"", "toolongname", "x", "x"})
	if !errors.Is(err, ErrEmpty) || !errors.Is(err, ErrTooLong) || !errors.Is(err, ErrDuplicate) {
		t.Fatal("joined error must expose every sentinel")
	}

	var fieldErr *FieldError
	if !errors.As(err, &fieldErr) || fieldErr == nil || fieldErr.Field != "" || !errors.Is(fieldErr, ErrEmpty) {
		t.Fatal("errors.As must return the first FieldError")
	}

	longDup := Validate([]string{"123456789", "123456789"})
	if !errors.Is(longDup, ErrTooLong) || !errors.Is(longDup, ErrDuplicate) {
		t.Fatal("a field can be both too long and duplicate")
	}

	solveCases := []struct {
		fields []string
		want   []string
	}{
		{nil, []string{}},
		{[]string{"ok"}, []string{}},
		{[]string{"", ""}, []string{":duplicate", ":empty", ":empty"}},
	}
	for index, testCase := range solveCases {
		if got := Solve(testCase.fields); !slices.Equal(got, testCase.want) {
			t.Fatalf("solve case %d failed", index+1)
		}
	}
}
