package solution

import (
	"strconv"
	"strings"
)

func Solve(input string) string {
	parts := strings.SplitN(input, "|", 2)
	shift, _ := strconv.Atoi(parts[0])
	shift = ((shift % 26) + 26) % 26
	result := []rune(parts[1])
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
