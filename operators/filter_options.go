package operators

import "context"

// FilterOperatorOption is a function that can be used to configure a
// FilterOperator.
type FilterOperatorOption[T any] func(*FilterOperator[T])

// WithContextForFilter returns a FilterOperatorOption that sets the context
// for the FilterOperator.
func WithContextForFilter[T any](
	ctx context.Context,
) FilterOperatorOption[T] {
	return func(f *FilterOperator[T]) {
		f.ctx = ctx
	}
}

// WithBufferSizeForFilter returns a FilterOperatorOption that sets the buffer
// size for the FilterOperator.
func WithBufferSizeForFilter[T any](
	bufferSize uint,
) FilterOperatorOption[T] {
	return func(f *FilterOperator[T]) {
		f.bufferSize = bufferSize
	}
}
