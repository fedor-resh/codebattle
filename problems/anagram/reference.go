package solution

func Solve(first, second string) bool {
	counts := map[rune]int{}
	for _, char := range first {
		counts[char]++
	}
	for _, char := range second {
		counts[char]--
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}
