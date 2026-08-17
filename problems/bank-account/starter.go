package solution

import (
	"strconv"
	"strings"
)

type Account struct {
	balance int
}

func (account *Account) Deposit(amount int) {
	// Напишите решение
}

func (account *Account) Withdraw(amount int) bool {
	// Напишите решение
	return false
}

func (account *Account) Balance() int {
	// Напишите решение
	return 0
}

// Solve адаптирует методы Account к JSON-совместимому контракту judge.
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
