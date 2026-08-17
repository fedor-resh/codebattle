package solution

import (
	"slices"
	"testing"
)

func TestHidden(t *testing.T) {
	counter := &Counter{}
	if counter.Value() != 0 {
		t.Fatal("zero-value counter is not zero")
	}
	counter.Inc()
	counter.Inc()
	counter.Add(4)
	counter.Add(-1)
	if counter.Value() != 5 {
		t.Fatal("counter methods were not applied")
	}

	cases := []struct {
		commands []string
		want     []int
	}{
		{nil, []int{}},
		{[]string{"inc", "noop", "add x", "add 2"}, []int{1, 1, 1, 3}},
		{[]string{" add   7 ", "inc"}, []int{7, 8}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.commands); !slices.Equal(got, testCase.want) {
			t.Fatalf("solve case %d failed", index+1)
		}
	}
}
