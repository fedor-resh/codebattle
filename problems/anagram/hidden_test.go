package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		first  string
		second string
		want   bool
	}{
		{"🙂a", "a🙂", true},
		{"Aa", "aa", false},
		{"", "", true},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.first, testCase.second); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
