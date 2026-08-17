package solution

import "strconv"

type ValidationError struct {
	Index int
	Value int
}

func (err *ValidationError) Error() string {
	if err == nil {
		return "<nil>"
	}
	return "invalid value at index " + strconv.Itoa(err.Index)
}

// Validate проверяет, что все числа неотрицательны.
func Validate(values []int) error {
	var problem *ValidationError
	for index, value := range values {
		if value < 0 {
			problem = &ValidationError{Index: index, Value: value}
			break
		}
	}
	return problem
}

// Solve адаптирует Validate к JSON-совместимому контракту judge.
func Solve(values []int) string {
	if err := Validate(values); err != nil {
		return err.Error()
	}
	return "ok"
}
