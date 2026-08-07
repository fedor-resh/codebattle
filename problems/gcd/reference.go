package solution

import (
	"strconv"
	"strings"
)

func Solve(input string) string {
	parts := strings.SplitN(input, ",", 2)
	a, _ := strconv.Atoi(parts[0])
	b, _ := strconv.Atoi(parts[1])
	for b != 0 {
		a, b = b, a%b
	}
	return strconv.Itoa(a)
}
