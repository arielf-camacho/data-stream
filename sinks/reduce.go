package sinks

import (
	"context"
	"sync"

	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = primitives.Sink[int](&ReduceSink[int, any]{})
var _ = primitives.WaitableSink[int](&ReduceSink[int, any]{})

// ReduceSink is a sink that reduces the values from the input channel to a
// single value using the given reduce function.
//
// Graphically, the ReduceSink looks like this:
//
// -- 1 -- 2 -- 3 -- 4 -- 5 ------- | -->
//
// -- ReduceSink f(result, value, index) = result + value --
//
// -> ----------------------- 15 -- |
type ReduceSink[IN, OUT any] struct {
	ctx          context.Context
	bufferSize   uint
	mu           sync.RWMutex
	errorHandler func(error, uint, IN, OUT)
	fn           func(result OUT, value IN, index uint) (OUT, error)
	wg           sync.WaitGroup

	in     chan IN
	result OUT
}

// ReduceSinkBuilder is a fluent builder for ReduceSink.
type ReduceSinkBuilder[IN, OUT any] struct {
	fn           func(result OUT, value IN, index uint) (OUT, error)
	ctx          context.Context
	bufferSize   uint
	initial      OUT
	errorHandler func(error, uint, IN, OUT)
}

// Reduce creates a new ReduceSinkBuilder for building a ReduceSink.
func Reduce[IN, OUT any](
	fn func(result OUT, value IN, index uint) (OUT, error),
	initial OUT,
) *ReduceSinkBuilder[IN, OUT] {
	if fn == nil {
		panic("fn cannot be nil")
	}

	return &ReduceSinkBuilder[IN, OUT]{
		fn:      fn,
		initial: initial,
		ctx:     context.Background(),
	}
}

// Context sets the context for the ReduceSink.
func (b *ReduceSinkBuilder[IN, OUT]) Context(
	ctx context.Context,
) *ReduceSinkBuilder[IN, OUT] {
	b.ctx = ctx
	return b
}

// BufferSize sets the buffer size for the ReduceSink input channel.
func (b *ReduceSinkBuilder[IN, OUT]) BufferSize(
	size uint,
) *ReduceSinkBuilder[IN, OUT] {
	b.bufferSize = size
	return b
}

// ErrorHandler sets the error handler for the ReduceSink.
func (b *ReduceSinkBuilder[IN, OUT]) ErrorHandler(
	handler func(error, uint, IN, OUT),
) *ReduceSinkBuilder[IN, OUT] {
	b.errorHandler = handler
	return b
}

// Build creates and starts the ReduceSink.
func (b *ReduceSinkBuilder[IN, OUT]) Build() *ReduceSink[IN, OUT] {
	sink := &ReduceSink[IN, OUT]{
		fn:           b.fn,
		ctx:          b.ctx,
		bufferSize:   b.bufferSize,
		errorHandler: b.errorHandler,
		result:       b.initial,
	}

	sink.in = make(chan IN, sink.bufferSize)

	sink.wg.Add(1)
	go sink.start()

	return sink
}

// In returns the channel from which the values can be read.
func (s *ReduceSink[IN, OUT]) In() chan<- IN {
	return s.in
}

// Result returns the result (as of now) of the reduce operation.
func (s *ReduceSink[IN, OUT]) Result() OUT {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.result
}

// Wait waits for the ReduceSink to finish processing all values.
func (s *ReduceSink[IN, OUT]) Wait() error {
	s.wg.Wait()
	return nil
}

// start starts the ReduceSink.
func (s *ReduceSink[IN, OUT]) start() {
	defer s.wg.Done()

	index := uint(0)

	for {
		select {
		case <-s.ctx.Done():
			return
		case v, ok := <-s.in:
			if !ok {
				return
			}

			var err error
			s.result, err = s.accumulate(v, index)
			if err != nil {
				if s.errorHandler != nil {
					s.errorHandler(err, index, v, s.result)
				}
				return
			}

			index++
		}
	}
}

func (s *ReduceSink[IN, OUT]) accumulate(v IN, index uint) (OUT, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.fn(s.result, v, index)
}
