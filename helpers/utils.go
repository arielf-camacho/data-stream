package helpers

import (
	"context"
)

// Collect collects the values from the given channel into a slice and returns
// it. If the context is done, the function returns the collected values so far.
func Collect[T any](ctx context.Context, source <-chan T) []T {
	var result []T
	for {
		select {
		case <-ctx.Done():
			return result
		case v, ok := <-source:
			if !ok {
				return result
			}
			result = append(result, v)
		}
	}
}

// Drain drains the given channel until it is closed.
func Drain[T any](source <-chan T) {
	for range source {
	}
}
