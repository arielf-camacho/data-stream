package operators

import (
	"context"
	"sync"

	"github.com/arielf-camacho/data-stream/primitives"
)

type MergeOperator struct {
	from []primitives.Out[any]

	ctx        context.Context
	bufferSize uint
	out        chan any
}

// MergeBuilder is a fluent builder for MergeOperator.
type MergeBuilder struct {
	from       []primitives.Out[any]
	ctx        context.Context
	bufferSize uint
}

// Merge creates a new MergeBuilder for building a MergeOperator.
func Merge(from []primitives.Out[any]) *MergeBuilder {
	return &MergeBuilder{
		from: from,
		ctx:  context.Background(),
	}
}

// Context sets the context for the MergeOperator.
func (b *MergeBuilder) Context(ctx context.Context) *MergeBuilder {
	b.ctx = ctx
	return b
}

// BufferSize sets the buffer size for the MergeOperator output
// channel.
func (b *MergeBuilder) BufferSize(size uint) *MergeBuilder {
	b.bufferSize = size
	return b
}

// Build creates and starts the MergeOperator.
func (b *MergeBuilder) Build() *MergeOperator {
	merge := &MergeOperator{
		from:       b.from,
		ctx:        b.ctx,
		bufferSize: b.bufferSize,
	}

	merge.out = make(chan any, merge.bufferSize)

	go merge.start()

	return merge
}

func (m *MergeOperator) To(in primitives.In[any]) {
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

func (m *MergeOperator) start() {
	defer close(m.out)

	var wg sync.WaitGroup

	for _, from := range m.from {
		wg.Add(1)
		go func(from primitives.Out[any]) {
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
