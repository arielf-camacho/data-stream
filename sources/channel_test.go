package sources_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arielf-camacho/data-stream/flows"
	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
	"github.com/arielf-camacho/data-stream/sources"
)

func TestChannelSource_Out(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		expected []int
		subject  func() primitives.Source[int]
	}{
		"emits-multiple-values": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() primitives.Source[int] {
				ch := make(chan int, 5)
				go func() {
					defer close(ch)
					for i := 1; i <= 5; i++ {
						ch <- i
					}
				}()
				return sources.Channel(ch).Build()
			},
		},
		"emits-single-value": {
			expected: []int{42},
			subject: func() primitives.Source[int] {
				ch := make(chan int, 1)
				go func() {
					defer close(ch)
					ch <- 42
				}()
				return sources.Channel(ch).Build()
			},
		},
		"empty-channel": {
			expected: nil,
			subject: func() primitives.Source[int] {
				ch := make(chan int)
				go func() {
					defer close(ch)
				}()
				return sources.Channel(ch).Build()
			},
		},
		"cancelled-context": {
			expected: nil,
			subject: func() primitives.Source[int] {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				ch := make(chan int, 5)
				go func() {
					defer close(ch)
					for i := 1; i <= 5; i++ {
						ch <- i
					}
				}()
				return sources.Channel(ch).Context(ctx).Build()
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			source := c.subject()

			// When
			collected := helpers.Collect(ctx, source.Out())

			// Then
			assert.Equal(t, c.expected, collected)
		})
	}
}

func TestChannelSource_ToFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		expected []int
		subject  func() *sources.ChannelSource[int]
	}{
		"streams-all-values-to-collector": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() *sources.ChannelSource[int] {
				ch := make(chan int, 5)
				go func() {
					defer close(ch)
					for i := 1; i <= 5; i++ {
						ch <- i
					}
				}()
				return sources.Channel(ch).Build()
			},
		},
		"streams-single-value": {
			expected: []int{42},
			subject: func() *sources.ChannelSource[int] {
				ch := make(chan int, 1)
				go func() {
					defer close(ch)
					ch <- 42
				}()
				return sources.Channel(ch).Build()
			},
		},
		"empty-channel": {
			expected: nil,
			subject: func() *sources.ChannelSource[int] {
				ch := make(chan int)
				go func() {
					defer close(ch)
				}()
				return sources.Channel(ch).Build()
			},
		},
		"cancelled-context": {
			expected: nil,
			subject: func() *sources.ChannelSource[int] {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				ch := make(chan int, 5)
				go func() {
					defer close(ch)
					for i := 1; i <= 5; i++ {
						ch <- i
					}
				}()
				return sources.Channel(ch).Context(ctx).Build()
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			source := c.subject()
			collector := helpers.NewCollector[int](ctx)
			flow := flows.PassThrough[int, int]().Context(ctx).Build()

			// When
			source.ToFlow(flow).ToSink(collector)

			// Then
			assert.Equal(t, c.expected, collector.Items())
		})
	}

	t.Run("panics-on-multiple-ToFlow-calls", func(t *testing.T) {
		t.Parallel()

		// Given
		ch := make(chan int, 1)
		go func() {
			defer close(ch)
			ch <- 42
		}()
		source := sources.Channel(ch).Build()
		collector1 := helpers.NewCollector[int](ctx)
		collector2 := helpers.NewCollector[int](ctx)

		// When
		source.ToSink(collector1)

		// Then
		assert.Panics(t, func() { source.ToSink(collector2) })
	})
}

func TestChannelSource_ToSink(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		expected []int
		subject  func() *sources.ChannelSource[int]
	}{
		"streams-all-values-to-sink": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() *sources.ChannelSource[int] {
				ch := make(chan int, 5)
				go func() {
					for i := 1; i <= 5; i++ {
						ch <- i
					}
					close(ch)
				}()
				return sources.Channel(ch).Build()
			},
		},
		"streams-single-value": {
			expected: []int{42},
			subject: func() *sources.ChannelSource[int] {
				ch := make(chan int, 1)
				go func() {
					ch <- 42
					close(ch)
				}()
				return sources.Channel(ch).Build()
			},
		},
		"empty-channel": {
			expected: nil,
			subject: func() *sources.ChannelSource[int] {
				ch := make(chan int)
				go func() {
					close(ch)
				}()
				return sources.Channel(ch).Build()
			},
		},
		"cancelled-context": {
			expected: nil,
			subject: func() *sources.ChannelSource[int] {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				ch := make(chan int, 5)
				go func() {
					for i := 1; i <= 5; i++ {
						ch <- i
					}
					close(ch)
				}()
				return sources.Channel(ch).Context(ctx).Build()
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			source := c.subject()
			collector := helpers.NewCollector[int](ctx)

			// When
			source.ToSink(collector)

			// Then
			assert.Equal(t, c.expected, collector.Items())
		})
	}

	t.Run("panics-on-multiple-ToSink-calls", func(t *testing.T) {
		t.Parallel()

		// Given
		ch := make(chan int, 1)
		go func() {
			ch <- 42
			close(ch)
		}()
		source := sources.Channel(ch).Build()
		collector1 := helpers.NewCollector[int](ctx)
		collector2 := helpers.NewCollector[int](ctx)

		// When
		source.ToSink(collector1)

		// Then
		assert.Panics(t, func() { source.ToSink(collector2) })
	})
}
