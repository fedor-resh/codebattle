package solution

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrNotFound = errors.New("not found")

func Lookup(scores map[string]int, name string) (int, error) {
	value, ok := scores[name]
	if !ok {
		return 0, fmt.Errorf("player %s: %w", name, ErrNotFound)
	}
	return value, nil
}

func Solve(scores map[string]int, names []string) []string {
	result := make([]string, 0, len(names))
	for _, name := range names {
		value, err := Lookup(scores, name)
		if err != nil {
			result = append(result, err.Error())
			continue
		}
		result = append(result, strconv.Itoa(value))
	}
	return result
}
