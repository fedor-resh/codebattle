package solution

func Solve(nums []int) []int {
	result := make([]int, len(nums))
	sum := 0
	for index, value := range nums {
		sum += value
		result[index] = sum
	}
	return result
}
