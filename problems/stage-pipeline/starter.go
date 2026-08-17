package solution

// Solve применяет две стадии к каждому элементу, сохраняя порядок результатов.
func Solve(nums []int, stageA func(int) int, stageB func(int) int) []int {
	result := make([]int, len(nums))
	for index, value := range nums {
		result[index] = stageB(stageA(value))
	}
	return result
}
