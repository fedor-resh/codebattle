package solution

import (
	"sort"
	"strconv"
	"strings"
)

func Solve(input string) string {
	counts := map[string]int{}
	for _, word := range strings.Fields(input) {
		counts[word]++
	}
	words := make([]string, 0, len(counts))
	for word := range counts {
		words = append(words, word)
	}
	sort.Strings(words)
	result := make([]string, 0, len(words))
	for _, word := range words {
		result = append(result, word+":"+strconv.Itoa(counts[word]))
	}
	return strings.Join(result, ",")
}
