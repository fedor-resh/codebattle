package solution

import (
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		nums []int
		want []int
	}{
		{nil, []int{}},
		{[]int{7}, []int{64}},
		{[]int{-3, 0, 4, 10}, []int{4, 1, 25, 121}},
	}
	for index, testCase := range cases {
		got := Solve(testCase.nums, func(value int) int { return value + 1 }, func(value int) int { return value * value })
		if !slices.Equal(got, testCase.want) {
			t.Fatalf("functional case %d failed", index+1)
		}
	}

	values := []int{1, 2, 3, 4, 5, 6}
	want := []int{4, 6, 8, 10, 12, 14}
	var callsA atomic.Int32
	var callsB atomic.Int32
	done := make(chan []int, 1)
	go func() {
		done <- Solve(values, func(value int) int {
			callsA.Add(1)
			time.Sleep(100 * time.Millisecond)
			return value + 1
		}, func(value int) int {
			callsB.Add(1)
			time.Sleep(100 * time.Millisecond)
			return value * 2
		})
	}()
	select {
	case got := <-done:
		if !slices.Equal(got, want) || callsA.Load() != int32(len(values)) || callsB.Load() != int32(len(values)) {
			t.Fatal("pipeline produced incorrect results or call counts")
		}
	case <-time.After(950 * time.Millisecond):
		t.Fatal("work was not performed in parallel")
	}
}
