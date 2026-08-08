package solution

func Solve(n int) bool {
	if n < 2 {
		return false
	}
	for divisor := 2; divisor <= n/divisor; divisor++ {
		if n%divisor == 0 {
			return false
		}
	}
	return true
}
