package solution

import (
	"sort"
	"strconv"
)

type User struct {
	Name  string
	Score int
}

type ByScore []User

func (users ByScore) Len() int {
	// Напишите решение
	return 0
}

func (users ByScore) Less(i, j int) bool {
	// Напишите решение
	return false
}

func (users ByScore) Swap(i, j int) {
	// Напишите решение
}

// Solve адаптирует sort.Interface к JSON-совместимому контракту judge.
func Solve(names []string, scores []int) []string {
	count := len(names)
	if len(scores) < count {
		count = len(scores)
	}
	users := make(ByScore, count)
	for index := 0; index < count; index++ {
		users[index] = User{Name: names[index], Score: scores[index]}
	}
	sort.Sort(users)
	result := make([]string, count)
	for index, user := range users {
		result[index] = user.Name + ":" + strconv.Itoa(user.Score)
	}
	return result
}
