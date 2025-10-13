package operators_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/operators"
	"github.com/arielf-camacho/data-stream/sources"
)

func TestSplitOperator_Matching(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	isEven := func(x int) (bool, error) { return x%2 == 0, nil }

	cases := map[string]struct {
		expected []int
		subject  func() *operators.SplitOperator[int]
	}{
		"splits-even-numbers-to-matching": {
			expected: []int{2, 4, 6, 8, 10},
			subject: func() *operators.SplitOperator[int] {
				source := sources.Slice(items).Build()
				return operators.Split(source, isEven).Build()
			},
		},
		"all-values-match-predicate": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() *operators.SplitOperator[int] {
				alwaysTrue := func(x int) (bool, error) { return true, nil }
				source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()
				return operators.Split(source, alwaysTrue).Build()
			},
		},
		"no-values-match-predicate": {
			expected: nil,
			subject: func() *operators.SplitOperator[int] {
				alwaysFalse := func(x int) (bool, error) { return false, nil }
				source := sources.Slice(items).Build()
				return operators.Split(source, alwaysFalse).Build()
			},
		},
		"empty-input": {
			expected: nil,
			subject: func() *operators.SplitOperator[int] {
				source := sources.Slice([]int{}).Build()
				return operators.Split(source, isEven).Build()
			},
		},
		"greater-than-five-matches": {
			expected: []int{6, 7, 8, 9, 10},
			subject: func() *operators.SplitOperator[int] {
				greaterThan5 := func(x int) (bool, error) { return x > 5, nil }
				source := sources.Slice(items).Build()
				return operators.Split(source, greaterThan5).Build()
			},
		},
		"with-buffer-size": {
			expected: []int{2, 4, 6, 8, 10},
			subject: func() *operators.SplitOperator[int] {
				source := sources.Slice(items).Build()
				return operators.Split(source, isEven).BufferSize(5).Build()
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			split := c.subject()

			// When
			var matching []int
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				matching = helpers.Collect(ctx, split.Matching().Out())
			}()

			go func() {
				defer wg.Done()
				helpers.Drain(split.NonMatching().Out())
			}()

			wg.Wait()

			// Then
			assert.ElementsMatch(t, c.expected, matching)
		})
	}
}

func TestSplitOperator_NonMatching(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	isEven := func(x int) (bool, error) { return x%2 == 0, nil }

	cases := map[string]struct {
		expected []int
		subject  func() *operators.SplitOperator[int]
	}{
		"splits-odd-numbers-to-non-matching": {
			expected: []int{1, 3, 5, 7, 9},
			subject: func() *operators.SplitOperator[int] {
				source := sources.Slice(items).Build()
				return operators.Split(source, isEven).Build()
			},
		},
		"all-values-non-match-predicate": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() *operators.SplitOperator[int] {
				alwaysFalse := func(x int) (bool, error) { return false, nil }
				source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()
				return operators.Split(source, alwaysFalse).Build()
			},
		},
		"no-values-non-match-predicate": {
			expected: nil,
			subject: func() *operators.SplitOperator[int] {
				alwaysTrue := func(x int) (bool, error) { return true, nil }
				source := sources.Slice(items).Build()
				return operators.Split(source, alwaysTrue).Build()
			},
		},
		"empty-input": {
			expected: nil,
			subject: func() *operators.SplitOperator[int] {
				source := sources.Slice([]int{}).Build()
				return operators.Split(source, isEven).Build()
			},
		},
		"less-than-or-equal-five-non-matching": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() *operators.SplitOperator[int] {
				greaterThan5 := func(x int) (bool, error) { return x > 5, nil }
				source := sources.Slice(items).Build()
				return operators.Split(source, greaterThan5).Build()
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			split := c.subject()

			// When
			var nonMatching []int
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				helpers.Drain(split.Matching().Out())
			}()

			go func() {
				defer wg.Done()
				nonMatching = helpers.Collect(ctx, split.NonMatching().Out())
			}()

			wg.Wait()

			// Then
			assert.ElementsMatch(t, c.expected, nonMatching)
		})
	}
}

