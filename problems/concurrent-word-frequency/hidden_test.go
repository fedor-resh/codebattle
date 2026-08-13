package solution

import (
	"reflect"
	"testing"
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
			if got := Solve(testCase.words, testCase.workers); !reflect.DeepEqual(got, testCase.want) {
				t.Fatalf("attempt %d: got %v, want %v", attempt, got, testCase.want)
			}
		}
	}
}
