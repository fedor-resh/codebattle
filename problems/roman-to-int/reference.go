package solution

import "strconv"

func Solve(input string) string {
	values := map[byte]int{'I': 1, 'V': 5, 'X': 10, 'L': 50, 'C': 100, 'D': 500, 'M': 1000}
	total := 0
	for i := 0; i < len(input); i++ {
		value := values[input[i]]
		if i+1 < len(input) && value < values[input[i+1]] {
			total -= value
		} else {
			total += value
		}
	}
	return strconv.Itoa(total)
}
