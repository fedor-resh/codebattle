package solution

func Solve(text string) int {
	count := 0
	for _, char := range text {
		switch char {
		case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
			count++
		}
	}
	return count
}
