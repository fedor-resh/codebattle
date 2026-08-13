package solution

import (
	"strings"
	"sync"
)

func Solve(n int) string {
	if n <= 0 {
		return ""
	}

	ping := make(chan struct{})
	pong := make(chan struct{})
	done := make(chan struct{})
	var result strings.Builder
	var wait sync.WaitGroup
	wait.Add(2)

	go func() {
		defer wait.Done()
		for range n {
			<-ping
			result.WriteString("ping")
			pong <- struct{}{}
		}
	}()
	go func() {
		defer wait.Done()
		for index := range n {
			<-pong
			result.WriteString("pong")
			if index+1 == n {
				close(done)
			} else {
				ping <- struct{}{}
			}
		}
	}()

	ping <- struct{}{}
	<-done
	wait.Wait()
	return result.String()
}
