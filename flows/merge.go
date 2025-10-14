package flows

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = (primitives.Outlet[any])(&MergeFlow[any]{})

// MergeFlow is an operator that merges the values from the input channels
// to the output channel.
//
// Graphically, the MergeFlow looks like this:
//
// -- 1 ------- 3 ------- 5 -- | -->
//
// ------- 2 ------- 4 ------- | -->
//
// -- MergeFlow ---------- | -->
//
// -> 1 -- 2 -- 3 -- 4 -- 5 -- | -->
type MergeFlow[T any] struct {
	from []primitives.Outlet[T]

	ctx        context.Context
	bufferSize uint
	activated  atomic.Bool

	out chan T
}

// MergeBuilder is a fluent builder for MergeFlow.
type MergeBuilder[T any] struct {
	from       []primitives.Outlet[T]
	ctx        context.Context
	bufferSize uint
}

// Merge creates a new MergeBuilder for building a MergeFlow.
func Merge[T any](from ...primitives.Outlet[T]) *MergeBuilder[T] {
	return &MergeBuilder[T]{
		from: from,
		ctx:  context.Background(),
	}
}

// Context sets the context for the MergeFlow.
func (b *MergeBuilder[T]) Context(ctx context.Context) *MergeBuilder[T] {
	b.ctx = ctx
	return b
}

// BufferSize sets the buffer size for the MergeFlow output
// channel.
func (b *MergeBuilder[T]) BufferSize(size uint) *MergeBuilder[T] {
	b.bufferSize = size
	return b
}

// Build creates and starts the MergeFlow.
func (b *MergeBuilder[T]) Build() *MergeFlow[T] {
	merge := &MergeFlow[T]{
		from:       b.from,
		ctx:        b.ctx,
		bufferSize: b.bufferSize,
	}

	merge.out = make(chan T, merge.bufferSize)

	go merge.start()

	return merge
}

func (m *MergeFlow[T]) Out() <-chan T {
	return m.out
}

func (m *MergeFlow[T]) ToFlow(in primitives.Flow[T, T]) primitives.Flow[T, T] {
	m.assertNotActive()

	go func() {
		defer close(in.In())
		for {
			select {
			case <-m.ctx.Done():
				return
			case v, ok := <-m.out:
				if !ok {
					return
				}
				select {
				case <-m.ctx.Done():
					return
				case in.In() <- v:
				}
			}
		}
	}()

	return in
}

func (m *MergeFlow[T]) ToSink(in primitives.Sink[T]) {
	m.assertNotActive()

	go func() {
		defer close(in.In())
		for {
			select {
			case <-m.ctx.Done():
				return
			case v, ok := <-m.out:
				if !ok {
					return
				}
				select {
				case <-m.ctx.Done():
					return
				case in.In() <- v:
				}
			}
		}
	}()
}

func (m *MergeFlow[T]) assertNotActive() {
	if !m.activated.CompareAndSwap(false, true) {
		// TODO: Use a logger to print this error, don't panic
		panic("MergeFlow is already streaming, cannot be used as a flow again")
	}
}

func (m *MergeFlow[T]) start() {
	defer close(m.out)

	var wg sync.WaitGroup

	for _, from := range m.from {
		wg.Add(1)
		go func(from primitives.Outlet[T]) {
			defer wg.Done()

			for {
				select {
				case <-m.ctx.Done():
					return
				case v, ok := <-from.Out():
					if !ok {
						return
					}
					select {
					case <-m.ctx.Done():
						return
					case m.out <- v:
					}
				}
			}
		}(from)
	}

	wg.Wait()
}
