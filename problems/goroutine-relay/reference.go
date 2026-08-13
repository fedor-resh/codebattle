package solution

import (
	"strings"
	"sync"
)

func Solve(n int, work func(string) string) string {
	if n <= 0 {
		return ""
	}

	pings := make([]string, n)
	pongs := make([]string, n)
	var wait sync.WaitGroup
	wait.Add(2)

	go func() {
		defer wait.Done()
		for index := range n {
			pings[index] = work("ping")
		}
	}()
	go func() {
		defer wait.Done()
		for index := range n {
			pongs[index] = work("pong")
		}
	}()

	wait.Wait()
	var result strings.Builder
	for index := range n {
		result.WriteString(pings[index])
		result.WriteString(pongs[index])
	}
	return result.String()
}
