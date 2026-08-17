package solution

import (
	"errors"
	"slices"
	"testing"
)

func TestHidden(t *testing.T) {
	stack := &Stack{}
	value, err := stack.Pop()
	if value != 0 || !errors.Is(err, ErrEmpty) {
		t.Fatal("empty pop must return zero and ErrEmpty")
	}

	stack.Push(4)
	stack.Push(9)
	value, err = stack.Pop()
	if err != nil || value != 9 {
		t.Fatal("last pushed value was not popped")
	}
	value, err = stack.Pop()
	if err != nil || value != 4 {
		t.Fatal("stack is not LIFO")
	}
	if _, err = stack.Pop(); !errors.Is(err, ErrEmpty) {
		t.Fatal("emptied stack must return ErrEmpty")
	}

	cases := []struct {
		commands []string
		want     []string
	}{
		{nil, []string{}},
		{[]string{"push x", "pop", " push  3 ", "pop"}, []string{"ok", "empty stack", "ok", "3"}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.commands); !slices.Equal(got, testCase.want) {
			t.Fatalf("solve case %d failed", index+1)
		}
	}
}
