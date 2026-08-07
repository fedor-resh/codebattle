package solution

import (
	"strconv"
	"strings"
)

func Solve(input string) string {
	sum := 0
	for _, field := range strings.Fields(input) {
		value, _ := strconv.Atoi(field)
		sum += value
	}
	return strconv.Itoa(sum)
}
