package sources

import (
	"context"
	"sync/atomic"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = primitives.Source[any](&SingleSource[any]{})

// SingleSource is a source that emits a single value.
//
// Graphically, the SingleSource looks like this:
//
// ----------------------------- | -->
//
// -- SingleSource f(x) = 1 ----------
//
// -----------------------1 ---- | -->
type SingleSource[T any] struct {
	get func() (T, error)

	ctx          context.Context
	errorHandler func(error)
	activated    atomic.Bool

	out chan T
}

// SingleSourceBuilder is a fluent builder for SingleSource.
type SingleSourceBuilder[T any] struct {
	ctx          context.Context
	errorHandler func(error)
	get          func() (T, error)
}

// Single creates a new SingleSourceBuilder for building a SingleSource.
func Single[T any](get func() (T, error)) *SingleSourceBuilder[T] {
	return &SingleSourceBuilder[T]{
		get: get,
		ctx: context.Background(),
	}
}

// Build creates and starts the SingleSource.
func (b *SingleSourceBuilder[T]) Build() *SingleSource[T] {
	source := &SingleSource[T]{
		get:          b.get,
		ctx:          b.ctx,
		errorHandler: b.errorHandler,
		out:          make(chan T, 1),
	}

	go source.start()

	return source
}

// Context sets the context for the SingleSource.
func (b *SingleSourceBuilder[T]) Context(
	ctx context.Context,
) *SingleSourceBuilder[T] {
	b.ctx = ctx
	return b
}

// ErrorHandler sets the error handler for the SingleSource.
func (b *SingleSourceBuilder[T]) ErrorHandler(
	handler func(error),
) *SingleSourceBuilder[T] {
	b.errorHandler = handler
	return b
}

// Out returns the channel from which the values can be read.
func (s *SingleSource[T]) Out() <-chan T {
	return s.out
}

// ToFlow streams the values from the SingleSource to the given flow. This is
// exclusive with ToSink, either one must be called.
func (s *SingleSource[T]) ToFlow(
	in primitives.Flow[T, T],
) primitives.Flow[T, T] {
	s.assertNotActive()
	helpers.StreamTo(s.ctx, s.out, in.In())
	return in
}

// ToSink streams the values from the SingleSource to the given sink. This is
// exclusive with ToFlow, either one must be called.
func (s *SingleSource[T]) ToSink(in primitives.Sink[T]) primitives.Sink[T] {
	s.assertNotActive()
	helpers.StreamTo(s.ctx, s.out, in.In())
	return in
}

func (s *SingleSource[T]) assertNotActive() {
	if !s.activated.CompareAndSwap(false, true) {
		// TODO: Use a logger to print this error, don't panic
		panic("SingleSource is already streaming, cannot be used as a source again")
	}
}

func (s *SingleSource[T]) start() {
	defer close(s.out)

	// Early return if context already cancelled
	select {
	case <-s.ctx.Done():
		return
	default:
	}

	value, err := s.get()
	if err != nil {
		if s.errorHandler != nil {
			s.errorHandler(err)
		}
		return
	}

	select {
	case <-s.ctx.Done():
		return
	case s.out <- value:
	}
}
