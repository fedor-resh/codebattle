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
	return len(users)
}

func (users ByScore) Less(i, j int) bool {
	if users[i].Score != users[j].Score {
		return users[i].Score > users[j].Score
	}
	return users[i].Name < users[j].Name
}

func (users ByScore) Swap(i, j int) {
	users[i], users[j] = users[j], users[i]
}

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
