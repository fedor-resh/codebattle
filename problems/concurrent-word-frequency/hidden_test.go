package solution

import (
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		words   []string
		workers int
		want    map[string]int
	}{
		{nil, -1, map[string]int{}},
		{[]string{"one"}, 20, map[string]int{"one": 1}},
		{[]string{"x", "y", "x", "z", "x", "y"}, 1, map[string]int{"x": 3, "y": 2, "z": 1}},
		{[]string{"", "go", "", "Go"}, 4, map[string]int{"": 2, "go": 1, "Go": 1}},
	}
	for attempt := range 20 {
		for _, testCase := range cases {
			if got := Solve(testCase.words, testCase.workers, func(word string) string { return word }); !reflect.DeepEqual(got, testCase.want) {
				t.Fatalf("attempt %d: got %v, want %v", attempt, got, testCase.want)
			}
		}
	}

	words := []string{"a", "b", "a", "c", "a", "b", "d", "a"}
	want := map[string]int{"a": 4, "b": 2, "c": 1, "d": 1}
	var calls atomic.Int32
	done := make(chan map[string]int, 1)
	go func() {
		done <- Solve(words, 4, func(word string) string {
			calls.Add(1)
			time.Sleep(120 * time.Millisecond)
			return word
		})
	}()
	select {
	case got := <-done:
		if !reflect.DeepEqual(got, want) || calls.Load() != int32(len(words)) {
			t.Fatalf("timed case: got %v with %d work calls", got, calls.Load())
		}
	case <-time.After(650 * time.Millisecond):
		t.Fatal("work was not performed in parallel")
	}
}
