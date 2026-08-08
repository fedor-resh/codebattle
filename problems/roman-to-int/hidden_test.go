package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		roman string
		want  int
	}{
		{"IV", 4},
		{"LVIII", 58},
		{"MMMCMXCIX", 3999},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.roman); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
