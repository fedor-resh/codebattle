package solution

import (
	"context"
	"sync"
	"time"
)

func Solve(delays []int, timeoutMS int) int {
	if len(delays) == 0 || timeoutMS <= 0 {
		return -1
	}

	ctx, cancel := context.WithCancel(context.Background())
	results := make(chan int, len(delays))
	var wait sync.WaitGroup
	wait.Add(len(delays))
	for index, delay := range delays {
		go func() {
			defer wait.Done()
			timer := time.NewTimer(time.Duration(delay) * time.Millisecond)
			defer timer.Stop()
			select {
			case <-timer.C:
				select {
				case results <- index:
				case <-ctx.Done():
				}
			case <-ctx.Done():
			}
		}()
	}

	timeout := time.NewTimer(time.Duration(timeoutMS) * time.Millisecond)
	defer timeout.Stop()
	answer := -1
	select {
	case answer = <-results:
	case <-timeout.C:
	}
	cancel()
	wait.Wait()
	return answer
}
