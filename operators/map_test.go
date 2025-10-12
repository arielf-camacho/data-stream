package operators_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/operators"
	"github.com/arielf-camacho/data-stream/primitives"
)

func TestMapOperator_Out(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}
	double := func(x int) (int, error) { return x * 2, nil }

	cases := map[string]struct {
		expected []int
		subject  func() (primitives.Operator[int, int], context.Context)
	}{
		"sync-mode-transforms-all-values": {
			expected: []int{2, 4, 6, 8, 10},
			subject: func() (primitives.Operator[int, int], context.Context) {
				op := operators.Map(double).Build()
				go func() {
					for _, item := range items {
						op.In() <- item
					}
					close(op.In())
				}()
				return op, ctx
			},
		},
		"async-mode-transforms-all-values": {
			expected: []int{2, 4, 6, 8, 10},
			subject: func() (primitives.Operator[int, int], context.Context) {
				op := operators.Map(double).Parallelism(3).Build()
				go func() {
					for _, item := range items {
						op.In() <- item
					}
					close(op.In())
				}()
				return op, ctx
			},
		},
		"empty-input": {
			expected: nil,
			subject: func() (primitives.Operator[int, int], context.Context) {
				op := operators.Map(double).Build()
				close(op.In())
				return op, ctx
			},
		},
		"cancelled-context-sync-mode": {
			expected: nil,
			subject: func() (primitives.Operator[int, int], context.Context) {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				op := operators.Map(double).Context(ctx).Build()
				go func() {
					for _, item := range items {
						op.In() <- item
					}
					close(op.In())
				}()
				return op, ctx
			},
		},
		"cancelled-context-async-mode": {
			expected: nil,
			subject: func() (primitives.Operator[int, int], context.Context) {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				op := operators.Map(double).Context(ctx).Parallelism(3).Build()
				go func() {
					for _, item := range items {
						op.In() <- item
					}
					close(op.In())
				}()
				return op, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			operator, ctx := c.subject()

			// When
			collected := helpers.Collect(ctx, operator.Out())

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}

func TestMapOperator_ErrorHandling(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	cases := map[string]struct {
		expectedErr       error
		expectedCollected []int
		parallelism       uint
		subject           func(uint, chan error) *operators.MapOperator[int, int]
	}{
		"error-stops-processing-sync-mode": {
			expectedErr:       assert.AnError,
			expectedCollected: []int{2, 4, 6, 8},
			parallelism:       1,
			subject: func(p uint, errCh chan error) *operators.MapOperator[int, int] {
				errTransform := func(x int) (int, error) {
					if x == 5 {
						return 0, assert.AnError
					}
					return x * 2, nil
				}
				op := operators.
					Map(errTransform).
					Parallelism(p).
					ErrorHandler(func(err error) { errCh <- err }).
					Build()
				go func() {
					for _, item := range items {
						op.In() <- item
					}
					close(op.In())
				}()
				return op
			},
		},
		"error-stops-processing-async-mode": {
			expectedErr:       assert.AnError,
			expectedCollected: nil,
			parallelism:       3,
			subject: func(p uint, errCh chan error) *operators.MapOperator[int, int] {
				errTransform := func(x int) (int, error) {
					if x == 5 {
						return 0, assert.AnError
					}
					return x * 2, nil
				}
				op := operators.
					Map(errTransform).
					Parallelism(p).
					ErrorHandler(func(err error) { errCh <- err }).
					Build()
				go func() {
					for _, item := range items {
						op.In() <- item
					}
					close(op.In())
				}()
				return op
			},
		},
		"multiple-errors-only-first-handled-sync": {
			expectedErr:       assert.AnError,
			expectedCollected: []int{2},
			parallelism:       1,
			subject: func(p uint, errCh chan error) *operators.MapOperator[int, int] {
				errTransform := func(x int) (int, error) {
					if x == 2 {
						return 0, assert.AnError
					}
					if x == 5 {
						return 0, assert.AnError
					}
					return x * 2, nil
				}
				op := operators.
					Map(errTransform).
					Parallelism(p).
					ErrorHandler(func(err error) { errCh <- err }).
					Build()
				go func() {
					for _, item := range items {
						op.In() <- item
					}
					close(op.In())
				}()
				return op
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			errCh := make(chan error, 1)
			operator := c.subject(c.parallelism, errCh)

			// When
			var receivedErr error
			var collected []int

			// Collect both output and error
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				collected = helpers.Collect(ctx, operator.Out())
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
			assert.NotNil(t, receivedErr)
			assert.Equal(t, c.expectedErr.Error(), receivedErr.Error())
			if c.expectedCollected != nil {
				assert.Subset(t, c.expectedCollected, collected)
			}
		})
	}
}

func TestMapOperator_To(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}
	double := func(x int) (int, error) { return x * 2, nil }

	cases := map[string]struct {
		expected []int
		subject  func() (*operators.MapOperator[int, int], *helpers.Collector[int])
	}{
		"sync-mode-streams-all-values-to-collector": {
			expected: []int{2, 4, 6, 8, 10},
			subject: func() (
				*operators.MapOperator[int, int],
				*helpers.Collector[int],
			) {
				op := operators.Map(double).Build()
				collector := helpers.NewCollector[int](ctx)

				go func() {
					for _, item := range items {
						op.In() <- item
					}
					close(op.In())
				}()

				return op, collector
			},
		},
		"async-mode-streams-all-values-to-collector": {
			expected: []int{2, 4, 6, 8, 10},
			subject: func() (
				*operators.MapOperator[int, int],
				*helpers.Collector[int],
			) {
				op := operators.Map(double).Parallelism(3).Build()
				collector := helpers.NewCollector[int](ctx)

				go func() {
					for _, item := range items {
						op.In() <- item
					}
					close(op.In())
				}()

				return op, collector
			},
		},
		"high-parallelism-preserves-all-values": {
			expected: []int{2, 4, 6, 8, 10},
			subject: func() (
				*operators.MapOperator[int, int],
				*helpers.Collector[int],
			) {
				op := operators.Map(double).Parallelism(10).Build()
				collector := helpers.NewCollector[int](ctx)

				go func() {
					for _, item := range items {
						op.In() <- item
					}
					close(op.In())
				}()

				return op, collector
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			operator, collector := c.subject()

			// When
			operator.To(collector)

			// Then
			assert.ElementsMatch(t, c.expected, collector.Items())
		})
	}
}

func TestMapOperator_ParallelExecution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		parallelism        uint
		items              []int
		minConcurrentCalls int32
		expectedResults    []int
	}{
		"parallelism-1-runs-sequentially": {
			parallelism:        1,
			items:              []int{1, 2, 3, 4, 5},
			minConcurrentCalls: 1,
			expectedResults:    []int{2, 4, 6, 8, 10},
		},
		"parallelism-3-runs-concurrently": {
			parallelism:        3,
			items:              []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			minConcurrentCalls: 2,
			expectedResults:    []int{2, 4, 6, 8, 10, 12, 14, 16, 18},
		},
		"parallelism-5-runs-highly-concurrent": {
			parallelism:        5,
			items:              []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			minConcurrentCalls: 3,
			expectedResults:    []int{2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Each test case has its own counters
			var concurrentCalls atomic.Int32
			var maxConcurrent atomic.Int32

			slowFunc := func(x int) (int, error) {
				current := concurrentCalls.Add(1)
				defer concurrentCalls.Add(-1)

				// Track max concurrent executions
				for {
					max := maxConcurrent.Load()
					if current <= max {
						break
					}
					if maxConcurrent.CompareAndSwap(max, current) {
						break
					}
				}

				time.Sleep(10 * time.Millisecond)
				return x * 2, nil
			}

			// Given
			op := operators.Map(slowFunc).Parallelism(c.parallelism).Build()

			go func() {
				for _, item := range c.items {
					op.In() <- item
				}
				close(op.In())
			}()

			// When
			collected := helpers.Collect(ctx, op.Out())

			// Then
			assert.ElementsMatch(t, c.expectedResults, collected)
			assert.GreaterOrEqual(
				t,
				maxConcurrent.Load(),
				c.minConcurrentCalls,
				"Expected at least %d concurrent calls",
				c.minConcurrentCalls,
			)
		})
	}
}
