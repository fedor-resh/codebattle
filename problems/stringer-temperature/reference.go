package solution

import "strconv"

type Celsius int

func (value Celsius) String() string {
	return strconv.Itoa(int(value)) + "°C"
}

func Solve(values []int) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, Celsius(value).String())
	}
	return result
}
