package solution

import (
	"errors"
	"slices"
	"testing"
)

var _ Sink = (*LimitedSink)(nil)

func TestHidden(t *testing.T) {
	sink := &LimitedSink{Limit: 5}
	written, err := sink.Write("abc")
	if err != nil || written != 3 {
		t.Fatal("first write should succeed")
	}

	written, err = sink.Write("xyz")
	if written != 0 || !errors.Is(err, ErrFull) {
		t.Fatal("oversized write must return ErrFull")
	}

	written, err = sink.Write("de")
	if err != nil || written != 2 {
		t.Fatal("rejected write must not consume capacity")
	}

	written, err = sink.Write("")
	if err != nil || written != 0 {
		t.Fatal("empty write must succeed")
	}

	if _, err = sink.Write("!"); !errors.Is(err, ErrFull) {
		t.Fatal("full sink must keep rejecting non-empty chunks")
	}

	zero := &LimitedSink{}
	if _, err = zero.Write("x"); !errors.Is(err, ErrFull) {
		t.Fatal("zero-value sink must reject non-empty writes")
	}

	solveCases := []struct {
		limit  int
		chunks []string
		want   []string
	}{
		{0, nil, []string{}},
		{3, []string{"abc", "d"}, []string{"3", "sink is full"}},
		{6, []string{"й", "я"}, []string{"2", "2"}},
	}
	for index, testCase := range solveCases {
		if got := Solve(testCase.limit, testCase.chunks); !slices.Equal(got, testCase.want) {
			t.Fatalf("solve case %d failed", index+1)
		}
	}
}
