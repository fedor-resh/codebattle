package solution

import (
	"strings"
	"testing"
)

func TestHidden(t *testing.T) {
	for attempt := range 20 {
		for _, n := range []int{-3, 1, 2, 8} {
			if got, want := Solve(n), strings.Repeat("pingpong", max(n, 0)); got != want {
				t.Fatalf("attempt %d, n=%d: got %q, want %q", attempt, n, got, want)
			}
		}
	}
}
