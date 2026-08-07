package solution

import (
	"strconv"
	"strings"
)

func Solve(input string) string {
	n, _ := strconv.Atoi(strings.TrimSpace(input))
	result := make([]string, 0, n)
	for i := 1; i <= n; i++ {
		switch {
		case i%15 == 0:
			result = append(result, "FizzBuzz")
		case i%3 == 0:
			result = append(result, "Fizz")
		case i%5 == 0:
			result = append(result, "Buzz")
		default:
			result = append(result, strconv.Itoa(i))
		}
	}
	return strings.Join(result, ",")
}
