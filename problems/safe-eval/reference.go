package solution

import (
	"errors"
	"regexp"
	"strconv"
)

var (
	ErrInvalid = errors.New("invalid expression")
	ErrDivZero = errors.New("division by zero")
)

var expressionPattern = regexp.MustCompile(`^[[:space:]]*([+-]?[0-9]+)[[:space:]]*([-+*/])[[:space:]]*([+-]?[0-9]+)[[:space:]]*$`)

func Eval(expr string) (int, error) {
	parts := expressionPattern.FindStringSubmatch(expr)
	if parts == nil {
		return 0, ErrInvalid
	}
	left, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, ErrInvalid
	}
	right, err := strconv.Atoi(parts[3])
	if err != nil {
		return 0, ErrInvalid
	}

	switch parts[2] {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0 {
			return 0, ErrDivZero
		}
		return left / right, nil
	default:
		return 0, ErrInvalid
	}
}

func Solve(expr string) string {
	value, err := Eval(expr)
	if err != nil {
		return err.Error()
	}
	return strconv.Itoa(value)
}
