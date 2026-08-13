package solution

import (
	"context"
	"sync"
	"time"
)

func Solve(delays []int, timeoutMS int, work func(context.Context, int) bool) int {
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
			if work(ctx, delay) {
				select {
				case results <- index:
				case <-ctx.Done():
				}
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
