package solution

import (
	"strings"
	"unicode/utf8"
)

func Solve(text string) string {
	longest := ""
	for _, word := range strings.Fields(text) {
		if utf8.RuneCountInString(word) > utf8.RuneCountInString(longest) {
			longest = word
		}
	}
	return longest
}
