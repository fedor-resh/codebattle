package solution

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestHidden(t *testing.T) {
	for attempt := range 20 {
		for _, n := range []int{-3, 1, 2, 8} {
			if got, want := Solve(n, func(value string) string { return value }), strings.Repeat("pingpong", max(n, 0)); got != want {
				t.Fatalf("attempt %d, n=%d: got %q, want %q", attempt, n, got, want)
			}
		}
	}

	var calls atomic.Int32
	done := make(chan string, 1)
	go func() {
		done <- Solve(4, func(value string) string {
			calls.Add(1)
			time.Sleep(120 * time.Millisecond)
			return value
		})
	}()
	select {
	case got := <-done:
		if want := strings.Repeat("pingpong", 4); got != want || calls.Load() != 8 {
			t.Fatalf("timed case: got %q with %d work calls", got, calls.Load())
		}
	case <-time.After(650 * time.Millisecond):
		t.Fatal("ping and pong were not processed in parallel")
	}
}
