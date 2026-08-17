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

func Validate(values []int) error {
	for index, value := range values {
		if value < 0 {
			return &ValidationError{Index: index, Value: value}
		}
	}
	return nil
}

func Solve(values []int) string {
	if err := Validate(values); err != nil {
		return err.Error()
	}
	return "ok"
}
