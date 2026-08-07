package solution

import (
	"strconv"
	"strings"
)

func Solve(input string) string {
	parts := strings.SplitN(input, "|", 2)
	target, _ := strconv.Atoi(parts[0])
	seen := map[int]int{}
	for index, raw := range strings.Split(parts[1], ",") {
		value, _ := strconv.Atoi(raw)
		if other, ok := seen[target-value]; ok {
			return strconv.Itoa(other) + "," + strconv.Itoa(index)
		}
		seen[value] = index
	}
	return ""
}
