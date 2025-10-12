package slice

import "context"

// SliceSourceOption is a function that can be used to configure a SliceSource.
type SliceSourceOption[T any] func(*SliceSource[T])

// WithBuffer returns a SliceSourceOption that sets the buffer size for the
// SliceSource.
func WithBuffer[T any](buffer uint) SliceSourceOption[T] {
	return func(s *SliceSource[T]) {
		s.bufferSize = buffer
	}
}

// WithContext returns a SliceSourceOption that sets the context for the
// SliceSource.
func WithContext[T any](ctx context.Context) SliceSourceOption[T] {
	return func(s *SliceSource[T]) {
		s.ctx = ctx
	}
}
