package operators

import "context"

// MapOperatorOption is a function that can be used to configure a MapOperator.
type MapOperatorOption[IN any, OUT any] func(*MapOperator[IN, OUT])

// WithContext returns a MapOperatorOption that sets the context for the
// MapOperator.
func WithContext[IN any, OUT any](
	ctx context.Context,
) MapOperatorOption[IN, OUT] {
	return func(o *MapOperator[IN, OUT]) {
		o.ctx = ctx
	}
}

// WithParallelism returns a MapOperatorOption that sets the parallelism for the
// MapOperator.
func WithParallelism[IN any, OUT any](
	parallelism uint,
) MapOperatorOption[IN, OUT] {
	return func(o *MapOperator[IN, OUT]) {
		o.parallelism = parallelism
	}
}
