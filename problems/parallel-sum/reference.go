package solution

import "sync"

func Solve(nums []int, workers int) int {
	if workers <= 0 {
		workers = 1
	}
	jobs := make(chan int)
	partials := make(chan int, workers)
	var wait sync.WaitGroup
	wait.Add(workers)
	for range workers {
		go func() {
			defer wait.Done()
			sum := 0
			for value := range jobs {
				sum += value
			}
			partials <- sum
		}()
	}
	go func() {
		for _, value := range nums {
			jobs <- value
		}
		close(jobs)
	}()
	wait.Wait()
	close(partials)
	total := 0
	for partial := range partials {
		total += partial
	}
	return total
}
