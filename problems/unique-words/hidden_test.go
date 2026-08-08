package solution

import (
	"reflect"
	"testing"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		words []string
		want  []string
	}{
		{[]string{}, []string{}},
		{[]string{"one"}, []string{"one"}},
		{[]string{"A", "a", "A"}, []string{"A", "a"}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.words); !reflect.DeepEqual(got, testCase.want) {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
