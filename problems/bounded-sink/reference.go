package solution

import (
	"errors"
	"strconv"
)

var ErrFull = errors.New("sink is full")

type Sink interface {
	Write(string) (int, error)
}

type LimitedSink struct {
	Limit int
	used  int
}

func (sink *LimitedSink) Write(chunk string) (int, error) {
	if len(chunk) > sink.Limit-sink.used {
		return 0, ErrFull
	}
	sink.used += len(chunk)
	return len(chunk), nil
}

func Solve(limit int, chunks []string) []string {
	sink := &LimitedSink{Limit: limit}
	result := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		written, err := sink.Write(chunk)
		if err != nil {
			result = append(result, err.Error())
			continue
		}
		result = append(result, strconv.Itoa(written))
	}
	return result
}
