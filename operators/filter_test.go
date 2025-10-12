package operators_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/operators"
	"github.com/arielf-camacho/data-stream/primitives"
)

func TestFilterOperator_Out(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	greaterThan5 := func(x int) (bool, error) { return x > 5, nil }

	cases := map[string]struct {
		expected []int
		subject  func() (primitives.Operator[int, int], context.Context)
	}{
		"filters-values-matching-predicate": {
			expected: []int{6, 7, 8, 9, 10},
			subject: func() (primitives.Operator[int, int], context.Context) {
				op := operators.NewFilterOperator(greaterThan5)
				go func() {
					for _, item := range items {
						op.In() <- item
					}
					close(op.In())
				}()
				return op, ctx
			},
		},
		"all-values-match-predicate": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() (primitives.Operator[int, int], context.Context) {
				alwaysTrue := func(x int) (bool, error) { return true, nil }
				op := operators.NewFilterOperator(alwaysTrue)
				go func() {
					for _, item := range []int{1, 2, 3, 4, 5} {
						op.In() <- item
					}
					close(op.In())
				}()
				return op, ctx
			},
		},
		"no-values-match-predicate": {
			expected: nil,
			subject: func() (primitives.Operator[int, int], context.Context) {
				alwaysFalse := func(x int) (bool, error) { return false, nil }
				op := operators.NewFilterOperator(alwaysFalse)
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
				op := operators.NewFilterOperator(greaterThan5)
				close(op.In())
				return op, ctx
			},
		},
		"even-numbers-only": {
			expected: []int{2, 4, 6, 8, 10},
			subject: func() (primitives.Operator[int, int], context.Context) {
				isEven := func(x int) (bool, error) { return x%2 == 0, nil }
				op := operators.NewFilterOperator(isEven)
				go func() {
					for _, item := range items {
						op.In() <- item
					}
					close(op.In())
				}()
				return op, ctx
			},
		},
		"cancelled-context": {
			expected: nil,
			subject: func() (primitives.Operator[int, int], context.Context) {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				op := operators.NewFilterOperator(
					greaterThan5,
					operators.WithContextForFilter[int](ctx),
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
		"with-buffer-size": {
			expected: []int{6, 7, 8, 9, 10},
			subject: func() (primitives.Operator[int, int], context.Context) {
				op := operators.NewFilterOperator(
					greaterThan5,
					operators.WithBufferSizeForFilter[int](5),
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

func TestFilterOperator_ErrorHandling(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	cases := map[string]struct {
		expectedErr       error
		expectedCollected []int
		subject           func(chan error) *operators.FilterOperator[int]
	}{
		"error-stops-processing": {
			expectedErr:       assert.AnError,
			expectedCollected: []int{4},
			subject: func(errCh chan error) *operators.FilterOperator[int] {
				errPredicate := func(x int) (bool, error) {
					if x == 5 {
						return false, assert.AnError
					}
					return x > 3, nil
				}
				op := operators.NewFilterOperator(
					errPredicate,
					operators.WithErrorHandlerForFilter[int](
						func(err error) {
							errCh <- err
						},
					),
				)
				go func() {
					for _, item := range items {
						op.In() <- item
					}
					close(op.In())
				}()
				return op
			},
		},
		"multiple-errors-only-first-handled": {
			expectedErr:       assert.AnError,
			expectedCollected: []int{1},
			subject: func(errCh chan error) *operators.FilterOperator[int] {
				errPredicate := func(x int) (bool, error) {
					if x == 2 {
						return false, assert.AnError
					}
					if x == 5 {
						return false, assert.AnError
					}
					return true, nil
				}
				op := operators.NewFilterOperator(
					errPredicate,
					operators.WithErrorHandlerForFilter[int](
						func(err error) {
							errCh <- err
						},
					),
				)
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
			operator := c.subject(errCh)

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

func TestFilterOperator_To(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	greaterThan5 := func(x int) (bool, error) {
		return x > 5, nil
	}

	cases := map[string]struct {
		expected []int
		subject  func() (*operators.FilterOperator[int], *helpers.Collector[int])
	}{
		"streams-filtered-values-to-collector": {
			expected: []int{6, 7, 8, 9, 10},
			subject: func() (
				*operators.FilterOperator[int],
				*helpers.Collector[int],
			) {
				op := operators.NewFilterOperator(greaterThan5)
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
		"odd-numbers-only": {
			expected: []int{1, 3, 5, 7, 9},
			subject: func() (
				*operators.FilterOperator[int],
				*helpers.Collector[int],
			) {
				isOdd := func(x int) (bool, error) {
					return x%2 != 0, nil
				}
				op := operators.NewFilterOperator(isOdd)
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
		"with-buffer-streams-all-matching-values": {
			expected: []int{6, 7, 8, 9, 10},
			subject: func() (
				*operators.FilterOperator[int],
				*helpers.Collector[int],
			) {
				op := operators.NewFilterOperator(
					greaterThan5,
					operators.WithBufferSizeForFilter[int](10),
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
			assert.ElementsMatch(t, c.expected,
				collector.Items())
		})
	}
}

func TestFilterOperator_ChainedFilters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	cases := map[string]struct {
		expected []int
		subject  func() <-chan int
	}{
		"chain-two-filters": {
			expected: []int{6, 8, 10},
			subject: func() <-chan int {
				// Filter 1: > 5
				filter1 := operators.NewFilterOperator(
					func(x int) (bool, error) {
						return x > 5, nil
					},
				)
				// Filter 2: even numbers
				filter2 := operators.NewFilterOperator(
					func(x int) (bool, error) {
						return x%2 == 0, nil
					},
				)

				go func() {
					for _, item := range items {
						filter1.In() <- item
					}
					close(filter1.In())
				}()

				go func() {
					for v := range filter1.Out() {
						filter2.In() <- v
					}
					close(filter2.In())
				}()

				return filter2.Out()
			},
		},
		"chain-three-filters": {
			expected: []int{6, 9},
			subject: func() <-chan int {
				// Filter 1: > 5
				filter1 := operators.NewFilterOperator(
					func(x int) (bool, error) {
						return x > 5, nil
					},
				)
				// Filter 2: < 10
				filter2 := operators.NewFilterOperator(
					func(x int) (bool, error) {
						return x < 10, nil
					},
				)
				// Filter 3: divisible by 3
				filter3 := operators.NewFilterOperator(
					func(x int) (bool, error) {
						return x%3 == 0, nil
					},
				)

				go func() {
					for _, item := range items {
						filter1.In() <- item
					}
					close(filter1.In())
				}()

				go func() {
					for v := range filter1.Out() {
						filter2.In() <- v
					}
					close(filter2.In())
				}()

				go func() {
					for v := range filter2.Out() {
						filter3.In() <- v
					}
					close(filter3.In())
				}()

				return filter3.Out()
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given & When
			out := c.subject()

			// Then
			collected := helpers.Collect(ctx, out)
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}
