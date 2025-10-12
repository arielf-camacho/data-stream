package operators

import (
	"context"

	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = (primitives.Operator[int, int])(&FilterOperator[int]{})

// FilterOperator is an operator that filters values from the input channel to the
// output channel using the given predicate function. Only values for which the
// predicate returns true are passed through.
//
// Graphically, the FilterOperator looks like this:
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
//
// -- FilterOperator f(x) = x > 2 --
//
// ------------ 3 -- 4 -- 5 -- | -->
type FilterOperator[T any] struct {
	ctx context.Context

	bufferSize uint

	predicate func(T) bool
	in        chan T
	out       chan T
}

// NewFilterOperator returns a new FilterOperator given the predicate function.
func NewFilterOperator[T any](
	predicate func(T) bool,
	opts ...FilterOperatorOption[T],
) *FilterOperator[T] {
	operator := &FilterOperator[T]{
		predicate: predicate,
		ctx:       context.Background(),
	}

	for _, opt := range opts {
		opt(operator)
	}

	operator.in = make(chan T, operator.bufferSize)
	operator.out = make(chan T, operator.bufferSize)

	go operator.start()

	return operator
}

func (f *FilterOperator[T]) In() chan<- T {
	return f.in
}

func (f *FilterOperator[T]) Out() <-chan T {
	return f.out
}

func (f *FilterOperator[T]) To(in primitives.In[T]) {
	go func() {
		defer close(in.In())
		for v := range f.out {
			select {
			case <-f.ctx.Done():
				return
			case in.In() <- v:
			}
		}
	}()
}

func (f *FilterOperator[T]) start() {
	defer close(f.out)

	for v := range f.in {
		select {
		case <-f.ctx.Done():
			return
		default:
			passes := f.predicate(v)
			if passes {
				select {
				case <-f.ctx.Done():
					return
				case f.out <- v:
				}
			}
		}
	}
}
