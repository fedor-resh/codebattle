package solution

func Solve(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
