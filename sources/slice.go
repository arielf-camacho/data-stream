package sources

import (
	"context"
	"sync/atomic"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = primitives.Source[any](&SliceSource[any]{})

// SliceSource is a source that emits the values of a given slice. It emits all
// values in the order they are given through the Out() channel. If the context
// is cancelled, the SliceSource will stop emitting values and close the Out()
// channel. When the SliceSource has emitted all values, the Out() channel will
// be closed automatically.
//
// Graphically, the SliceSource looks like this:
//
//	SliceSource (1, 2, 3, 4, 5, ...)
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
type SliceSource[T any] struct {
	slice []T

	activated  atomic.Bool
	ctx        context.Context
	bufferSize uint

	out chan T
}

// SliceSourceBuilder is a fluent builder for SliceSource.
type SliceSourceBuilder[T any] struct {
	slice      []T
	ctx        context.Context
	bufferSize uint
}

// Slice creates a new SliceSourceBuilder for building a SliceSource.
func Slice[T any](slice []T) *SliceSourceBuilder[T] {
	return &SliceSourceBuilder[T]{
		slice: slice,
		ctx:   context.Background(),
	}
}

// Context sets the context for the SliceSource.
func (b *SliceSourceBuilder[T]) Context(
	ctx context.Context,
) *SliceSourceBuilder[T] {
	b.ctx = ctx
	return b
}

// BufferSize sets the buffer size for the SliceSource output channel.
func (b *SliceSourceBuilder[T]) BufferSize(size uint) *SliceSourceBuilder[T] {
	b.bufferSize = size
	return b
}

// Build creates and starts the SliceSource.
func (b *SliceSourceBuilder[T]) Build() *SliceSource[T] {
	source := &SliceSource[T]{
		slice:      b.slice,
		ctx:        b.ctx,
		bufferSize: b.bufferSize,
	}

	source.out = make(chan T, source.bufferSize)

	go source.start()

	return source
}

// Out returns the channel from which the values can be read.
func (s *SliceSource[T]) Out() <-chan T {
	return s.out
}

// ToFlow streams the values from the SliceSource to the given flow. This is
// exclusive with ToSink, either one must be called.
func (s *SliceSource[T]) ToFlow(
	in primitives.Flow[T, T],
) primitives.Flow[T, T] {
	s.assertNotActive()

	go func() {
		defer close(in.In())
		for v := range s.out {
			select {
			case <-s.ctx.Done():
				helpers.Drain(s.out)
				return
			case in.In() <- v:
			}
		}
	}()

	return in
}

// ToSink streams the values from the SliceSource to the given sink. This is
// exclusive with ToFlow, either one must be called.
func (s *SliceSource[T]) ToSink(in primitives.Sink[T]) primitives.Sink[T] {
	s.assertNotActive()

	go func() {
		defer close(in.In())
		for v := range s.out {
			select {
			case <-s.ctx.Done():
				helpers.Drain(s.out)
				return
			case in.In() <- v:
			}
		}
	}()

	return in
}

func (s *SliceSource[T]) assertNotActive() {
	if !s.activated.CompareAndSwap(false, true) {
		// TODO: Use a logger to print this error, don't panic
		panic("SliceSource is already streaming, cannot be used as a source again")
	}
}

func (s *SliceSource[T]) start() {
	defer close(s.out)

	for _, v := range s.slice {
		select {
		case <-s.ctx.Done():
			return
		case s.out <- v:
		}
	}
}
