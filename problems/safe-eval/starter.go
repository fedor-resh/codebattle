package solution

import (
	"errors"
	"strconv"
)

var (
	ErrInvalid = errors.New("invalid expression")
	ErrDivZero = errors.New("division by zero")
)

// Eval вычисляет выражение из двух целых чисел и одного оператора.
func Eval(expr string) (int, error) {
	// Напишите решение
	return 0, ErrInvalid
}

// Solve адаптирует результат Eval к JSON-совместимому контракту judge.
func Solve(expr string) string {
	value, err := Eval(expr)
	if err != nil {
		return err.Error()
	}
	return strconv.Itoa(value)
}
