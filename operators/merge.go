package operators

import (
	"context"
	"sync"

	"github.com/arielf-camacho/data-stream/primitives"
)

// MergeOperator is an operator that merges the values from the input channels
// to the output channel.
//
// Graphically, the MergeOperator looks like this:
//
// -- 1 ------- 3 ------- 5 -- | -->
//
// ------- 2 ------- 4 ------- | -->
//
// -- MergeOperator ---------- | -->
//
// -> 1 -- 2 -- 3 -- 4 -- 5 -- | -->
type MergeOperator[T any] struct {
	from []primitives.Outlet[T]

	ctx        context.Context
	bufferSize uint
	out        chan T
}

// MergeBuilder is a fluent builder for MergeOperator.
type MergeBuilder[T any] struct {
	from       []primitives.Outlet[T]
	ctx        context.Context
	bufferSize uint
}

// Merge creates a new MergeBuilder for building a MergeOperator.
func Merge[T any](from ...primitives.Outlet[T]) *MergeBuilder[T] {
	return &MergeBuilder[T]{
		from: from,
		ctx:  context.Background(),
	}
}

// Context sets the context for the MergeOperator.
func (b *MergeBuilder[T]) Context(ctx context.Context) *MergeBuilder[T] {
	b.ctx = ctx
	return b
}

// BufferSize sets the buffer size for the MergeOperator output
// channel.
func (b *MergeBuilder[T]) BufferSize(size uint) *MergeBuilder[T] {
	b.bufferSize = size
	return b
}

// Build creates and starts the MergeOperator.
func (b *MergeBuilder[T]) Build() *MergeOperator[T] {
	merge := &MergeOperator[T]{
		from:       b.from,
		ctx:        b.ctx,
		bufferSize: b.bufferSize,
	}

	merge.out = make(chan T, merge.bufferSize)

	go merge.start()

	return merge
}

func (m *MergeOperator[T]) To(in primitives.Inlet[T]) {
	go func() {
		defer close(in.In())
		for v := range m.out {
			select {
			case <-m.ctx.Done():
				return
			case in.In() <- v:
			}
		}
	}()
}

func (m *MergeOperator[T]) ToSink(in primitives.Sink[T]) {
	go func() {
		defer close(in.In())
		for v := range m.out {
			select {
			case <-m.ctx.Done():
				return
			case in.In() <- v:
			}
		}
	}()
}

func (m *MergeOperator[T]) start() {
	defer close(m.out)

	var wg sync.WaitGroup

	for _, from := range m.from {
		wg.Add(1)
		go func(from primitives.Outlet[T]) {
			defer wg.Done()

			for v := range from.Out() {
				select {
				case <-m.ctx.Done():
					return
				case m.out <- v:
				}
			}
		}(from)
	}

	wg.Wait()
}
