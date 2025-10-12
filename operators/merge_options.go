package operators

import "context"

// MergeOperatorOption is a function that can be used to configure a
// MergeOperator.
type MergeOperatorOption func(*MergeOperator)

// WithBufferSizeForMerge returns a MergeOperatorOption that sets the buffer size
// for the MergeOperator.
func WithBufferSizeForMerge(bufferSize uint) MergeOperatorOption {
	return func(o *MergeOperator) {
		o.bufferSize = bufferSize
	}
}

// WithContext returns a MergeOperatorOption that sets the context for the
// MergeOperator.
func WithContextForMerge[IN any, OUT any](
	ctx context.Context,
) MergeOperatorOption {
	return func(o *MergeOperator) {
		o.ctx = ctx
	}
}
