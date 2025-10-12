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
	out        chan T
	bufferSize uint
}

// NewSliceSource returns a new SliceSource given the slice of values to stream
// and the options to configure the SliceSource.
func NewSliceSource[T any](
	slice []T,
	opts ...SliceSourceOption[T],
) *SliceSource[T] {
	source := &SliceSource[T]{
		slice: slice,
		ctx:   context.Background(),
	}

	for _, opt := range opts {
		opt(source)
	}

	source.out = make(chan T, source.bufferSize)

	go source.start()

	return source
}

func (s *SliceSource[T]) Out() <-chan T {
	return s.out
}

func (s *SliceSource[T]) To(in primitives.In[T]) {
	go func() {
		defer close(in.In())
		for v := range s.out {
			in.In() <- v
		}
	}()
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
