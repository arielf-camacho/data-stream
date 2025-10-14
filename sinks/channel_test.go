package sinks

import (
	"context"
	"sync"
	"testing"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/stretchr/testify/assert"
)

func TestChannelSink_In(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expected   []int
		subject    func() (*ChannelSink[int], chan int, context.Context)
		bufferSize uint
	}{
		"streams-all-values-to-channel": {
			expected:   []int{1, 2, 3, 4, 5},
			bufferSize: 1,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
		"streams-with-buffer": {
			expected:   []int{1, 2, 3, 4, 5},
			bufferSize: 10,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(10).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
		"streams-empty-input": {
			expected:   []int{},
			bufferSize: 1,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
		"streams-single-value": {
			expected:   []int{42},
			bufferSize: 1,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					sink.In() <- 42
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			_, out, _ := c.subject()

			// When
			var collected []int
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				for i := 0; i < len(c.expected); i++ {
					collected = append(collected, <-out)
				}
			}()

			wg.Wait()

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}

func TestChannelSink_ContextCancellation(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expectedCollected []int
		subject           func() (*ChannelSink[int], chan int, context.Context, context.CancelFunc)
	}{
		"cancelled-before-processing": {
			expectedCollected: []int{},
			subject: func() (*ChannelSink[int], chan int, context.Context, context.CancelFunc) {
				out := make(chan int, 10)
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				sink := Channel(out).Context(ctx).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx, cancel
			},
		},
		"cancelled-during-processing": {
			expectedCollected: []int{1, 2}, // Should collect some values before cancellation
			subject: func() (*ChannelSink[int], chan int, context.Context, context.CancelFunc) {
				out := make(chan int, 10)
				ctx, cancel := context.WithCancel(context.Background())
				sink := Channel(out).Context(ctx).BufferSize(1).Build()
				go func() {
					for i, item := range items {
						sink.In() <- item
						if i == 1 {
							cancel()
						}
					}
					close(sink.In())
				}()
				return sink, out, ctx, cancel
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			_, out, ctx, _ := c.subject()

			// When
			var collected []int
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				collected = helpers.Collect(ctx, out)
			}()
			wg.Wait()

			// Then
			if len(c.expectedCollected) > 0 {
				assert.Subset(t, c.expectedCollected, collected)
			} else {
				assert.Empty(t, collected)
			}
		})
	}
}

func TestChannelSink_BuilderPattern(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3}

	cases := map[string]struct {
		expected   []int
		subject    func() (*ChannelSink[int], chan int, context.Context)
		bufferSize uint
	}{
		"default-context": {
			expected:   []int{1, 2, 3},
			bufferSize: 1,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
		"custom-context": {
			expected:   []int{1, 2, 3},
			bufferSize: 5,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				customCtx := context.Background()
				out := make(chan int, 10)
				sink := Channel(out).Context(customCtx).BufferSize(5).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, customCtx
			},
		},
		"zero-buffer-size": {
			expected:   []int{1, 2, 3},
			bufferSize: 0,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(0).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			_, out, _ := c.subject()

			// When
			var collected []int
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				for i := 0; i < len(c.expected); i++ {
					collected = append(collected, <-out)
				}
			}()

			wg.Wait()

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}

func TestChannelSink_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	cases := map[string]struct {
		expected   []int
		subject    func() (*ChannelSink[int], chan int, context.Context)
		concurrent bool
	}{
		"sequential-processing": {
			expected:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			concurrent: false,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 20)
				sink := Channel(out).BufferSize(10).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
		"concurrent-sending": {
			expected:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			concurrent: true,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 20)
				sink := Channel(out).BufferSize(10).Build()
				var wg sync.WaitGroup
				for _, item := range items {
					wg.Add(1)
					go func(val int) {
						defer wg.Done()
						sink.In() <- val
					}(item)
				}
				go func() {
					wg.Wait()
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			_, out, _ := c.subject()

			// When
			var collected []int
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				for i := 0; i < len(c.expected); i++ {
					collected = append(collected, <-out)
				}
			}()

			wg.Wait()

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}

func TestChannelSink_TypeSafety(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		expected []string
		subject  func() (*ChannelSink[string], chan string, context.Context)
	}{
		"string-values": {
			expected: []string{"a", "b", "c"},
			subject: func() (*ChannelSink[string], chan string, context.Context) {
				out := make(chan string, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					sink.In() <- "a"
					sink.In() <- "b"
					sink.In() <- "c"
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			_, out, _ := c.subject()

			// When
			var collected []string
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				for i := 0; i < len(c.expected); i++ {
					collected = append(collected, <-out)
				}
			}()

			wg.Wait()

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}
