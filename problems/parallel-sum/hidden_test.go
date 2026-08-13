package solution

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		nums    []int
		workers int
		want    int
	}{
		{nil, -2, 0},
		{[]int{42}, 20, 42},
		{[]int{-10, -20, 5, 7, 18}, 1, 0},
		{[]int{1, 2, 3, 4, 5, 6, 7, 8, 9}, 4, 45},
	}
	for attempt := range 20 {
		for _, testCase := range cases {
			if got := Solve(testCase.nums, testCase.workers, func(value int) int { return value }); got != testCase.want {
				t.Fatalf("attempt %d: got %d, want %d", attempt, got, testCase.want)
			}
		}
	}

	values := []int{1, 2, 3, 4, 5, 6, 7, 8}
	var calls atomic.Int32
	done := make(chan int, 1)
	go func() {
		done <- Solve(values, 4, func(value int) int {
			calls.Add(1)
			time.Sleep(120 * time.Millisecond)
			return value
		})
	}()
	select {
	case got := <-done:
		if got != 36 || calls.Load() != int32(len(values)) {
			t.Fatalf("timed case: got %d with %d work calls", got, calls.Load())
		}
	case <-time.After(650 * time.Millisecond):
		t.Fatal("work was not performed in parallel")
	}
}
