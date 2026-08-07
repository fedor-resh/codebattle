package solution

import "strconv"

func Solve(input string) string {
	count := 0
	for _, char := range input {
		switch char {
		case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
			count++
		}
	}
	return strconv.Itoa(count)
}
