package solution

func Solve(words []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, word := range words {
		if !seen[word] {
			seen[word] = true
			result = append(result, word)
		}
	}
	return result
}
