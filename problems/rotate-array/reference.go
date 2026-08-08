package solution

func Solve(nums []int, k int) []int {
	if len(nums) == 0 {
		return []int{}
	}
	k %= len(nums)
	result := append([]int{}, nums[len(nums)-k:]...)
	result = append(result, nums[:len(nums)-k]...)
	return result
}