func TestSplitOperator_BothOutputs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	cases := map[string]struct {
		expectedMatching    []int
		expectedNonMatching []int
		subject             func() *operators.SplitOperator[int]
	}{
		"splits-even-and-odd": {
			expectedMatching:    []int{2, 4, 6, 8, 10},
			expectedNonMatching: []int{1, 3, 5, 7, 9},
			subject: func() *operators.SplitOperator[int] {
				isEven := func(x int) (bool, error) { return x%2 == 0, nil }
				source := sources.Slice(items).Build()
				return operators.Split(source, isEven).Build()
			},
		},
		"splits-by-range": {
			expectedMatching:    []int{6, 7, 8, 9, 10},
			expectedNonMatching: []int{1, 2, 3, 4, 5},
			subject: func() *operators.SplitOperator[int] {
				greaterThan5 := func(x int) (bool, error) { return x > 5, nil }
				source := sources.Slice(items).Build()
				return operators.Split(source, greaterThan5).Build()
			},
		},
		"splits-by-divisibility": {
			expectedMatching:    []int{3, 6, 9},
			expectedNonMatching: []int{1, 2, 4, 5, 7, 8, 10},
			subject: func() *operators.SplitOperator[int] {
				divisibleBy3 := func(x int) (bool, error) { return x%3 == 0, nil }
				source := sources.Slice(items).Build()
				return operators.Split(source, divisibleBy3).Build()
			},
		},
		"empty-input-both-empty": {
			expectedMatching:    nil,
			expectedNonMatching: nil,
			subject: func() *operators.SplitOperator[int] {
				isEven := func(x int) (bool, error) { return x%2 == 0, nil }
				source := sources.Slice([]int{}).Build()
				return operators.Split(source, isEven).Build()
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			split := c.subject()

			// When
			var matching, nonMatching []int
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				matching = helpers.Collect(ctx, split.Matching().Out())
			}()

			go func() {
				defer wg.Done()
				nonMatching = helpers.Collect(ctx, split.NonMatching().Out())
			}()

			wg.Wait()

			// Then
			assert.ElementsMatch(t, c.expectedMatching, matching)
			assert.ElementsMatch(t, c.expectedNonMatching, nonMatching)
		})
	}
}

func TestSplitOperator_ErrorHandling(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	cases := map[string]struct {
		expectedErr         error
		expectedMatching    []int
		expectedNonMatching []int
		subject             func(
			chan error,
		) *operators.SplitOperator[int]
	}{
		"error-stops-processing-both-outputs": {
			expectedErr:         assert.AnError,
			expectedMatching:    []int{2, 4},
			expectedNonMatching: []int{1, 3},
			subject: func(errCh chan error) *operators.SplitOperator[int] {
				errPredicate := func(x int) (bool, error) {
					if x == 5 {
						return false, assert.AnError
					}
					return x%2 == 0, nil
				}
				source := sources.Slice(items).Build()
				return operators.
					Split(source, errPredicate).
					ErrorHandler(func(err error) { errCh <- err }).
					Build()
			},
		},
		"error-on-first-value": {
			expectedErr:         assert.AnError,
			expectedMatching:    nil,
			expectedNonMatching: nil,
			subject: func(errCh chan error) *operators.SplitOperator[int] {
				errPredicate := func(x int) (bool, error) {
					if x == 1 {
						return false, assert.AnError
					}
					return x%2 == 0, nil
				}
				source := sources.Slice(items).Build()
				return operators.
					Split(source, errPredicate).
					ErrorHandler(func(err error) { errCh <- err }).
					Build()
			},
		},
		"error-without-handler-does-not-panic": {
			expectedErr:         nil,
			expectedMatching:    nil,
			expectedNonMatching: nil,
			subject: func(errCh chan error) *operators.SplitOperator[int] {
				errPredicate := func(x int) (bool, error) {
					return false, assert.AnError
				}
				source := sources.Slice(items).Build()
				return operators.Split(source, errPredicate).Build()
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			errCh := make(chan error, 1)
			split := c.subject(errCh)

			// When
			var receivedErr error
			var matching, nonMatching []int

			var wg sync.WaitGroup
			wg.Add(3)

			go func() {
				defer wg.Done()
				matching = helpers.Collect(ctx, split.Matching().Out())
			}()

			go func() {
				defer wg.Done()
				nonMatching = helpers.Collect(ctx, split.NonMatching().Out())
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
			if c.expectedMatching != nil {
				assert.Subset(t, c.expectedMatching, matching)
			}
			if c.expectedNonMatching != nil {
				assert.Subset(t, c.expectedNonMatching, nonMatching)
			}
		})
	}
}

func TestSplitOperator_ContextCancellation(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	isEven := func(x int) (bool, error) { return x%2 == 0, nil }

	cases := map[string]struct {
		expectedMatching    []int
		expectedNonMatching []int
		subject             func() (
			*operators.SplitOperator[int],
			context.Context,
		)
	}{
		"cancelled-context-stops-processing": {
			expectedMatching:    nil,
			expectedNonMatching: nil,
			subject: func() (
				*operators.SplitOperator[int],
				context.Context,
			) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				source := sources.Slice(items).Context(ctx).Build()
				split := operators.Split(source, isEven).Context(ctx).Build()
				return split, ctx
			},
		},
		"context-cancelled-during-processing": {
			expectedMatching:    nil,
			expectedNonMatching: nil,
			subject: func() (
				*operators.SplitOperator[int],
				context.Context,
			) {
				ctx, cancel := context.WithTimeout(
					context.Background(),
					1*time.Millisecond,
				)
				defer cancel()
				// Use large slice to ensure cancellation occurs
				largeSlice := make([]int, 10000)
				for i := range largeSlice {
					largeSlice[i] = i
				}
				source := sources.Slice(largeSlice).Context(ctx).Build()
				split := operators.Split(source, isEven).Context(ctx).Build()
				return split, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			split, ctx := c.subject()

			// When
			var matching, nonMatching []int
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				matching = helpers.Collect(ctx, split.Matching().Out())
			}()

			go func() {
				defer wg.Done()
				nonMatching = helpers.Collect(ctx, split.NonMatching().Out())
			}()

			wg.Wait()

			// Then
			// With cancelled context, should collect few or
			// no items
			if c.expectedMatching == nil {
				assert.LessOrEqual(t, len(matching), 10)
			} else {
				assert.ElementsMatch(t, c.expectedMatching, matching)
			}
			if c.expectedNonMatching == nil {
				assert.LessOrEqual(t, len(nonMatching), 10)
			} else {
				assert.ElementsMatch(t, c.expectedNonMatching, nonMatching)
			}
		})
	}
}

