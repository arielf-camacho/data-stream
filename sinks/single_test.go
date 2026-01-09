package sinks_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/arielf-camacho/data-stream/sinks"
)

func TestSingleSink_In(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		expected int
		subject  func() (*sinks.SingleSink[int], context.Context)
	}{
		"receives-single-value": {
			expected: 42,
			subject: func() (*sinks.SingleSink[int], context.Context) {
				sink := sinks.Single[int]().Build()
				go func() {
					defer close(sink.In())
					sink.In() <- 42
				}()
				return sink, ctx
			},
		},
		"receives-zero-value": {
			expected: 0,
			subject: func() (*sinks.SingleSink[int], context.Context) {
				sink := sinks.Single[int]().Build()
				go func() {
					defer close(sink.In())
					sink.In() <- 0
				}()
				return sink, ctx
			},
		},
		"empty-input-returns-zero-value": {
			expected: 0,
			subject: func() (*sinks.SingleSink[int], context.Context) {
				sink := sinks.Single[int]().Build()
				go func() {
					defer close(sink.In())
				}()
				return sink, ctx
			},
		},
		"second-value-replaces-first": {
			expected: 100,
			subject: func() (*sinks.SingleSink[int], context.Context) {
				sink := sinks.Single[int]().Build()
				go func() {
					defer close(sink.In())
					sink.In() <- 42
					sink.In() <- 100
				}()
				return sink, ctx
			},
		},
		"cancelled-context-before-value": {
			expected: 0,
			subject: func() (*sinks.SingleSink[int], context.Context) {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				sink := sinks.Single[int]().Context(ctx).Build()
				go func() {
					defer close(sink.In())
					time.Sleep(100 * time.Millisecond)
					sink.In() <- 42
				}()
				return sink, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			sink, _ := c.subject()

			// When
			waitErr := sink.Wait()
			result := sink.Result()

			// Then
			assert.NoError(t, waitErr)
			assert.Equal(t, c.expected, result)
		})
	}
}

func TestSingleSink_Result(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		expected int
		subject  func() (*sinks.SingleSink[int], context.Context)
	}{
		"returns-received-value": {
			expected: 42,
			subject: func() (*sinks.SingleSink[int], context.Context) {
				sink := sinks.Single[int]().Build()
				go func() {
					sink.In() <- 42
					close(sink.In())
				}()
				return sink, ctx
			},
		},
		"returns-zero-value-when-empty": {
			expected: 0,
			subject: func() (*sinks.SingleSink[int], context.Context) {
				sink := sinks.Single[int]().Build()
				go func() {
					close(sink.In())
				}()
				return sink, ctx
			},
		},
		"thread-safe-concurrent-access": {
			expected: 42,
			subject: func() (*sinks.SingleSink[int], context.Context) {
				sink := sinks.Single[int]().Build()
				go func() {
					sink.In() <- 42
					close(sink.In())
				}()
				return sink, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			sink, _ := c.subject()

			// When
			time.Sleep(100 * time.Millisecond)

			if name == "thread-safe-concurrent-access" {
				var wg sync.WaitGroup
				results := make([]int, 10)
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						results[idx] = sink.Result()
					}(i)
				}
				wg.Wait()

				// Then
				for _, result := range results {
					assert.Equal(t, c.expected, result)
				}
			} else {
				result := sink.Result()

				// Then
				assert.Equal(t, c.expected, result)
			}
		})
	}
}
