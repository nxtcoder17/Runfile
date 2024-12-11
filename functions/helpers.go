package functions

import (
	"fmt"
	"os"
)

func DefaultIfNil[T any](v *T, dv T) T {
	if v == nil {
		return dv
	}
	return *v
}

// Must panics if err is not nil
// It is intended to be used very sparingly, and only in cases where the caller is
// certain that the error will never be nil in ideal scenarios
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func New[T any](v T) *T {
	return &v
}

func ToEnviron(m map[string]string) []string {
	results := os.Environ()
	for k, v := range m {
		results = append(results, fmt.Sprintf("%s=%v", k, v))
	}
	return results
}
