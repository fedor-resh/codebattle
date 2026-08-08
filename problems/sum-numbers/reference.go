package solution

func Solve(nums []int) int {
	sum := 0
	for _, value := range nums {
		sum += value
	}
	return sum
}
