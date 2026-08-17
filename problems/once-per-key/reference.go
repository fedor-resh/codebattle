package solution

import "sync"

func Solve(keys []int, work func(int) int) []int {
	positions := make(map[int][]int)
	for index, key := range keys {
		positions[key] = append(positions[key], index)
	}

	result := make([]int, len(keys))
	var wait sync.WaitGroup
	wait.Add(len(positions))
	for key, indexes := range positions {
		go func() {
			defer wait.Done()
			value := work(key)
			for _, index := range indexes {
				result[index] = value
			}
		}()
	}
	wait.Wait()
	return result
}
