package solution

func Solve(nums []int, target int) []int {
	seen := map[int]int{}
	for index, value := range nums {
		if other, ok := seen[target-value]; ok {
			return []int{other, index}
		}
		seen[value] = index
	}
	return nil
}
