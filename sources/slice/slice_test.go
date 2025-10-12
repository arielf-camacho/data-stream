package slice_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
	"github.com/arielf-camacho/data-stream/sources/slice"
)

func TestSliceSource_Out(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		ctx      context.Context
		expected []int
		subject  func() (primitives.Source[int], context.Context)
	}{
		"with-buffer": {
			ctx:      ctx,
			expected: []int{1, 2, 3, 4, 5},
			subject: func() (primitives.Source[int], context.Context) {
				return slice.NewSliceSource(items, slice.WithBuffer[int](2)), ctx
			},
		},
		"without-buffer": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() (primitives.Source[int], context.Context) {
				return slice.NewSliceSource(items), ctx
			},
		},
		"empty-slice": {
			expected: nil,
			subject: func() (primitives.Source[int], context.Context) {
				return slice.NewSliceSource([]int{}), ctx
			},
		},
		"nil-slice": {
			expected: nil,
			subject: func() (primitives.Source[int], context.Context) {
				return slice.NewSliceSource(([]int)(nil)), ctx
			},
		},
		"cancelled-context": {
			expected: ([]int)(nil),
			subject: func() (primitives.Source[int], context.Context) {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				return slice.NewSliceSource(items, slice.WithContext[int](ctx)), ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			source, ctx := c.subject()

			// When
			collected := helpers.Collect(ctx, source.Out())

			// Then
			assert.Equal(t, c.expected, collected)
		})
	}
}

func TestSliceSource_To(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		ctx      context.Context
		expected []int
	}{
		"streams-all-values-to-collector": {
			ctx:      ctx,
			expected: []int{1, 2, 3, 4, 5},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			source := slice.NewSliceSource(items)
			collector := helpers.NewCollector[int](ctx)

			// When
			source.To(collector)

			// Then
			assert.Equal(t, c.expected, collector.Items())
		})
	}
}
