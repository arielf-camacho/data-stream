package sinks

import (
	"context"

	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = primitives.Sink[any](&ChannelSink[any]{})

type ChannelSink[T any] struct {
	in  chan T
	out chan T

	ctx        context.Context
	bufferSize uint
}

func NewChannelSink[T any](
	channel chan T,
	opts ...ChannelSinkOption[T],
) *ChannelSink[T] {
	ch := &ChannelSink[T]{
		out: channel,
		ctx: context.Background(),
	}

	for _, opt := range opts {
		opt(ch)
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

	for v := range c.in {
		select {
		case <-c.ctx.Done():
			return
		case c.out <- v:
		}
	}
}
