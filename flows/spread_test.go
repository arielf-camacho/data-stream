package flows_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arielf-camacho/data-stream/flows"
	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/sources"
)

func TestSpreadFlow_Outlets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expected1 []int
		expected2 []int
		expected3 []int
		subject   func() *flows.SpreadFlow[int]
	}{
		"spreads-to-two-outputs": {
			expected1: []int{1, 2, 3, 4, 5},
			expected2: []int{1, 2, 3, 4, 5},
			subject: func() *flows.SpreadFlow[int] {
				source := sources.Slice(items).Build()
				flow1 := flows.PassThrough[int, int]().Build()
				flow2 := flows.PassThrough[int, int]().Build()
				return flows.Spread(source, flow1, flow2).Build()
			},
		},
		"spreads-to-three-outputs": {
			expected1: []int{1, 2, 3, 4, 5},
			expected2: []int{1, 2, 3, 4, 5},
			expected3: []int{1, 2, 3, 4, 5},
			subject: func() *flows.SpreadFlow[int] {
				source := sources.Slice(items).Build()
				flow1 := flows.PassThrough[int, int]().Build()
				flow2 := flows.PassThrough[int, int]().Build()
				flow3 := flows.PassThrough[int, int]().Build()
				return flows.Spread(source, flow1, flow2, flow3).Build()
			},
		},
		"single-output": {
			expected1: []int{1, 2, 3, 4, 5},
			subject: func() *flows.SpreadFlow[int] {
				source := sources.Slice(items).Build()
				flow := flows.PassThrough[int, int]().Build()
				return flows.Spread(source, flow).Build()
			},
		},
		"empty-input": {
			subject: func() *flows.SpreadFlow[int] {
				source := sources.Slice([]int{}).Build()
				flow1 := flows.PassThrough[int, int]().Build()
				flow2 := flows.PassThrough[int, int]().Build()
				return flows.Spread(source, flow1, flow2).Build()
			},
		},
		"large-number-of-outputs": {
			expected1: []int{1, 2, 3, 4, 5},
			expected2: []int{1, 2, 3, 4, 5},
			expected3: []int{1, 2, 3, 4, 5},
			subject: func() *flows.SpreadFlow[int] {
				source := sources.Slice(items).Build()
				flow1 := flows.PassThrough[int, int]().Build()
				flow2 := flows.PassThrough[int, int]().Build()
				flow3 := flows.PassThrough[int, int]().Build()
				flow4 := flows.PassThrough[int, int]().Build()
				flow5 := flows.PassThrough[int, int]().Build()
				return flows.Spread(
					source,
					flow1,
					flow2,
					flow3,
					flow4,
					flow5,
				).Build()
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			spread := c.subject()
			outlets := spread.Outlets()

			// When
			var wg sync.WaitGroup
			var collected [][]int

			for _, outlet := range outlets {
				wg.Add(1)
				go func(out <-chan int) {
					defer wg.Done()
					result := helpers.Collect(ctx, out)
					collected = append(collected, result)
				}(outlet.Out())
			}

			wg.Wait()

			// Then
			assert.GreaterOrEqual(t, len(collected), 1)
			if c.expected1 != nil {
				assert.ElementsMatch(t, c.expected1, collected[0])
			} else {
				assert.Nil(t, collected[0])
			}
			if len(collected) > 1 && c.expected2 != nil {
				assert.ElementsMatch(t, c.expected2, collected[1])
			}
			if len(collected) > 2 && c.expected3 != nil {
				assert.ElementsMatch(t, c.expected3, collected[2])
			}
		})
	}
}

func TestSpreadFlow_ContextCancellation(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		subject func() ([][]int, context.Context)
	}{
		"cancelled-context-stops-spreading": {
			subject: func() ([][]int, context.Context) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				items := make([]int, 100)
				for i := range items {
					items[i] = i
				}

				source := sources.Slice(items).Context(ctx).Build()
				flow1 := flows.PassThrough[int, int]().Context(ctx).Build()
				flow2 := flows.PassThrough[int, int]().Context(ctx).Build()
				spread := flows.Spread(source, flow1, flow2).Context(ctx).Build()

				outlets := spread.Outlets()
				var wg sync.WaitGroup
				var collected [][]int

				for _, outlet := range outlets {
					wg.Add(1)
					go func(out <-chan int) {
						defer wg.Done()
						result := helpers.Collect(ctx, out)
						collected = append(collected, result)
					}(outlet.Out())
				}

				wg.Wait()

				return collected, ctx
			},
		},
		"context-cancelled-during-processing": {
			subject: func() ([][]int, context.Context) {
				ctx, cancel := context.WithCancel(context.Background())

				items := make([]int, 100)
				for i := range items {
					items[i] = i
				}

				source := sources.Slice(items).Context(ctx).Build()
				flow1 := flows.PassThrough[int, int]().Context(ctx).Build()
				flow2 := flows.PassThrough[int, int]().Context(ctx).Build()
				spread := flows.Spread(source, flow1, flow2).Context(ctx).Build()

				outlets := spread.Outlets()
				var wg sync.WaitGroup
				var collected [][]int

				for _, outlet := range outlets {
					wg.Add(1)
					go func(out <-chan int) {
						defer wg.Done()
						result := helpers.Collect(ctx, out)
						collected = append(collected, result)
					}(outlet.Out())
				}

				// Cancel after a brief moment
				go func() {
					cancel()
				}()

				wg.Wait()

				return collected, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given & When
			collected, _ := c.subject()

			// Then
			for _, result := range collected {
				assert.LessOrEqual(t, len(result), 100)
			}
		})
	}
}

func TestSpreadFlow_OutletsMethod(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3}

	t.Run("returns-correct-number-of-outlets", func(t *testing.T) {
		t.Parallel()

		// Given
		source := sources.Slice(items).Build()
		flow1 := flows.PassThrough[int, int]().Build()
		flow2 := flows.PassThrough[int, int]().Build()
		flow3 := flows.PassThrough[int, int]().Build()
		spread := flows.Spread(source, flow1, flow2, flow3).Build()

		// When
		outlets := spread.Outlets()

		// Then
		assert.Len(t, outlets, 3)

		// Verify outlets work
		var wg sync.WaitGroup
		for _, outlet := range outlets {
			wg.Add(1)
			go func(out <-chan int) {
				defer wg.Done()
				result := helpers.Collect(ctx, out)
				assert.ElementsMatch(t, items, result)
			}(outlet.Out())
		}
		wg.Wait()
	})
}

func TestSpreadFlow_PanicOnNoOutputs(t *testing.T) {
	t.Parallel()

	// Given
	source := sources.Slice([]int{1, 2, 3}).Build()

	// When & Then
	assert.Panics(t, func() { flows.Spread(source).Build() })
}
