package operators_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/operators"
	"github.com/arielf-camacho/data-stream/primitives"
	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources"
)

func TestMergeOperator_To(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		expected []int
		subject  func() (*operators.MergeOperator[int], *helpers.Collector[int])
	}{
		"merges-two-sources": {
			expected: []int{1, 2, 3, 4, 5, 6},
			subject: func() (*operators.MergeOperator[int], *helpers.Collector[int]) {
				source1 := sources.Slice([]int{1, 2, 3}).Build()
				source2 := sources.Slice([]int{4, 5, 6}).Build()

				collector := helpers.NewCollector[int](ctx)
				merge := operators.Merge(source1, source2).Build()

				return merge, collector
			},
		},
		"merges-three-sources": {
			expected: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
			subject: func() (*operators.MergeOperator[int], *helpers.Collector[int]) {
				source1 := sources.Slice([]int{1, 2, 3}).Build()
				source2 := sources.Slice([]int{4, 5, 6}).Build()
				source3 := sources.Slice([]int{7, 8, 9}).Build()

				collector := helpers.NewCollector[int](ctx)
				merge := operators.Merge(source1, source2, source3).Build()

				return merge, collector
			},
		},
		"single-source": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() (*operators.MergeOperator[int], *helpers.Collector[int]) {
				source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()

				collector := helpers.NewCollector[int](ctx)
				merge := operators.Merge(source).Build()

				return merge, collector
			},
		},
		"empty-sources": {
			expected: nil,
			subject: func() (*operators.MergeOperator[int], *helpers.Collector[int]) {
				source1 := sources.Slice([]int{}).Build()
				source2 := sources.Slice([]int{}).Build()

				collector := helpers.NewCollector[int](ctx)
				merge := operators.Merge(source1, source2).Build()

				return merge, collector
			},
		},
		"mixed-empty-and-populated-sources": {
			expected: []int{1, 2, 3},
			subject: func() (*operators.MergeOperator[int], *helpers.Collector[int]) {
				source1 := sources.Slice([]int{}).Build()
				source2 := sources.Slice([]int{1, 2, 3}).Build()
				source3 := sources.Slice([]int{}).Build()

				collector := helpers.NewCollector[int](ctx)
				merge := operators.Merge(source1, source2, source3).Build()

				return merge, collector
			},
		},
		"with-buffer-size": {
			expected: []int{1, 2, 3, 4},
			subject: func() (*operators.MergeOperator[int], *helpers.Collector[int]) {
				source1 := sources.Slice([]int{1, 2}).Build()
				source2 := sources.Slice([]int{3, 4}).Build()

				collector := helpers.NewCollector[int](ctx)
				merge := operators.Merge(source1, source2).BufferSize(5).Build()

				return merge, collector
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			merge, collector := c.subject()

			// When
			merge.To(collector)

			// Then
			result := collector.Items()
			assert.ElementsMatch(t, c.expected, result)
		})
	}
}

func TestMergeOperator_ConcurrentMerging(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	createSlowSource := func(
		values []int,
		delay time.Duration,
	) primitives.Outlet[int] {
		ch := make(chan int)
		go func() {
			defer close(ch)
			for _, v := range values {
				time.Sleep(delay)
				ch <- v
			}
		}()
		return &testSource{ch: ch}
	}

	cases := map[string]struct {
		subject    func() (*operators.MergeOperator[int], *helpers.Collector[int])
		assertFunc func(*testing.T, []int)
	}{
		"merges-sources-concurrently-not-sequentially": {
			subject: func() (*operators.MergeOperator[int], *helpers.Collector[int]) {
				// Source1: slow, emits 1, 2
				// Source2: fast, emits 3, 4
				// If sequential: [1, 2, 3, 4]
				// If concurrent: likely [1, 3, 2, 4] or similar
				source1 := createSlowSource([]int{1, 2}, 50*time.Millisecond)
				source2 := createSlowSource([]int{3, 4}, 10*time.Millisecond)

				collector := helpers.NewCollector[int](ctx)
				merge := operators.Merge(source1, source2).Build()

				return merge, collector
			},
			assertFunc: func(t *testing.T, result []int) {
				// Should have all 4 values
				assert.Len(t, result, 4)
				assert.ElementsMatch(t, []int{1, 2, 3, 4}, result)
				// If truly concurrent, fast source values should appear before slow
				// source completes. Check that value 3 appears before value 2.
				pos2 := -1
				pos3 := -1
				for i, v := range result {
					if v == 2 {
						pos2 = i
					}
					if v == 3 {
						pos3 = i
					}
				}
				// This proves concurrent merge vs sequential concatenation
				assert.Less(t, pos3, pos2, "Fast source should emit before slow")
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			merge, collector := c.subject()

			// When
			merge.To(collector)

			// Then
			result := collector.Items()
			c.assertFunc(t, result)
		})
	}
}

func TestMergeOperator_ContextCancellation(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		subject func() (context.Context, []int)
	}{
		"cancelled-context-stops-merging": {
			subject: func() (context.Context, []int) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				source1 := sources.Slice([]int{1, 2, 3}).Context(ctx).Build()
				source2 := sources.Slice([]int{4, 5, 6}).Context(ctx).Build()

				outputCh := make(chan int)
				sink := sinks.Channel(outputCh).Context(ctx).Build()

				merge := operators.Merge(source1, source2).Context(ctx).Build()

				merge.ToSink(sink)

				result := helpers.Collect(ctx, outputCh)

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
			// With cancelled context, should collect nothing or very few items
			assert.LessOrEqual(t, len(result), 2)
		})
	}
}

// testSource is a simple Out implementation for testing
type testSource struct {
	ch chan int
}

func (t *testSource) Out() <-chan int {
	return t.ch
}
