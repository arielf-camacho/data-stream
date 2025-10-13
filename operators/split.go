package operators

import (
	"context"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
)

// SplitOperator is an operator that splits the values from the input channel to
// the output channels based on the given predicate function.
//
// Graphically, the SplitOperator looks like this:
//
// -- 1 -- 2 -- 3 -- 4 -- 5 ------ | -->
//
// -- SplitOperator f(x) = x % 2 == 0 --
//
// Matching:
// ------- 2 ------- 4 ----------- | -->
//
// NonMatching:
// -- 1 ------- 3 ------- 5 ------ | -->
type SplitOperator[T any] struct {
	ctx          context.Context
	bufferSize   uint
	errorHandler func(error)
	predicate    func(T) (bool, error)

	in          primitives.Outlet[T]
	matching    primitives.Flow[T, T]
	nonMatching primitives.Flow[T, T]
}

// SplitOperatorBuilder is a fluent builder for SplitOperator.
type SplitOperatorBuilder[T any] struct {
	ctx          context.Context
	bufferSize   uint
	errorHandler func(error)
	predicate    func(T) (bool, error)

	in primitives.Outlet[T]
}

// Split creates a new SplitOperatorBuilder for building a SplitOperator.
func Split[T any](
	in primitives.Outlet[T],
	predicate func(T) (bool, error),
) *SplitOperatorBuilder[T] {
	return &SplitOperatorBuilder[T]{
		in:        in,
		ctx:       context.Background(),
		predicate: predicate,
	}
}

// Context sets the context for the SplitOperator.
func (s *SplitOperatorBuilder[T]) Context(
	ctx context.Context,
) *SplitOperatorBuilder[T] {
	s.ctx = ctx
	return s
}

// BufferSize sets the buffer size for the SplitOperator.
func (s *SplitOperatorBuilder[T]) BufferSize(
	size uint,
) *SplitOperatorBuilder[T] {
	s.bufferSize = size
	return s
}

// ErrorHandler sets the error handler for the SplitOperator.
func (s *SplitOperatorBuilder[T]) ErrorHandler(
	handler func(error),
) *SplitOperatorBuilder[T] {
	s.errorHandler = handler
	return s
}

// Build creates and starts the SplitOperator.
func (s *SplitOperatorBuilder[T]) Build() *SplitOperator[T] {
	if s.predicate == nil {
		panic("SplitOperator requires a non-nil predicate function")
	}

	split := &SplitOperator[T]{
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
func (s *SplitOperator[T]) Matching() primitives.Flow[T, T] {
	return s.matching
}

// NonMatching returns the non-matching output flow.
func (s *SplitOperator[T]) NonMatching() primitives.Flow[T, T] {
	return s.nonMatching
}

func (s *SplitOperator[T]) start() {
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
