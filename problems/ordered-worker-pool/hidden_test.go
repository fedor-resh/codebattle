package solution

import (
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		nums    []int
		workers int
		want    []int
	}{
		{nil, -1, []int{}},
		{[]int{7}, 12, []int{49}},
		{[]int{9, -8, 7, -6, 5, -4}, 1, []int{81, 64, 49, 36, 25, 16}},
		{[]int{1, 10, 2, 9, 3, 8}, 4, []int{1, 100, 4, 81, 9, 64}},
	}
	for attempt := range 20 {
		for _, testCase := range cases {
			if got := Solve(testCase.nums, testCase.workers, func(value int) int { return value * value }); !slices.Equal(got, testCase.want) {
				t.Fatalf("attempt %d: got %v, want %v", attempt, got, testCase.want)
			}
		}
	}

	values := []int{1, 2, 3, 4, 5, 6, 7, 8}
	want := []int{1, 4, 9, 16, 25, 36, 49, 64}
	var calls atomic.Int32
	done := make(chan []int, 1)
	go func() {
		done <- Solve(values, 4, func(value int) int {
			calls.Add(1)
			time.Sleep(120 * time.Millisecond)
			return value * value
		})
	}()
	select {
	case got := <-done:
		if !slices.Equal(got, want) || calls.Load() != int32(len(values)) {
			t.Fatalf("timed case: got %v with %d work calls", got, calls.Load())
		}
	case <-time.After(650 * time.Millisecond):
		t.Fatal("work was not performed in parallel")
	}
}
