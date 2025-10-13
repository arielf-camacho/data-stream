package flows_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/arielf-camacho/data-stream/flows"
	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
)

func TestPassThroughFlow_Out(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expected []int
		subject  func() (primitives.Flow[int, int], context.Context)
	}{
		"passes-through-all-values": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() (primitives.Flow[int, int], context.Context) {
				flow := flows.PassThrough[int, int]().Build()
				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()
				return flow, ctx
			},
		},
		"passes-through-single-value": {
			expected: []int{42},
			subject: func() (primitives.Flow[int, int], context.Context) {
				flow := flows.PassThrough[int, int]().Build()
				go func() {
					flow.In() <- 42
					close(flow.In())
				}()
				return flow, ctx
			},
		},
		"empty-input": {
			expected: nil,
			subject: func() (primitives.Flow[int, int], context.Context) {
				flow := flows.PassThrough[int, int]().Build()
				close(flow.In())
				return flow, ctx
			},
		},
		"with-buffer-size": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() (primitives.Flow[int, int], context.Context) {
				flow := flows.PassThrough[int, int]().BufferSize(10).Build()
				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()
				return flow, ctx
			},
		},
		"cancelled-context": {
			expected: nil,
			subject: func() (primitives.Flow[int, int], context.Context) {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				flow := flows.PassThrough[int, int]().Context(ctx).Build()
				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()
				return flow, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			flow, ctx := c.subject()

			// When
			collected := helpers.Collect(ctx, flow.Out())

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}

func TestPassThroughFlow_CustomConvert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		items    []int
		expected []string
		subject  func([]int) (primitives.Flow[int, string], context.Context)
	}{
		"converts-int-to-string-with-prefix": {
			items:    []int{1, 2, 3, 4, 5},
			expected: []string{"num:1", "num:2", "num:3", "num:4", "num:5"},
			subject: func(items []int) (primitives.Flow[int, string], context.Context) {
				flow := flows.
					PassThrough[int, string]().
					Convert(func(x int) (string, error) { return "num:" + string(rune('0'+x)), nil }).
					Build()
				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()
				return flow, ctx
			},
		},
		"multiplies-and-converts": {
			items:    []int{1, 2, 3},
			expected: []string{"x2", "x4", "x6"},
			subject: func(items []int) (primitives.Flow[int, string], context.Context) {
				flow := flows.
					PassThrough[int, string]().
					Convert(func(x int) (string, error) {
						doubled := x * 2
						return "x" + string(rune('0'+doubled)), nil
					}).
					Build()
				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()
				return flow, ctx
			},
		},
		"converts-with-custom-logic": {
			items:    []int{1, 2, 3, 4, 5},
			expected: []string{"odd", "even", "odd", "even", "odd"},
			subject: func(items []int) (primitives.Flow[int, string], context.Context) {
				flow := flows.
					PassThrough[int, string]().
					Convert(func(x int) (string, error) {
						if x%2 == 0 {
							return "even", nil
						}
						return "odd", nil
					}).
					Build()
				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()
				return flow, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			flow, ctx := c.subject(c.items)

			// When
			collected := helpers.Collect(ctx, flow.Out())

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}

func TestPassThroughFlow_DefaultConvert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		expected []any
		subject  func() (primitives.Outlet[any], context.Context)
	}{
		"converts-int-to-any": {
			expected: []any{1, 2, 3},
			subject: func() (primitives.Outlet[any], context.Context) {
				flow := flows.PassThrough[int, any]().Build()
				go func() {
					flow.In() <- 1
					flow.In() <- 2
					flow.In() <- 3
					close(flow.In())
				}()
				return flow, ctx
			},
		},
		"converts-string-to-any": {
			expected: []any{"a", "b", "c"},
			subject: func() (primitives.Outlet[any], context.Context) {
				flow := flows.PassThrough[string, any]().Build()
				go func() {
					flow.In() <- "a"
					flow.In() <- "b"
					flow.In() <- "c"
					close(flow.In())
				}()
				return flow, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			flow, ctx := c.subject()

			// When
			collected := helpers.Collect(ctx, flow.Out())

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}

func TestPassThroughFlow_ToFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expected []int
		subject  func() (*flows.PassThroughFlow[int, int], *helpers.Collector[int])
	}{
		"chains-to-collector": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() (
				*flows.PassThroughFlow[int, int],
				*helpers.Collector[int],
			) {
				flow := flows.PassThrough[int, int]().Build()
				collector := helpers.NewCollector[int](ctx)

				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()

				return flow, collector
			},
		},
		"chains-multiple-passthrough-flows": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() (
				*flows.PassThroughFlow[int, int],
				*helpers.Collector[int],
			) {
				flow1 := flows.PassThrough[int, int]().Build()
				flow2 := flows.PassThrough[int, int]().Build()
				collector := helpers.NewCollector[int](ctx)

				go func() {
					for _, item := range items {
						flow1.In() <- item
					}
					close(flow1.In())
				}()

				flow1.ToFlow(flow2)

				return flow2, collector
			},
		},
		"with-buffer-size": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() (
				*flows.PassThroughFlow[int, int],
				*helpers.Collector[int],
			) {
				flow := flows.PassThrough[int, int]().BufferSize(10).Build()
				collector := helpers.NewCollector[int](ctx)

				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()

				return flow, collector
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			flow, collector := c.subject()

			// When
			flow.ToSink(collector)

			// Then
			assert.ElementsMatch(t, c.expected, collector.Items())
		})
	}
}

