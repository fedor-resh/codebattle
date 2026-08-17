package solution

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestHidden(t *testing.T) {
	value, err := Lookup(map[string]int{"Ann": 3}, "Ann")
	if err != nil || value != 3 {
		t.Fatal("existing key must return the score")
	}

	value, err = Lookup(nil, "Ada")
	if value != 0 || !errors.Is(err, ErrNotFound) {
		t.Fatal("missing key must wrap ErrNotFound")
	}
	if err == nil || !strings.Contains(err.Error(), "Ada") {
		t.Fatal("wrapped error must mention the player name")
	}

	cases := []struct {
		scores map[string]int
		names  []string
		want   []string
	}{
		{nil, nil, []string{}},
		{map[string]int{"Zed": -4}, []string{"Zed", "missing"}, []string{"-4", "player missing: not found"}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.scores, testCase.names); !slices.Equal(got, testCase.want) {
			t.Fatalf("solve case %d failed", index+1)
		}
	}
}
