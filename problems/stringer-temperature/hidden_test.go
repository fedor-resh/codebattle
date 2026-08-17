package solution

import (
	"fmt"
	"slices"
	"testing"
)

func TestHidden(t *testing.T) {
	var _ fmt.Stringer = Celsius(0)

	if got := Celsius(0).String(); got != "0°C" {
		t.Fatal("zero value must format as 0°C")
	}
	if got := fmt.Sprint(Celsius(-12)); got != "-12°C" {
		t.Fatal("fmt.Sprint must use String()")
	}

	cases := []struct {
		values []int
		want   []string
	}{
		{nil, []string{}},
		{[]int{1, 2, 3}, []string{"1°C", "2°C", "3°C"}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.values); !slices.Equal(got, testCase.want) {
			t.Fatalf("solve case %d failed", index+1)
		}
	}
}
