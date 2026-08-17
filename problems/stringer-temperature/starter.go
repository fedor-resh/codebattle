package solution

type Celsius int

func (value Celsius) String() string {
	// Напишите решение
	return ""
}

// Solve адаптирует Celsius к JSON-совместимому контракту judge.
func Solve(values []int) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, Celsius(value).String())
	}
	return result
}
