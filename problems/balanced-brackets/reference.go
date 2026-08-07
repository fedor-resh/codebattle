package solution

func Solve(input string) string {
	stack := []rune{}
	pairs := map[rune]rune{')': '(', ']': '[', '}': '{'}
	for _, char := range input {
		switch char {
		case '(', '[', '{':
			stack = append(stack, char)
		default:
			if len(stack) == 0 || stack[len(stack)-1] != pairs[char] {
				return "false"
			}
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) == 0 {
		return "true"
	}
	return "false"
}
