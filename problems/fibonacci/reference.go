package solution

import (
	"strconv"
	"strings"
)

func Solve(input string) string {
	n, _ := strconv.Atoi(strings.TrimSpace(input))
	var a, b uint64 = 0, 1
	for i := 0; i < n; i++ {
		a, b = b, a+b
	}
	return strconv.FormatUint(a, 10)
}
