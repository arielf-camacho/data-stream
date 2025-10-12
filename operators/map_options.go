package operators

import "context"

// MapOperatorOption is a function that can be used to configure a MapOperator.
type MapOperatorOption[IN any, OUT any] func(*MapOperator[IN, OUT])

// WithContextForMap returns a MapOperatorOption that sets the context for the
// MapOperator.
func WithContextForMap[IN any, OUT any](
	ctx context.Context,
) MapOperatorOption[IN, OUT] {
	return func(o *MapOperator[IN, OUT]) {
		o.ctx = ctx
	}
}

// WithErrorHandlerForMap returns a MapOperatorOption that sets the error
// handler for the MapOperator.
func WithErrorHandlerForMap[IN any, OUT any](
	errorHandler func(error),
) MapOperatorOption[IN, OUT] {
	return func(o *MapOperator[IN, OUT]) {
		o.errorHandler = errorHandler
	}
}

// WithParallelismForMap returns a MapOperatorOption that sets the parallelism
// for the MapOperator.
func WithParallelismForMap[IN any, OUT any](
	parallelism uint,
) MapOperatorOption[IN, OUT] {
	return func(o *MapOperator[IN, OUT]) {
		o.parallelism = parallelism
	}
}
