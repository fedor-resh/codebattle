package solution

func Solve(nums []int) bool {
	seen := make(map[int]struct{}, len(nums))
	for _, value := range nums {
		if _, exists := seen[value]; exists {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}
