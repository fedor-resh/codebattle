package solution

import (
	"errors"
	"strconv"
)

var ErrPanic = errors.New("panic recovered")

// SafeDivide делит нацело и перехватывает панику деления на ноль.
func SafeDivide(a, b int) (result int, err error) {
	// Напишите решение
	return a / b, nil
}

// Solve адаптирует SafeDivide к JSON-совместимому контракту judge.
func Solve(nums []int, divisors []int) []string {
	count := len(nums)
	if len(divisors) < count {
		count = len(divisors)
	}
	result := make([]string, 0, count)
	for index := 0; index < count; index++ {
		value, err := SafeDivide(nums[index], divisors[index])
		if err != nil {
			result = append(result, err.Error())
			continue
		}
		result = append(result, strconv.Itoa(value))
	}
	return result
}
