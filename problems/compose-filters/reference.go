package solution

import (
	"strconv"
	"strings"
)

type Filter interface {
	Apply(int) int
}

type Add struct {
	Value int
}

func (filter Add) Apply(value int) int {
	return value + filter.Value
}

type Mul struct {
	Value int
}

func (filter Mul) Apply(value int) int {
	return value * filter.Value
}

type Clamp struct {
	Min int
	Max int
}

func (filter Clamp) Apply(value int) int {
	if value < filter.Min {
		return filter.Min
	}
	if value > filter.Max {
		return filter.Max
	}
	return value
}

func parseFilter(operation string) (Filter, bool) {
	parts := strings.Fields(operation)
	if len(parts) == 2 && (parts[0] == "add" || parts[0] == "mul") {
		value, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, false
		}
		if parts[0] == "add" {
			return Add{Value: value}, true
		}
		return Mul{Value: value}, true
	}
	if len(parts) == 3 && parts[0] == "clamp" {
		minimum, minErr := strconv.Atoi(parts[1])
		maximum, maxErr := strconv.Atoi(parts[2])
		if minErr != nil || maxErr != nil || minimum > maximum {
			return nil, false
		}
		return Clamp{Min: minimum, Max: maximum}, true
	}
	return nil, false
}

func Solve(nums []int, ops []string) []int {
	filters := make([]Filter, 0, len(ops))
	for _, operation := range ops {
		if filter, ok := parseFilter(operation); ok {
			filters = append(filters, filter)
		}
	}
	result := make([]int, len(nums))
	for index, value := range nums {
		for _, filter := range filters {
			value = filter.Apply(value)
		}
		result[index] = value
	}
	return result
}
