package solution

import "strings"

func Solve(input string) string {
	parts := strings.SplitN(input, "|", 2)
	if len(parts) != 2 {
		return "false"
	}
	counts := map[rune]int{}
	for _, char := range parts[0] {
		counts[char]++
	}
	for _, char := range parts[1] {
		counts[char]--
	}
	for _, count := range counts {
		if count != 0 {
			return "false"
		}
	}
	return "true"
}
