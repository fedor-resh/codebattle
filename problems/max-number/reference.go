package solution

func Solve(nums []int) int {
	maximum := nums[0]
	for _, value := range nums[1:] {
		if value > maximum {
			maximum = value
		}
	}
	return maximum
}
