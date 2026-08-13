package solution

import "sync"

func Solve(words []string, workers int, work func(string) string) map[string]int {
	if workers <= 0 {
		workers = 1
	}
	jobs := make(chan string)
	frequencies := make(map[string]int)
	var mutex sync.Mutex
	var wait sync.WaitGroup
	wait.Add(workers)
	for range workers {
		go func() {
			defer wait.Done()
			for word := range jobs {
				word = work(word)
				mutex.Lock()
				frequencies[word]++
				mutex.Unlock()
			}
		}()
	}
	for _, word := range words {
		jobs <- word
	}
	close(jobs)
	wait.Wait()
	return frequencies
}
