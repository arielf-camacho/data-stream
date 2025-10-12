package sources

import (
	"context"

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

func (s *SingleSource[T]) Out() <-chan T {
	return s.out
}

func (s *SingleSource[T]) To(in primitives.In[T]) {
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

func (s *SingleSource[T]) start() {
	defer close(s.out)

	select {
	case <-s.ctx.Done():
		return
	default:
		value, err := s.get()
		if err != nil {
			if s.errorHandler != nil {
				s.errorHandler(err)
			}
			return
		}
		s.out <- value
	}
}
