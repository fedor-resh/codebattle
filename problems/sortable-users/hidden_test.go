package solution

import (
	"slices"
	"sort"
	"testing"
)

var _ sort.Interface = ByScore{}

func TestHidden(t *testing.T) {
	users := ByScore{
		{Name: "bob", Score: 1},
		{Name: "ann", Score: 3},
		{Name: "cara", Score: 3},
	}
	if users.Len() != 3 {
		t.Fatal("Len is wrong")
	}
	if !users.Less(1, 0) || users.Less(0, 1) {
		t.Fatal("higher score must come first")
	}
	if !users.Less(1, 2) || users.Less(2, 1) {
		t.Fatal("equal scores must break ties by name")
	}

	same := ByScore{{Name: "ann", Score: 2}, {Name: "ann", Score: 2}}
	if same.Less(0, 1) || same.Less(1, 0) {
		t.Fatal("equal items must not be less than each other")
	}

	users.Swap(0, 2)
	if users[0].Name != "cara" || users[2].Name != "bob" {
		t.Fatal("Swap did not exchange items")
	}

	sorted := ByScore{
		{Name: "b", Score: -1},
		{Name: "a", Score: -1},
		{Name: "c", Score: 0},
	}
	sort.Sort(sorted)
	if sorted[0].Name != "c" || sorted[1].Name != "a" || sorted[2].Name != "b" {
		t.Fatal("sort.Sort did not apply Less and Swap")
	}

	solveCases := []struct {
		names  []string
		scores []int
		want   []string
	}{
		{nil, nil, []string{}},
		{[]string{"b", "a"}, []int{0}, []string{"b:0"}},
		{[]string{" ken", "Ada", "ada"}, []int{1, 1, 2}, []string{"ada:2", " ken:1", "Ada:1"}},
	}
	for index, testCase := range solveCases {
		if got := Solve(testCase.names, testCase.scores); !slices.Equal(got, testCase.want) {
			t.Fatalf("solve case %d failed", index+1)
		}
	}
}
