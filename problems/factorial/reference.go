package solution

func Solve(n int) uint64 {
	var result uint64 = 1
	for i := 2; i <= n; i++ {
		result *= uint64(i)
	}
	return result
}
