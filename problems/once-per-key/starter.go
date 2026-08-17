package solution

// Solve возвращает результат work для каждого ключа.
func Solve(keys []int, work func(int) int) []int {
	result := make([]int, len(keys))
	for index, key := range keys {
		result[index] = work(key)
	}
	return result
}
