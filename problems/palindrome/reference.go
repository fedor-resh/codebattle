package solution

func Solve(text string) bool {
	runes := []rune(text)
	for left, right := 0, len(runes)-1; left < right; left, right = left+1, right-1 {
		if runes[left] != runes[right] {
			return false
		}
	}
	return true
}
