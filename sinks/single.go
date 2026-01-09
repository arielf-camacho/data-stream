package sinks

import (
	"context"
	"sync"

	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = primitives.Sink[int](&SingleSink[int]{})
var _ = primitives.WaitableSink[int](&SingleSink[int]{})

// SingleSink is a sink that gathers the expected SINGLE value from upstream.
// Immediately receives the value and closes the input channel, so, please, make
// sure that the upstream will emit ONLY ONE value up to this point. If the
// previous condition is not met, and a second value is emitted, that second
// value will be taken instead of the first one, and that's not what we want.
// This sink also implements the WaitableSink interface, so you can wait for the
// sink to finish receiving the value, so, until the upstream closes, this
// sink will still be waiting for the value, and the Wait() method will block
// until that value is received.
//
// Graphically, the SingleSink looks like this:
//
// -- 1 --------------------------- | -->
//
// -> 1 --------------------------- | -->
type SingleSink[T any] struct {
	ctx    context.Context
	mu     sync.RWMutex
	wg     sync.WaitGroup
	result T

	in chan T
}

// SingleSinkBuilder is a fluent builder for SingleSink.
type SingleSinkBuilder[T any] struct {
	ctx context.Context
}

// Single creates a new SingleSinkBuilder for building a SingleSink.
func Single[T any]() *SingleSinkBuilder[T] {
	return &SingleSinkBuilder[T]{
		ctx: context.Background(),
	}
}

// Context sets the context for the ReduceSink.
func (b *SingleSinkBuilder[T]) Context(
	ctx context.Context,
) *SingleSinkBuilder[T] {
	b.ctx = ctx
	return b
}

// Build creates and starts the ReduceSink.
func (b *SingleSinkBuilder[T]) Build() *SingleSink[T] {
	sink := &SingleSink[T]{
		ctx: b.ctx,
	}

	sink.in = make(chan T, 1)

	sink.wg.Add(1)
	go sink.start()

	return sink
}

// In returns the channel from which the values can be read.
func (s *SingleSink[T]) In() chan<- T {
	return s.in
}

// Wait waits for the SingleSink to finish receiving the value.
func (s *SingleSink[T]) Wait() error {
	s.wg.Wait()
	return nil
}

// Result returns the result (as of now) of the reduce operation.
func (s *SingleSink[T]) Result() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.result
}

func (s *SingleSink[T]) start() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case v, ok := <-s.in:
			if !ok {
				return
			}

			s.mu.Lock()
			s.result = v
			s.mu.Unlock()
		}
	}
}
