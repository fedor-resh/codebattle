package solution

func Solve(input string) string {
	runes := []rune(input)
	for left, right := 0, len(runes)-1; left < right; left, right = left+1, right-1 {
		if runes[left] != runes[right] {
			return "false"
		}
	}
	return "true"
}
