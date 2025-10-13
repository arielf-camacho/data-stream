package flows

import (
	"context"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
)

// SplitFlow is an operator that splits the values from the input channel to
// the output channels based on the given predicate function.
//
// Graphically, the SplitFlow looks like this:
//
// -- 1 -- 2 -- 3 -- 4 -- 5 ------ | -->
//
// -- SplitFlow f(x) = x % 2 == 0 --
//
// Matching:
// ------- 2 ------- 4 ----------- | -->
//
// NonMatching:
// -- 1 ------- 3 ------- 5 ------ | -->
type SplitFlow[T any] struct {
	ctx          context.Context
	bufferSize   uint
	errorHandler func(error)
	predicate    func(T) (bool, error)

	in          primitives.Outlet[T]
	matching    primitives.Flow[T, T]
	nonMatching primitives.Flow[T, T]
}

// SplitBuilder is a fluent builder for SplitFlow.
type SplitBuilder[T any] struct {
	ctx          context.Context
	bufferSize   uint
	errorHandler func(error)
	predicate    func(T) (bool, error)

	in primitives.Outlet[T]
}

// Split creates a new SplitFlowBuilder for building a SplitFlow.
func Split[T any](
	in primitives.Outlet[T],
	predicate func(T) (bool, error),
) *SplitBuilder[T] {
	return &SplitBuilder[T]{
		in:        in,
		ctx:       context.Background(),
		predicate: predicate,
	}
}

// Context sets the context for the SplitFlow.
func (s *SplitBuilder[T]) Context(
	ctx context.Context,
) *SplitBuilder[T] {
	s.ctx = ctx
	return s
}

// BufferSize sets the buffer size for the SplitFlow.
func (s *SplitBuilder[T]) BufferSize(
	size uint,
) *SplitBuilder[T] {
	s.bufferSize = size
	return s
}

// ErrorHandler sets the error handler for the SplitFlow.
func (s *SplitBuilder[T]) ErrorHandler(
	handler func(error),
) *SplitBuilder[T] {
	s.errorHandler = handler
	return s
}

// Build creates and starts the SplitFlow.
func (s *SplitBuilder[T]) Build() *SplitFlow[T] {
	if s.predicate == nil {
		panic("SplitFlow requires a non-nil predicate function")
	}

	split := &SplitFlow[T]{
		ctx:          s.ctx,
		bufferSize:   s.bufferSize,
		errorHandler: s.errorHandler,
		predicate:    s.predicate,
		in:           s.in,
	}

	split.matching = PassThrough[T, T]().
		BufferSize(split.bufferSize).
		Context(split.ctx).
		Build()
	split.nonMatching = PassThrough[T, T]().
		BufferSize(split.bufferSize).
		Context(split.ctx).
		Build()

	go split.start()

	return split
}

// Matching returns the matching output flow.
func (s *SplitFlow[T]) Matching() primitives.Flow[T, T] {
	return s.matching
}

// NonMatching returns the non-matching output flow.
func (s *SplitFlow[T]) NonMatching() primitives.Flow[T, T] {
	return s.nonMatching
}

func (s *SplitFlow[T]) start() {
	defer close(s.matching.In())
	defer close(s.nonMatching.In())
	defer helpers.Drain(s.in.Out())

	for v := range s.in.Out() {
		// Check context before processing
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		passes, err := s.predicate(v)
		if err != nil {
			if s.errorHandler != nil {
				s.errorHandler(err)
			}
			return
		}

		if passes {
			select {
			case <-s.ctx.Done():
				return
			case s.matching.In() <- v:
			}
		} else {
			select {
			case <-s.ctx.Done():
				return
			case s.nonMatching.In() <- v:
			}
		}
	}
}
