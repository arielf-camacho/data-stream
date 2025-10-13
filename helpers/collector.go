package helpers

import (
	"context"
	"sync"

	"github.com/arielf-camacho/data-stream/primitives"
)

// Collector is a helper that collects the values from the given channel into a
// slice and returns it. If the context is done, the function returns the
// collected values so far.
type Collector[T any] struct {
	ctx    context.Context
	source chan T

	items []T
	wg    sync.WaitGroup
}

var _ = primitives.Sink[any](&Collector[any]{})

// NewCollector returns a new Collector given the context.
func NewCollector[T any](ctx context.Context) *Collector[T] {
	collector := &Collector[T]{
		ctx:    ctx,
		source: make(chan T),
	}

	collector.wg.Add(1)
	go collector.start()

	return collector
}

// In returns the channel from which the values can be read.
func (c *Collector[T]) In() chan<- T {
	return c.source
}

// Items returns the items collected by the Collector.
func (c *Collector[T]) Items() []T {
	c.wg.Wait()
	return c.items
}

func (c *Collector[T]) start() {
	defer Drain(c.source)
	defer c.wg.Done()

	for v := range c.source {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
		c.items = append(c.items, v)
	}
}
