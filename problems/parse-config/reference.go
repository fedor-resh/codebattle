package solution

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var (
	ErrBadLine      = errors.New("bad line")
	ErrDuplicate    = errors.New("duplicate key")
	ErrInvalidValue = errors.New("invalid value")
)

func Parse(text string) (map[string]int, error) {
	config := make(map[string]int)
	for index, rawLine := range strings.Split(text, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Count(line, "=") != 1 {
			return nil, fmt.Errorf("%w: line %d", ErrBadLine, index+1)
		}
		key, rawValue, _ := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("%w: line %d", ErrBadLine, index+1)
		}
		if _, exists := config[key]; exists {
			return nil, fmt.Errorf("%w: line %d", ErrDuplicate, index+1)
		}
		value, err := strconv.Atoi(strings.TrimSpace(rawValue))
		if err != nil {
			return nil, fmt.Errorf("%w: line %d", ErrInvalidValue, index+1)
		}
		config[key] = value
	}
	return config, nil
}

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
