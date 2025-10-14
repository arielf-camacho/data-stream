package flows

import (
	"context"
	"sync/atomic"

	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = (primitives.Flow[int, int])(&FilterFlow[int]{})

// FilterFlow is an operator that filters values from the input channel to the
// output channel using the given predicate function. Only values for which the
// predicate returns true are passed through.
//
// Graphically, the FilterFlow looks like this:
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
//
// -- FilterFlow f(x) = x > 2 --
//
// ------------ 3 -- 4 -- 5 -- | -->
type FilterFlow[T any] struct {
	ctx context.Context

	activated    atomic.Bool
	errorHandler func(error)
	bufferSize   uint

	predicate func(T) (bool, error)
	in        chan T
	out       chan T
}

// FilterBuilder is a fluent builder for FilterFlow.
type FilterBuilder[T any] struct {
	predicate    func(T) (bool, error)
	ctx          context.Context
	errorHandler func(error)
	bufferSize   uint
}

// Filter creates a new FilterBuilder for building a FilterFlow.
func Filter[T any](predicate func(T) (bool, error)) *FilterBuilder[T] {
	return &FilterBuilder[T]{
		predicate: predicate,
		ctx:       context.Background(),
	}
}

// Context sets the context for the FilterFlow.
func (b *FilterBuilder[T]) Context(ctx context.Context) *FilterBuilder[T] {
	b.ctx = ctx
	return b
}

// ErrorHandler sets the error handler for the FilterFlow.
func (b *FilterBuilder[T]) ErrorHandler(handler func(error)) *FilterBuilder[T] {
	b.errorHandler = handler
	return b
}

// BufferSize sets the buffer size for the FilterFlow channels.
func (b *FilterBuilder[T]) BufferSize(size uint) *FilterBuilder[T] {
	b.bufferSize = size
	return b
}

// Build creates and starts the FilterFlow.
func (b *FilterBuilder[T]) Build() *FilterFlow[T] {
	operator := &FilterFlow[T]{
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

func (f *FilterFlow[T]) In() chan<- T {
	return f.in
}

func (f *FilterFlow[T]) Out() <-chan T {
	return f.out
}

func (f *FilterFlow[T]) ToFlow(
	in primitives.Flow[T, T],
) primitives.Flow[T, T] {
	f.assertNotActive()

	go func() {
		defer close(in.In())
		for {
			select {
			case <-f.ctx.Done():
				return
			case v, ok := <-f.out:
				if !ok {
					return
				}
				select {
				case <-f.ctx.Done():
					return
				case in.In() <- v:
				}
			}
		}
	}()

	return in
}

func (f *FilterFlow[T]) ToSink(in primitives.Sink[T]) {
	f.assertNotActive()

	go func() {
		defer close(in.In())
		for {
			select {
			case <-f.ctx.Done():
				return
			case v, ok := <-f.out:
				if !ok {
					return
				}
				select {
				case <-f.ctx.Done():
					return
				case in.In() <- v:
				}
			}
		}
	}()
}

func (f *FilterFlow[T]) assertNotActive() {
	if !f.activated.CompareAndSwap(false, true) {
		// TODO: Use a logger to print this error, don't panic
		panic("FilterFlow is already streaming, cannot be used as a flow again")
	}
}

func (f *FilterFlow[T]) start() {
	defer close(f.out)

	for {
		select {
		case <-f.ctx.Done():
			return
		case v, ok := <-f.in:
			if !ok {
				return
			}
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
