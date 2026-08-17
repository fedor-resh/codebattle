package solution

import (
	"slices"
	"testing"
)

var (
	_ Greeter = Base{}
	_ Greeter = Shouty{}
	_ Greeter = Polite{}
)

func TestHidden(t *testing.T) {
	base := Base{Name: "Ann"}
	if base.Greet() != "Hello, Ann" {
		t.Fatal("base greeting is wrong")
	}

	shouty := Shouty{Base: Base{Name: "Ann"}}
	if shouty.Greet() != "HEY, ANN!" {
		t.Fatal("shouty greeting is wrong")
	}
	if shouty.Base.Greet() != "Hello, Ann" {
		t.Fatal("embedded Base.Greet must keep the original implementation")
	}

	polite := Polite{Base: Base{Name: "Ann"}}
	if polite.Greet() != "Please, Hello, Ann" {
		t.Fatal("polite greeting must reuse Base.Greet")
	}

	solveCases := []struct {
		kinds []string
		names []string
		want  []string
	}{
		{nil, nil, []string{}},
		{[]string{"base", "shouty"}, []string{"Ира"}, []string{"Hello, Ира"}},
		{[]string{"polite", "shouty", "base"}, []string{"Ada", "bob", "Eve"}, []string{"Please, Hello, Ada", "HEY, BOB!", "Hello, Eve"}},
	}
	for index, testCase := range solveCases {
		if got := Solve(testCase.kinds, testCase.names); !slices.Equal(got, testCase.want) {
			t.Fatalf("solve case %d failed", index+1)
		}
	}
}
