package solution

type pipelineItem struct {
	index int
	value int
}

func Solve(nums []int, stageA func(int) int, stageB func(int) int) []int {
	result := make([]int, len(nums))
	intermediate := make(chan pipelineItem)
	go func() {
		defer close(intermediate)
		for index, value := range nums {
			intermediate <- pipelineItem{index: index, value: stageA(value)}
		}
	}()
	for item := range intermediate {
		result[item.index] = stageB(item.value)
	}
	return result
}
