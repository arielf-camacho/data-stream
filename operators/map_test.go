package operators_test

import (
	"context"
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
	double := func(x int) int { return x * 2 }

	cases := map[string]struct {
		expected []int
		subject  func() (primitives.Operator[int, int], context.Context)
	}{
		"sync-mode-transforms-all-values": {
			expected: []int{2, 4, 6, 8, 10},
			subject: func() (primitives.Operator[int, int], context.Context) {
				op := operators.NewMapOperator(double)
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
				op := operators.NewMapOperator(
					double,
					operators.WithParallelism[int, int](3),
				)
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
				op := operators.NewMapOperator(double)
				close(op.In())
				return op, ctx
			},
		},
		"cancelled-context-sync-mode": {
			expected: nil,
			subject: func() (primitives.Operator[int, int], context.Context) {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				op := operators.NewMapOperator(
					double,
					operators.WithContext[int, int](ctx),
				)
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
				op := operators.NewMapOperator(
					double,
					operators.WithContext[int, int](ctx),
					operators.WithParallelism[int, int](3),
				)
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

func TestMapOperator_To(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}
	double := func(x int) int { return x * 2 }

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
				op := operators.NewMapOperator(double)
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
				op := operators.NewMapOperator(
					double,
					operators.WithParallelism[int, int](3),
				)
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
				op := operators.NewMapOperator(
					double,
					operators.WithParallelism[int, int](10),
				)
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
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	var concurrentCalls atomic.Int32
	var maxConcurrent atomic.Int32

	slowFunc := func(x int) int {
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
		return x * 2
	}

	cases := map[string]struct {
		parallelism        uint
		minConcurrentCalls int32
		expectedResults    []int
	}{
		"parallelism-1-runs-sequentially": {
			parallelism:        1,
			minConcurrentCalls: 1,
			expectedResults: []int{2, 4, 6, 8, 10, 12, 14, 16,
				18, 20},
		},
		"parallelism-3-runs-concurrently": {
			parallelism:        3,
			minConcurrentCalls: 2,
			expectedResults: []int{2, 4, 6, 8, 10, 12, 14, 16,
				18, 20},
		},
		"parallelism-5-runs-highly-concurrent": {
			parallelism:        5,
			minConcurrentCalls: 3,
			expectedResults: []int{2, 4, 6, 8, 10, 12, 14, 16,
				18, 20},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Reset counters
			concurrentCalls.Store(0)
			maxConcurrent.Store(0)

			// Given
			op := operators.NewMapOperator(
				slowFunc,
				operators.WithParallelism[int, int](c.parallelism),
			)

			go func() {
				for _, item := range items {
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
