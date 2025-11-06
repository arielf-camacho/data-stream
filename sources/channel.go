package sources

import (
	"context"
	"sync/atomic"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = primitives.Source[any](&ChannelSource[any]{})

// ChannelSource is a source that emits the values of a given channel until it
// is closed or context is cancelled.
//
// Graphically, the ChannelSource looks like this:
//
// ---channel -> 1 -- 2 -- 3 -- 4 -- 5 -- | -->
//
// -- ChannelSource --------------------- | -->
//
// ------------- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
type ChannelSource[T any] struct {
	channel chan T

	ctx          context.Context
	errorHandler func(error)
	activated    atomic.Bool

	out chan T
}

// ChannelSourceBuilder is a fluent builder for ChannelSource.
type ChannelSourceBuilder[T any] struct {
	ctx          context.Context
	errorHandler func(error)
	channel      chan T
}

// Single creates a new ChannelSourceBuilder for building a ChannelSource.
func Channel[T any](channel chan T) *ChannelSourceBuilder[T] {
	return &ChannelSourceBuilder[T]{
		channel: channel,
		ctx:     context.Background(),
	}
}

// Build creates and starts the ChannelSource.
func (b *ChannelSourceBuilder[T]) Build() *ChannelSource[T] {
	source := &ChannelSource[T]{
		channel:      b.channel,
		ctx:          b.ctx,
		errorHandler: b.errorHandler,
		out:          make(chan T, 1),
	}

	go source.start()

	return source
}

// Context sets the context for the ChannelSource.
func (b *ChannelSourceBuilder[T]) Context(
	ctx context.Context,
) *ChannelSourceBuilder[T] {
	b.ctx = ctx
	return b
}

// ErrorHandler sets the error handler for the ChannelSource.
func (b *ChannelSourceBuilder[T]) ErrorHandler(
	handler func(error),
) *ChannelSourceBuilder[T] {
	b.errorHandler = handler
	return b
}

// Out returns the channel from which the values can be read.
func (s *ChannelSource[T]) Out() <-chan T {
	return s.out
}

// ToFlow streams the values from the ChannelSource to the given flow. This is
// exclusive with ToSink, either one must be called.
func (s *ChannelSource[T]) ToFlow(
	in primitives.Flow[T, T],
) primitives.Flow[T, T] {
	s.assertNotActive()
	helpers.StreamTo(s.ctx, s.out, in.In())
	return in
}

// ToSink streams the values from the ChannelSource to the given sink. This is
// exclusive with ToFlow, either one must be called.
func (s *ChannelSource[T]) ToSink(in primitives.Sink[T]) primitives.Sink[T] {
	s.assertNotActive()
	helpers.StreamTo(s.ctx, s.out, in.In())
	return in
}

func (s *ChannelSource[T]) assertNotActive() {
	if !s.activated.CompareAndSwap(false, true) {
		// TODO: Use a logger to print this error, don't panic
		panic("ChannelSource is already streaming, cannot be used as a source again")
	}
}

func (s *ChannelSource[T]) start() {
	defer close(s.out)

	for {
		select {
		case <-s.ctx.Done():
			return
		case value, ok := <-s.channel:
			if !ok {
				return
			}

			select {
			case <-s.ctx.Done():
				return
			case s.out <- value:
			}
		}
	}
}
