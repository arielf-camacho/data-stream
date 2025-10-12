package slice

import (
	"context"

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

	ctx        context.Context
	err        chan error
	out        chan T
	bufferSize uint
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

	source.err = make(chan error, 1)
	source.out = make(chan T, source.bufferSize)

	go source.start()

	return source
}

func (s *SliceSource[T]) Err() <-chan error {
	return s.err
}

func (s *SliceSource[T]) Out() <-chan T {
	return s.out
}

func (s *SliceSource[T]) To(in primitives.In[T]) {
	go func() {
		defer close(in.In())
		for v := range s.out {
			select {
			case <-s.ctx.Done():
				return
			case in.In() <- v:
			}
		}
	}()
}

func (s *SliceSource[T]) start() {
	defer close(s.err)
	defer close(s.out)

	for _, v := range s.slice {
		select {
		case <-s.ctx.Done():
			return
		case s.out <- v:
		}
	}
}
