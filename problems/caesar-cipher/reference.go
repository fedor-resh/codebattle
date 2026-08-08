package solution

func Solve(text string, shift int) string {
	shift = ((shift % 26) + 26) % 26
	result := []rune(text)
	for index, char := range result {
		switch {
		case char >= 'a' && char <= 'z':
			result[index] = 'a' + (char-'a'+rune(shift))%26
		case char >= 'A' && char <= 'Z':
			result[index] = 'A' + (char-'A'+rune(shift))%26
		}
	}
	return string(result)
}
