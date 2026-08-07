package solution

import (
	"strconv"
	"strings"
)

func Solve(input string) string {
	fields := strings.Fields(input)
	maximum, _ := strconv.Atoi(fields[0])
	for _, field := range fields[1:] {
		value, _ := strconv.Atoi(field)
		if value > maximum {
			maximum = value
		}
	}
	return strconv.Itoa(maximum)
}
