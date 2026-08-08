package solution

func Solve(n int) uint64 {
	var a, b uint64 = 0, 1
	for i := 0; i < n; i++ {
		a, b = b, a+b
	}
	return a
}
