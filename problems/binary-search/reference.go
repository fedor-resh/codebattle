package solution

import (
	"strconv"
	"strings"
)

func Solve(input string) string {
	parts := strings.SplitN(input, "|", 2)
	target, _ := strconv.Atoi(parts[0])
	items := strings.Split(parts[1], ",")
	left, right := 0, len(items)-1
	for left <= right {
		middle := left + (right-left)/2
		value, _ := strconv.Atoi(items[middle])
		if value == target {
			return strconv.Itoa(middle)
		}
		if value < target {
			left = middle + 1
		} else {
			right = middle - 1
		}
	}
	return "-1"
}
