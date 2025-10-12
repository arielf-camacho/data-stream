package operators

import (
	"context"
	"sync/atomic"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = (primitives.Flow[int, int])(&FilterOperator[int]{})

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

	activated    atomic.Bool
	errorHandler func(error)
	bufferSize   uint

	predicate func(T) (bool, error)
	in        chan T
	out       chan T
}

// FilterBuilder is a fluent builder for FilterOperator.
type FilterBuilder[T any] struct {
	predicate    func(T) (bool, error)
	ctx          context.Context
	errorHandler func(error)
	bufferSize   uint
}

// Filter creates a new FilterBuilder for building a FilterOperator.
func Filter[T any](predicate func(T) (bool, error)) *FilterBuilder[T] {
	return &FilterBuilder[T]{
		predicate: predicate,
		ctx:       context.Background(),
	}
}

// Context sets the context for the FilterOperator.
func (b *FilterBuilder[T]) Context(ctx context.Context) *FilterBuilder[T] {
	b.ctx = ctx
	return b
}

// ErrorHandler sets the error handler for the FilterOperator.
func (b *FilterBuilder[T]) ErrorHandler(handler func(error)) *FilterBuilder[T] {
	b.errorHandler = handler
	return b
}

// BufferSize sets the buffer size for the FilterOperator channels.
func (b *FilterBuilder[T]) BufferSize(size uint) *FilterBuilder[T] {
	b.bufferSize = size
	return b
}

// Build creates and starts the FilterOperator.
func (b *FilterBuilder[T]) Build() *FilterOperator[T] {
	operator := &FilterOperator[T]{
		predicate:    b.predicate,
		ctx:          b.ctx,
		errorHandler: b.errorHandler,
		bufferSize:   b.bufferSize,
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

func (f *FilterOperator[T]) ToFlow(
	in primitives.Flow[T, T],
) primitives.Flow[T, T] {
	f.assertNotActive()

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

	return in
}

func (f *FilterOperator[T]) ToSink(in primitives.Sink[T]) {
	f.assertNotActive()

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

func (f *FilterOperator[T]) assertNotActive() {
	if !f.activated.CompareAndSwap(false, true) {
		// TODO: Use a logger to print this error, don't panic
		panic("FilterOperator is already streaming, cannot be used as a flow again")
	}
}

func (f *FilterOperator[T]) start() {
	defer close(f.out)
	defer helpers.Drain(f.in)

	for v := range f.in {
		select {
		case <-f.ctx.Done():
			return
		default:
			passes, err := f.predicate(v)
			if err != nil {
				if f.errorHandler != nil {
					f.errorHandler(err)
				}
				return
			}
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