func TestSplitOperator_ToSink(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	isEven := func(x int) (bool, error) { return x%2 == 0, nil }

	cases := map[string]struct {
		expectedMatching    []int
		expectedNonMatching []int
		subject             func() (
			*operators.SplitOperator[int],
			*helpers.Collector[int],
			*helpers.Collector[int],
		)
	}{
		"streams-to-both-collectors": {
			expectedMatching:    []int{2, 4, 6, 8, 10},
			expectedNonMatching: []int{1, 3, 5, 7, 9},
			subject: func() (
				*operators.SplitOperator[int],
				*helpers.Collector[int],
				*helpers.Collector[int],
			) {
				source := sources.Slice(items).Build()
				split := operators.Split(source, isEven).Build()
				collector1 := helpers.NewCollector[int](ctx)
				collector2 := helpers.NewCollector[int](ctx)
				return split, collector1, collector2
			},
		},
		"streams-with-buffer": {
			expectedMatching:    []int{2, 4, 6, 8, 10},
			expectedNonMatching: []int{1, 3, 5, 7, 9},
			subject: func() (
				*operators.SplitOperator[int],
				*helpers.Collector[int],
				*helpers.Collector[int],
			) {
				source := sources.Slice(items).Build()
				split := operators.
					Split(source, isEven).
					BufferSize(10).
					Build()
				collector1 := helpers.NewCollector[int](ctx)
				collector2 := helpers.NewCollector[int](ctx)
				return split, collector1, collector2
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			split, collector1, collector2 := c.subject()

			// When
			split.Matching().ToSink(collector1)
			split.NonMatching().ToSink(collector2)

			// Then
			assert.ElementsMatch(t, c.expectedMatching, collector1.Items())
			assert.ElementsMatch(t, c.expectedNonMatching, collector2.Items())
		})
	}
}

func TestSplitOperator_PanicOnNilPredicate(t *testing.T) {
	t.Parallel()

	// Given
	items := []int{1, 2, 3, 4, 5}
	source := sources.Slice(items).Build()

	// When & Then
	assert.Panics(t, func() { operators.Split(source, nil).Build() })
}
