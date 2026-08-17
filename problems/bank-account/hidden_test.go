package solution

import (
	"slices"
	"testing"
)

func TestHidden(t *testing.T) {
	account := &Account{}
	if account.Balance() != 0 {
		t.Fatal("zero-value account has a non-zero balance")
	}
	account.Deposit(50)
	account.Deposit(25)
	account.Deposit(0)
	account.Deposit(-10)
	if account.Balance() != 75 {
		t.Fatal("deposit rules were not applied")
	}
	if account.Withdraw(-1) || account.Withdraw(0) || account.Withdraw(76) {
		t.Fatal("invalid withdrawal succeeded")
	}
	if account.Balance() != 75 {
		t.Fatal("failed withdrawal changed the balance")
	}
	if !account.Withdraw(75) || account.Balance() != 0 {
		t.Fatal("valid withdrawal failed")
	}

	cases := []struct {
		commands []string
		want     []int
	}{
		{nil, []int{}},
		{[]string{"deposit 10", "bad command", "withdraw 3"}, []int{10, 10, 7}},
		{[]string{" deposit   7 ", "withdraw nope", "withdraw 7"}, []int{7, 7, 0}},
	}
	for index, testCase := range cases {
		if got := Solve(testCase.commands); !slices.Equal(got, testCase.want) {
			t.Fatalf("solve case %d failed", index+1)
		}
	}
}
