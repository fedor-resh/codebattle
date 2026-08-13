package solution

import "sync"

type squareJob struct {
	index int
	value int
}

func Solve(nums []int, workers int, work func(int) int) []int {
	if workers <= 0 {
		workers = 1
	}
	result := make([]int, len(nums))
	jobs := make(chan squareJob)
	var wait sync.WaitGroup
	wait.Add(workers)
	for range workers {
		go func() {
			defer wait.Done()
			for job := range jobs {
				result[job.index] = work(job.value)
			}
		}()
	}
	for index, value := range nums {
		jobs <- squareJob{index: index, value: value}
	}
	close(jobs)
	wait.Wait()
	return result
}
