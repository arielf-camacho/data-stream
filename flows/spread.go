package flows

import (
	"context"

	"github.com/arielf-camacho/data-stream/primitives"
	"github.com/samber/lo"
)

// SpreadFlow is an operator that spreads the values from the input channel
// to the output channels. Spread, on every value, will send it to all the
// output channels sequentially. Be careful, if there is a slow consumer, it
// will block the spread operation. A workaround would be to connect to the
// Outlets, flows that have a buffer of reasonable size.
//
// Graphically, the SpreadFlow looks like this:
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
//
// -- SpreadFlow with 3 outputs --
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
type SpreadFlow[T any] struct {
	ctx context.Context

	in  primitives.Outlet[T]
	out []primitives.Flow[T, T]
}

// Spread creates a new SpreadBuilder for building a SpreadFlow.
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

// SpreadBuilder is a fluent builder for SpreadFlow.
type SpreadBuilder[T any] struct {
	ctx context.Context

	in  primitives.Outlet[T]
	out []primitives.Flow[T, T]
}

// Context sets the context for the SpreadFlow.
func (b *SpreadBuilder[T]) Context(ctx context.Context) *SpreadBuilder[T] {
	b.ctx = ctx
	return b
}

// Build creates and starts the SpreadFlow.
func (b *SpreadBuilder[T]) Build() *SpreadFlow[T] {
	if len(b.out) == 0 {
		panic("SpreadFlow requires at least one output")
	}

	operator := &SpreadFlow[T]{
		ctx: b.ctx,

		in:  b.in,
		out: b.out,
	}

	go operator.start()

	return operator
}

// Outlets returns the outlets of the SpreadFlow.
func (s *SpreadFlow[T]) Outlets() []primitives.Outlet[T] {
	return lo.Map(
		s.out,
		func(out primitives.Flow[T, T], _ int) primitives.Outlet[T] { return out },
	)
}

func (s *SpreadFlow[T]) start() {
	defer func() {
		for _, out := range s.out {
			close(out.In())
		}
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		case v, ok := <-s.in.Out():
			if !ok {
				return
			}
			for _, out := range s.out {
				select {
				case <-s.ctx.Done():
					return
				case out.In() <- v:
				}
			}
		}
	}
}
