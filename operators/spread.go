package operators

import (
	"context"

	"github.com/arielf-camacho/data-stream/primitives"
	"github.com/samber/lo"
)

// SpreadOperator is an operator that spreads the values from the input channel
// to the output channels.
//
// Graphically, the SpreadOperator looks like this:
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
//
// -- SpreadOperator with 3 outputs --
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
type SpreadOperator[T any] struct {
	ctx context.Context

	bufferSize uint

	in  primitives.Outlet[T]
	out []primitives.Flow[T, T]
}

// Spread creates a new SpreadBuilder for building a SpreadOperator.
func Spread[T any](
	in primitives.Outlet[T],
	out ...primitives.Flow[T, T],
) *SpreadBuilder[T] {
	return &SpreadBuilder[T]{
		ctx: context.Background(),

		in:  in,
		out: out,
	}
}

// SpreadBuilder is a fluent builder for SpreadOperator.
type SpreadBuilder[T any] struct {
	bufferSize uint
	ctx        context.Context

	in  primitives.Outlet[T]
	out []primitives.Flow[T, T]
}

// BufferSize sets the buffer size for the SpreadOperator channels.
func (b *SpreadBuilder[T]) BufferSize(size uint) *SpreadBuilder[T] {
	b.bufferSize = size
	return b
}

// Context sets the context for the SpreadOperator.
func (b *SpreadBuilder[T]) Context(ctx context.Context) *SpreadBuilder[T] {
	b.ctx = ctx
	return b
}

// Build creates and starts the SpreadOperator.
func (b *SpreadBuilder[T]) Build() *SpreadOperator[T] {
	if len(b.out) == 0 {
		panic("SpreadOperator requires at least one output")
	}

	operator := &SpreadOperator[T]{
		ctx:        b.ctx,
		bufferSize: b.bufferSize,

		in:  b.in,
		out: b.out,
	}

	go operator.start()

	return operator
}

func (s *SpreadOperator[T]) Outlets() []primitives.Outlet[T] {
	return lo.Map(
		s.out,
		func(out primitives.Flow[T, T], _ int) primitives.Outlet[T] { return out },
	)
}

func (s *SpreadOperator[T]) start() {
	defer func() {
		for _, out := range s.out {
			close(out.In())
		}
	}()

	for v := range s.in.Out() {
		for _, out := range s.out {
			select {
			case <-s.ctx.Done():
				return
			case out.In() <- v:
			}
		}
	}
}
