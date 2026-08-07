package solution

import "strings"

func Solve(input string) string {
	parts := strings.SplitN(input, "|", 2)
	items := strings.Split(parts[1], ",")
	if len(items) == 0 || parts[1] == "" {
		return ""
	}
	k := 0
	for _, char := range parts[0] {
		k = k*10 + int(char-'0')
	}
	k %= len(items)
	result := append([]string{}, items[len(items)-k:]...)
	result = append(result, items[:len(items)-k]...)
	return strings.Join(result, ",")
}