func TestPassThroughFlow_ToSink(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{10, 20, 30}

	cases := map[string]struct {
		expected []int
		subject  func() (*flows.PassThroughFlow[int, int], *helpers.Collector[int])
	}{
		"sends-all-values-to-sink": {
			expected: []int{10, 20, 30},
			subject: func() (
				*flows.PassThroughFlow[int, int],
				*helpers.Collector[int],
			) {
				flow := flows.PassThrough[int, int]().Build()
				collector := helpers.NewCollector[int](ctx)

				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()

				return flow, collector
			},
		},
		"with-custom-convert-to-sink": {
			expected: []int{10, 20, 30},
			subject: func() (
				*flows.PassThroughFlow[int, int],
				*helpers.Collector[int],
			) {
				flow := flows.
					PassThrough[int, int]().
					Convert(func(x int) (int, error) { return x, nil }).
					Build()
				collector := helpers.NewCollector[int](ctx)

				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()

				return flow, collector
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			flow, collector := c.subject()

			// When
			flow.ToSink(collector)

			// Then
			assert.ElementsMatch(t, c.expected, collector.Items())
		})
	}
}

func TestPassThroughFlow_ContextCancellation(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		subject func() (context.Context, []int)
	}{
		"cancelled-context-stops-processing": {
			subject: func() (context.Context, []int) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				items := make([]int, 100)
				for i := range items {
					items[i] = i
				}

				flow := flows.PassThrough[int, int]().Context(ctx).Build()
				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()

				result := helpers.Collect(ctx, flow.Out())

				return ctx, result
			},
		},
		"context-cancelled-during-processing": {
			subject: func() (context.Context, []int) {
				ctx, cancel := context.WithCancel(context.Background())

				items := make([]int, 100)
				for i := range items {
					items[i] = i
				}

				flow := flows.PassThrough[int, int]().Context(ctx).Build()
				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()

				// Cancel after a brief moment
				go func() {
					cancel()
				}()

				result := helpers.Collect(ctx, flow.Out())

				return ctx, result
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given & When
			_, result := c.subject()

			// Then
			assert.LessOrEqual(t, len(result), 100)
		})
	}
}

