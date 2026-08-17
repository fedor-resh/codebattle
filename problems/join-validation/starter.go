package solution

import (
	"errors"
	"sort"
)

var (
	ErrEmpty     = errors.New("empty")
	ErrTooLong   = errors.New("too long")
	ErrDuplicate = errors.New("duplicate")
)

type FieldError struct {
	Field string
	Err   error
}

func (err *FieldError) Error() string {
	if err == nil || err.Err == nil {
		return ""
	}
	return err.Field + ": " + err.Err.Error()
}

func (err *FieldError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

// Validate собирает все нарушения в одну ошибку.
func Validate(fields []string) error {
	// Напишите решение
	return nil
}

func fieldErrors(err error) []*FieldError {
	if err == nil {
		return nil
	}
	if multi, ok := err.(interface{ Unwrap() []error }); ok {
		var collected []*FieldError
		for _, inner := range multi.Unwrap() {
			collected = append(collected, fieldErrors(inner)...)
		}
		return collected
	}
	var fieldErr *FieldError
	if errors.As(err, &fieldErr) {
		return []*FieldError{fieldErr}
	}
	return nil
}

// Solve адаптирует Validate к JSON-совместимому контракту judge.
func Solve(fields []string) []string {
	items := make([]string, 0)
	for _, fieldErr := range fieldErrors(Validate(fields)) {
		items = append(items, fieldErr.Field+":"+fieldErr.Err.Error())
	}
	sort.Strings(items)
	return items
}
