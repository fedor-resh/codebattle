package solution

func Solve(words []string) map[string]int {
	counts := map[string]int{}
	for _, word := range words {
		counts[word]++
	}
	return counts
}
