package flows

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = (primitives.Flow[byte, int])(&MapFlow[byte, int]{})

// MapFlow is an operator that maps the values from the input channel to the
// output channel using the given transformation function.
//
// Graphically, the MapFlow looks like this:
//
// -- 1 -- 2 -- 3 -- 4 -- 5  -- | -->
//
// -- MapFlow f(x) = x*2 --
//
// -- 2 -- 4 -- 6 -- 8 -- 10 -- | -->
type MapFlow[IN any, OUT any] struct {
	ctx context.Context

	activated    atomic.Bool
	bufferSize   uint
	parallelism  uint
	errorHandler func(error)

	fn  func(IN) (OUT, error)
	in  chan IN
	out chan OUT
}

// MapBuilder is a fluent builder for MapFlow.
type MapBuilder[IN any, OUT any] struct {
	fn           func(IN) (OUT, error)
	ctx          context.Context
	parallelism  uint
	errorHandler func(error)
	bufferSize   uint
}

// Map creates a new MapBuilder for building a MapFlow.
func Map[IN, OUT any](
	fn func(IN) (OUT, error),
) *MapBuilder[IN, OUT] {
	return &MapBuilder[IN, OUT]{
		fn:          fn,
		parallelism: 1,
		ctx:         context.Background(),
	}
}

// Context sets the context for the MapFlow.
func (b *MapBuilder[IN, OUT]) Context(
	ctx context.Context,
) *MapBuilder[IN, OUT] {
	b.ctx = ctx
	return b
}

// Parallelism sets the parallelism level for the MapFlow.
func (b *MapBuilder[IN, OUT]) Parallelism(
	p uint,
) *MapBuilder[IN, OUT] {
	b.parallelism = p
	return b
}

// ErrorHandler sets the error handler for the MapFlow.
func (b *MapBuilder[IN, OUT]) ErrorHandler(
	handler func(error),
) *MapBuilder[IN, OUT] {
	b.errorHandler = handler
	return b
}

// BufferSize sets the buffer size for the MapFlow channels.
func (b *MapBuilder[IN, OUT]) BufferSize(size uint) *MapBuilder[IN, OUT] {
	b.bufferSize = size
	return b
}

// Build creates and starts the MapFlow.
func (b *MapBuilder[IN, OUT]) Build() *MapFlow[IN, OUT] {
	operator := &MapFlow[IN, OUT]{
		fn:           b.fn,
		ctx:          b.ctx,
		parallelism:  b.parallelism,
		errorHandler: b.errorHandler,
		bufferSize:   b.bufferSize,
	}

	operator.in = make(chan IN, operator.bufferSize)
	operator.out = make(chan OUT, operator.bufferSize)

	go operator.start()

	return operator
}

func (m *MapFlow[IN, OUT]) In() chan<- IN {
	return m.in
}

func (m *MapFlow[IN, OUT]) Out() <-chan OUT {
	return m.out
}

func (m *MapFlow[IN, OUT]) ToFlow(
	in primitives.Flow[OUT, OUT],
) primitives.Flow[OUT, OUT] {
	m.assertNotActive()

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

	return in
}

func (m *MapFlow[IN, OUT]) ToSink(in primitives.Sink[OUT]) {
	m.assertNotActive()

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

func (m *MapFlow[IN, OUT]) assertNotActive() {
	if !m.activated.CompareAndSwap(false, true) {
		// TODO: Use a logger to print this error, don't panic
		panic("MapFlow is already streaming, cannot be used as a flow again")
	}
}

func (m *MapFlow[IN, OUT]) start() {
	if m.parallelism <= 1 {
		m.sync()
	} else {
		m.async()
	}
}

func (m *MapFlow[IN, OUT]) sync() {
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

func (m *MapFlow[IN, OUT]) async() {
	defer close(m.out)

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
