package solution

func Solve(nums []int, target int) int {
	left, right := 0, len(nums)-1
	for left <= right {
		middle := left + (right-left)/2
		value := nums[middle]
		if value == target {
			return middle
		}
		if value < target {
			left = middle + 1
		} else {
			right = middle - 1
		}
	}
	return -1
}
