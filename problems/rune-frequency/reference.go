package solution

func Solve(text string) map[string]int {
	result := make(map[string]int)
	for _, symbol := range text {
		result[string(symbol)]++
	}
	return result
}
