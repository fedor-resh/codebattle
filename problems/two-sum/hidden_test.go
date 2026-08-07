package solution

import "testing"

func TestHidden(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"0|-1,1", "0,1"},
		{"10|5,5", "0,1"},
		{"8|1,2,3,5", "2,3"},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.input); got != testCase.want {
			t.Fatalf("case %d failed", index+1)
		}
	}
}
