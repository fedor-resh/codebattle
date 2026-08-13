package solution

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func testWork(ctx context.Context, delay int) bool {
	timer := time.NewTimer(time.Duration(delay) * time.Millisecond)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func TestHidden(t *testing.T) {
	cases := []struct {
		delays  []int
		timeout int
		want    int
	}{
		{nil, 10, -1},
		{[]int{80, 50, 1, 30}, 250, 2},
		{[]int{25, 15, 35}, 3, -1},
		{[]int{12}, 100, 0},
		{[]int{8, 2, 5}, -1, -1},
	}
	for attempt := range 10 {
		for _, testCase := range cases {
			if got := Solve(testCase.delays, testCase.timeout, testWork); got != testCase.want {
				t.Fatalf("attempt %d: got %d, want %d", attempt, got, testCase.want)
			}
		}
	}

	var calls atomic.Int32
	var canceled atomic.Int32
	done := make(chan int, 1)
	go func() {
		done <- Solve([]int{700, 500, 80, 350}, 1000, func(ctx context.Context, delay int) bool {
			calls.Add(1)
			timer := time.NewTimer(time.Duration(delay) * time.Millisecond)
			defer timer.Stop()
			select {
			case <-timer.C:
				return true
			case <-ctx.Done():
				canceled.Add(1)
				return false
			}
		})
	}()
	select {
	case got := <-done:
		if got != 2 || calls.Load() != 4 || canceled.Load() != 3 {
			t.Fatalf("timed case: got %d, calls %d, canceled %d", got, calls.Load(), canceled.Load())
		}
	case <-time.After(350 * time.Millisecond):
		t.Fatal("handlers were not started in parallel")
	}
}
