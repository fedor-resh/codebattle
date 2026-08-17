package solution

import (
	"errors"
	"strconv"
	"strings"
)

var ErrEmpty = errors.New("empty stack")

type Stack struct {
	items []int
}

func (stack *Stack) Push(value int) {
	// Напишите решение
}

func (stack *Stack) Pop() (int, error) {
	// Напишите решение
	return 0, nil
}

// Solve адаптирует методы Stack к JSON-совместимому контракту judge.
func Solve(commands []string) []string {
	stack := &Stack{}
	result := make([]string, 0, len(commands))
	for _, command := range commands {
		parts := strings.Fields(command)
		switch {
		case len(parts) == 2 && parts[0] == "push":
			if value, err := strconv.Atoi(parts[1]); err == nil {
				stack.Push(value)
			}
			result = append(result, "ok")
		case len(parts) == 1 && parts[0] == "pop":
			value, err := stack.Pop()
			if err != nil {
				result = append(result, err.Error())
				continue
			}
			result = append(result, strconv.Itoa(value))
		default:
			result = append(result, "ok")
		}
	}
	return result
}
