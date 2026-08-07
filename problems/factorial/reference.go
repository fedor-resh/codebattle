package solution

import (
	"strconv"
	"strings"
)

func Solve(input string) string {
	n, _ := strconv.Atoi(strings.TrimSpace(input))
	var result uint64 = 1
	for i := 2; i <= n; i++ {
		result *= uint64(i)
	}
	return strconv.FormatUint(result, 10)
}
