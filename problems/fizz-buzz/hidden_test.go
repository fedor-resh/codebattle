package solution

import (
	"reflect"
	"testing"
)

func TestHidden(t *testing.T) {
	cases := []struct {
		n    int
		want []string
	}{
		{3, []string{"1", "2", "Fizz"}},
		{15, []string{"1", "2", "Fizz", "4", "Buzz", "Fizz", "7", "8", "Fizz", "Buzz", "11", "Fizz", "13", "14", "FizzBuzz"}},
		{0, []string{}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.n); !reflect.DeepEqual(got, testCase.want) {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
