package solution

import "strings"

func Solve(input string) string {
	seen := map[string]bool{}
	result := []string{}
	for _, word := range strings.Fields(input) {
		if !seen[word] {
			seen[word] = true
			result = append(result, word)
		}
	}
	return strings.Join(result, ",")
}
