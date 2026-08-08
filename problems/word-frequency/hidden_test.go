package solution

import (
	"reflect"
	"testing"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		words []string
		want  map[string]int
	}{
		{[]string{}, map[string]int{}},
		{[]string{"x"}, map[string]int{"x": 1}},
		{[]string{"A", "a", "A"}, map[string]int{"A": 2, "a": 1}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.words); !reflect.DeepEqual(got, testCase.want) {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