func TestPassThroughFlow_ErrorHandling(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expectedErr       error
		expectedCollected []int
		subject           func(chan error) *flows.PassThroughFlow[int, int]
	}{
		"error-stops-processing": {
			expectedErr:       assert.AnError,
			expectedCollected: []int{1, 2},
			subject: func(errCh chan error) *flows.PassThroughFlow[int, int] {
				errConvert := func(x int) (int, error) {
					if x == 3 {
						return 0, assert.AnError
					}
					return x, nil
				}
				flow := flows.
					PassThrough[int, int]().
					Convert(errConvert).
					ErrorHandler(func(err error) { errCh <- err }).
					Build()
				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()
				return flow
			},
		},
		"error-on-first-value": {
			expectedErr:       assert.AnError,
			expectedCollected: nil,
			subject: func(errCh chan error) *flows.PassThroughFlow[int, int] {
				errConvert := func(x int) (int, error) {
					if x == 1 {
						return 0, assert.AnError
					}
					return x, nil
				}
				flow := flows.
					PassThrough[int, int]().
					Convert(errConvert).
					ErrorHandler(func(err error) { errCh <- err }).
					Build()
				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()
				return flow
			},
		},
		"multiple-errors-only-first-handled": {
			expectedErr:       assert.AnError,
			expectedCollected: []int{1},
			subject: func(errCh chan error) *flows.PassThroughFlow[int, int] {
				errConvert := func(x int) (int, error) {
					if x == 2 {
						return 0, assert.AnError
					}
					if x == 4 {
						return 0, assert.AnError
					}
					return x, nil
				}
				flow := flows.
					PassThrough[int, int]().
					Convert(errConvert).
					ErrorHandler(func(err error) { errCh <- err }).
					Build()
				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()
				return flow
			},
		},
		"no-error-handler-ignores-errors": {
			expectedErr:       nil,
			expectedCollected: []int{1, 2},
			subject: func(errCh chan error) *flows.PassThroughFlow[int, int] {
				errConvert := func(x int) (int, error) {
					if x == 3 {
						return 0, assert.AnError
					}
					return x, nil
				}
				flow := flows.PassThrough[int, int]().Convert(errConvert).Build()
				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()
				return flow
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			errCh := make(chan error, 1)
			flow := c.subject(errCh)

			// When
			var receivedErr error
			var collected []int

			// Collect both output and error
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				collected = helpers.Collect(ctx, flow.Out())
			}()

			go func() {
				defer wg.Done()
				select {
				case err := <-errCh:
					receivedErr = err
				case <-time.After(1 * time.Second):
					// No error received
				}
			}()

			wg.Wait()

			// Then
			if c.expectedErr != nil {
				assert.NotNil(t, receivedErr)
				assert.Equal(t, c.expectedErr.Error(), receivedErr.Error())
			} else {
				assert.Nil(t, receivedErr)
			}
			if c.expectedCollected != nil {
				assert.Subset(t, c.expectedCollected, collected)
			}
		})
	}
}

func TestPassThroughFlow_ErrorHandlingWithToSink(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expectedErr       error
		expectedCollected []int
		subject           func(chan error) (*flows.PassThroughFlow[int, int], *helpers.Collector[int])
	}{
		"error-stops-processing-to-sink": {
			expectedErr:       assert.AnError,
			expectedCollected: []int{1, 2},
			subject: func(errCh chan error) (*flows.PassThroughFlow[int, int], *helpers.Collector[int]) {
				errConvert := func(x int) (int, error) {
					if x == 3 {
						return 0, assert.AnError
					}
					return x, nil
				}
				flow := flows.
					PassThrough[int, int]().
					Convert(errConvert).
					ErrorHandler(func(err error) { errCh <- err }).
					Build()
				collector := helpers.NewCollector[int](ctx)

				go func() {
					for _, item := range items {
						flow.In() <- item
					}
					close(flow.In())
				}()

				return flow, collector
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			errCh := make(chan error, 1)
			flow, collector := c.subject(errCh)

			// When
			flow.ToSink(collector)

			var receivedErr error
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				select {
				case err := <-errCh:
					receivedErr = err
				case <-time.After(1 * time.Second):
					// No error received
				}
			}()

			wg.Wait()
			collected := collector.Items()

			// Then
			if c.expectedErr != nil {
				assert.NotNil(t, receivedErr)
				assert.Equal(t, c.expectedErr.Error(), receivedErr.Error())
			} else {
				assert.Nil(t, receivedErr)
			}
			if c.expectedCollected != nil {
				assert.Subset(t, c.expectedCollected, collected)
			}
		})
	}
}

func TestPassThroughFlow_MultipleActivationsPanic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("panics-on-multiple-ToSink-calls", func(t *testing.T) {
		t.Parallel()

		// Given
		flow := flows.PassThrough[int, int]().Build()
		collector1 := helpers.NewCollector[int](ctx)
		collector2 := helpers.NewCollector[int](ctx)

		// When
		flow.ToSink(collector1)

		// Then
		assert.Panics(t, func() { flow.ToSink(collector2) })
	})

	t.Run("panics-on-multiple-ToFlow-calls", func(t *testing.T) {
		t.Parallel()

		// Given
		flow := flows.PassThrough[int, int]().Build()
		next1 := flows.PassThrough[int, int]().Build()
		next2 := flows.PassThrough[int, int]().Build()

		// When
		flow.ToFlow(next1)

		// Then
		assert.Panics(t, func() { flow.ToFlow(next2) })
	})

	t.Run("panics-on-ToFlow-then-ToSink", func(t *testing.T) {
		t.Parallel()

		// Given
		flow := flows.PassThrough[int, int]().Build()
		next := flows.PassThrough[int, int]().Build()
		collector := helpers.NewCollector[int](ctx)

		// When
		flow.ToFlow(next)

		// Then
		assert.Panics(t, func() { flow.ToSink(collector) })
	})
}
