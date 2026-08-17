package solution

import (
	"errors"
	"sort"
	"strconv"
)

var (
	ErrBadLine      = errors.New("bad line")
	ErrDuplicate    = errors.New("duplicate key")
	ErrInvalidValue = errors.New("invalid value")
)

// Parse разбирает текстовую конфигурацию.
func Parse(text string) (map[string]int, error) {
	// Напишите решение
	return nil, ErrBadLine
}

// Solve адаптирует Parse к JSON-совместимому контракту judge.
func Solve(text string) []string {
	config, err := Parse(text)
	if err != nil {
		return []string{err.Error()}
	}
	keys := make([]string, 0, len(config))
	for key := range config {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, key+"="+strconv.Itoa(config[key]))
	}
	return result
}
