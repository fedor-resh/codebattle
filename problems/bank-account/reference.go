package solution

import (
	"strconv"
	"strings"
)

type Account struct {
	balance int
}

func (account *Account) Deposit(amount int) {
	if amount > 0 {
		account.balance += amount
	}
}

func (account *Account) Withdraw(amount int) bool {
	if amount <= 0 || amount > account.balance {
		return false
	}
	account.balance -= amount
	return true
}

func (account *Account) Balance() int {
	return account.balance
}

func Solve(commands []string) []int {
	account := &Account{}
	result := make([]int, 0, len(commands))
	for _, command := range commands {
		parts := strings.Fields(command)
		if len(parts) == 2 {
			amount, err := strconv.Atoi(parts[1])
			if err == nil {
				switch parts[0] {
				case "deposit":
					account.Deposit(amount)
				case "withdraw":
					account.Withdraw(amount)
				}
			}
		}
		result = append(result, account.Balance())
	}
	return result
}
