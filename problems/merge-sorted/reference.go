package solution

import (
	"strconv"
	"strings"
)

func Solve(input string) string {
	parts := strings.SplitN(input, "|", 2)
	left, right := splitNumbers(parts[0]), splitNumbers(parts[1])
	result := make([]string, 0, len(left)+len(right))
	i, j := 0, 0
	for i < len(left) && j < len(right) {
		a, _ := strconv.Atoi(left[i])
		b, _ := strconv.Atoi(right[j])
		if a <= b {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}
	result = append(result, left[i:]...)
	result = append(result, right[j:]...)
	return strings.Join(result, ",")
}

func splitNumbers(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}
