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

func NewMergeOperator(
	from []primitives.Out[any],
	opts ...MergeOperatorOption,
) *MergeOperator {
	merge := &MergeOperator{
		from: from,
		ctx:  context.Background(),
	}

	for _, opt := range opts {
		opt(merge)
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
