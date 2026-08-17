package solution

import (
	"slices"
	"sync"
	"testing"
	"time"
)

func TestHidden(t *testing.T) {
	functionalCases := []struct {
		keys []int
		want []int
	}{
		{nil, []int{}},
		{[]int{9}, []int{90}},
		{[]int{3, -1, 3, 0, -1, 3}, []int{30, -10, 30, 0, -10, 30}},
	}
	for index, testCase := range functionalCases {
		if got := Solve(testCase.keys, func(key int) int { return key * 10 }); !slices.Equal(got, testCase.want) {
			t.Fatalf("functional case %d failed", index+1)
		}
	}

	keys := []int{4, 2, 4, 1, 2, 4, 1}
	calls := make(map[int]int)
	var callsMu sync.Mutex
	got := Solve(keys, func(key int) int {
		callsMu.Lock()
		calls[key]++
		callsMu.Unlock()
		return key * key
	})
	if !slices.Equal(got, []int{16, 4, 16, 1, 4, 16, 1}) {
		t.Fatal("duplicate-key result was incorrect")
	}
	for _, key := range []int{1, 2, 4} {
		if calls[key] != 1 {
			t.Fatal("work was called more than once for a key")
		}
	}

	distinct := []int{1, 2, 3, 4, 5, 6}
	done := make(chan []int, 1)
	go func() {
		done <- Solve(distinct, func(key int) int {
			time.Sleep(120 * time.Millisecond)
			return key + 10
		})
	}()
	select {
	case timedResult := <-done:
		if !slices.Equal(timedResult, []int{11, 12, 13, 14, 15, 16}) {
			t.Fatal("parallel case returned an incorrect result")
		}
	case <-time.After(450 * time.Millisecond):
		t.Fatal("work was not performed in parallel")
	}
}
