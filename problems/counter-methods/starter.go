package solution

import (
	"strconv"
	"strings"
)

type Counter struct {
	value int
}

func (counter *Counter) Inc() {
	// Напишите решение
}

func (counter *Counter) Add(delta int) {
	// Напишите решение
}

func (counter *Counter) Value() int {
	// Напишите решение
	return 0
}

// Solve адаптирует методы Counter к JSON-совместимому контракту judge.
func Solve(commands []string) []int {
	counter := &Counter{}
	result := make([]int, 0, len(commands))
	for _, command := range commands {
		parts := strings.Fields(command)
		switch {
		case len(parts) == 1 && parts[0] == "inc":
			counter.Inc()
		case len(parts) == 2 && parts[0] == "add":
			if delta, err := strconv.Atoi(parts[1]); err == nil {
				counter.Add(delta)
			}
		}
		result = append(result, counter.Value())
	}
	return result
}
