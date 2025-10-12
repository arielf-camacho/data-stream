package operators

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = (primitives.Operator[byte, int])(&MapOperator[int, byte]{})

// MapOperator is an operator that maps the values from the input channel to the
// output channel using the given transformation function.
//
// Graphically, the MapOperator looks like this:
//
// -- 1 -- 2 -- 3 -- 4 -- 5  -- | -->
//
// -- MapOperator f(x) = x*2 --
//
// -- 2 -- 4 -- 6 -- 8 -- 10 -- | -->
type MapOperator[IN any, OUT any] struct {
	ctx context.Context

	bufferSize   uint
	parallelism  uint
	errorHandler func(error)

	fn  func(IN) (OUT, error)
	in  chan IN
	out chan OUT
}

// NewMapOperator returns a new MapOperator given the transformation function.
func NewMapOperator[IN any, OUT any](
	fn func(IN) (OUT, error),
	opts ...MapOperatorOption[IN, OUT],
) *MapOperator[IN, OUT] {
	operator := &MapOperator[IN, OUT]{
		fn:          fn,
		parallelism: 1,
		ctx:         context.Background(),
	}

	for _, opt := range opts {
		opt(operator)
	}

	operator.in = make(chan IN, operator.bufferSize)
	operator.out = make(chan OUT, operator.bufferSize)

	go operator.start()

	return operator
}

func (m *MapOperator[IN, OUT]) In() chan<- IN {
	return m.in
}

func (m *MapOperator[IN, OUT]) Out() <-chan OUT {
	return m.out
}

func (m *MapOperator[IN, OUT]) To(in primitives.In[OUT]) {
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

func (m *MapOperator[IN, OUT]) start() {
	if m.parallelism <= 1 {
		m.sync()
	} else {
		m.async()
	}
}

func (m *MapOperator[IN, OUT]) sync() {
	defer close(m.out)

	for v := range m.in {
		select {
		case <-m.ctx.Done():
			return
		default:
			transformed, err := m.fn(v)
			if err != nil {
				if m.errorHandler != nil {
					m.errorHandler(err)
				}
				helpers.Drain(m.in)
				return
			}

			select {
			case <-m.ctx.Done():
				return
			case m.out <- transformed:
			}
		}
	}
}

func (m *MapOperator[IN, OUT]) async() {
	defer close(m.out)
	defer helpers.Drain(m.in)

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, m.parallelism)

	ctx, cancel := context.WithCancel(m.ctx)
	defer cancel()

	hasError := atomic.Bool{}

loop:
	for v := range m.in {
		select {
		case <-ctx.Done():
			break loop
		case semaphore <- struct{}{}:
			wg.Add(1)
			go func(v IN) {
				defer func() {
					<-semaphore
					wg.Done()
				}()

				transformed, err := m.fn(v)
				if err != nil {
					if !hasError.Swap(true) {
						cancel()
						if m.errorHandler != nil {
							m.errorHandler(err)
						}
					}
					return
				}

				select {
				case <-ctx.Done():
					return
				case m.out <- transformed:
				}
			}(v)
		}
	}

	wg.Wait()
}
