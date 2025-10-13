package sinks

import (
	"context"

	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = primitives.Sink[any](&ChannelSink[any]{})

// ChannelSink is a sink that writes the values to a channel.
//
// Graphically, the ChannelSink looks like this:
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
// -- ChannelSink --
// -> 1 -- 2 -- 3 -- 4 -- 5 -- |
type ChannelSink[T any] struct {
	in  chan T
	out chan T

	ctx        context.Context
	bufferSize uint
}

// ChannelSinkBuilder is a fluent builder for ChannelSink.
type ChannelSinkBuilder[T any] struct {
	out        chan T
	ctx        context.Context
	bufferSize uint
}

// Channel creates a new ChannelSinkBuilder for building a ChannelSink.
func Channel[T any](channel chan T) *ChannelSinkBuilder[T] {
	return &ChannelSinkBuilder[T]{
		out: channel,
		ctx: context.Background(),
	}
}

// Context sets the context for the ChannelSink.
func (b *ChannelSinkBuilder[T]) Context(
	ctx context.Context,
) *ChannelSinkBuilder[T] {
	b.ctx = ctx
	return b
}

// BufferSize sets the buffer size for the ChannelSink input channel.
func (b *ChannelSinkBuilder[T]) BufferSize(size uint) *ChannelSinkBuilder[T] {
	b.bufferSize = size
	return b
}

// Build creates and starts the ChannelSink.
func (b *ChannelSinkBuilder[T]) Build() *ChannelSink[T] {
	ch := &ChannelSink[T]{
		out:        b.out,
		ctx:        b.ctx,
		bufferSize: b.bufferSize,
	}

	ch.in = make(chan T, ch.bufferSize)

	go ch.start()

	return ch
}

func (c *ChannelSink[T]) In() chan<- T {
	return c.in
}

func (c *ChannelSink[T]) start() {
	defer close(c.out)

	for {
		select {
		case <-c.ctx.Done():
			return
		case v, ok := <-c.in:
			if !ok {
				return
			}
			select {
			case <-c.ctx.Done():
				return
			case c.out <- v:
			}
		}
	}
}
