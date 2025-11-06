package flows_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/arielf-camacho/data-stream/flows"
	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
	"github.com/arielf-camacho/data-stream/sources"
	"github.com/stretchr/testify/assert"
)

func TestToFlow_Chaining(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expected []string
		subject  func() (primitives.Flow[string, string], context.Context)
	}{
		"chains-int-to-string-to-string": {
			expected: []string{"num:1", "num:2", "num:3", "num:4", "num:5"},
			subject: func() (primitives.Flow[string, string], context.Context) {
				// First flow: int -> string
				from := flows.
					Map(func(x int) (string, error) { return strconv.Itoa(x), nil }).
					Build()

				// Second flow: string -> string (adds prefix)
				to := flows.
					Map(func(x string) (string, error) { return "num:" + x, nil }).
					Build()

				// Chain them using ToFlow utility
				flows.ToFlow(ctx, from, to)

				// Feed input to first flow
				go func() {
					for _, item := range items {
						from.In() <- item
					}
					close(from.In())
				}()

				return to, ctx
			},
		},
		"chains-string-to-int-to-string": {
			expected: []string{"2", "4", "6", "8", "10"},
			subject: func() (primitives.Flow[string, string], context.Context) {
				// First flow: string -> int
				from := flows.
					Map(func(x string) (int, error) { return strconv.Atoi(x) }).
					Build()

				// Second flow: int -> string
				to := flows.
					Map(func(x int) (string, error) { return strconv.Itoa(x * 2), nil }).
					Build()

				// Third flow: string -> string for output
				out := flows.PassThrough[string, string]().Build()

				// Chain them using ToFlow utility
				flows.ToFlow(ctx, from, to).ToFlow(out)

				// Feed input to first flow
				go func() {
					for _, item := range items {
						from.In() <- strconv.Itoa(item)
					}
					close(from.In())
				}()

				return out, ctx
			},
		},
		"chains-with-filter-in-between": {
			expected: []string{"2", "4", "6", "8", "10"},
			subject: func() (primitives.Flow[string, string], context.Context) {
				// First flow: int -> int (doubles)
				from := flows.
					Map(func(x int) (int, error) { return x * 2, nil }).
					Build()

				// Second flow: int -> string (with filter for even numbers)
				filter := flows.
					Filter(func(x int) (bool, error) { return x%2 == 0, nil }).
					Build()

				to := flows.
					Map(func(x int) (string, error) { return strconv.Itoa(x), nil }).
					Build()

				// Chain from flow to filter
				flows.ToFlow(ctx, from, filter)

				// Chain filter to map
				flows.ToFlow(ctx, filter, to)

				// Last flow: string -> string for output
				output := flows.PassThrough[string, string]().Build()

				// Chain to last flow
				to.ToFlow(output)

				// Feed input to first flow
				go func() {
					for _, item := range items {
						from.In() <- item
					}
					close(from.In())
				}()

				return output, ctx
			},
		},
		"chains-empty-input": {
			expected: []string{},
			subject: func() (primitives.Flow[string, string], context.Context) {
				// First flow: int -> string
				from := flows.
					Map(func(x int) (string, error) { return strconv.Itoa(x), nil }).
					Build()

				// Second flow: string -> string
				to := flows.
					Map(func(x string) (string, error) { return "processed:" + x, nil }).
					Build()

				// Chain them using ToFlow utility
				flows.ToFlow(ctx, from, to)

				// Feed empty input to first flow
				go func() {
					close(from.In())
				}()

				return to, ctx
			},
		},
		"chains-single-value": {
			expected: []string{"value:42"},
			subject: func() (primitives.Flow[string, string], context.Context) {
				// First flow: int -> string
				from := flows.
					Map(func(x int) (string, error) { return strconv.Itoa(x), nil }).
					Build()

				// Second flow: string -> string
				to := flows.
					Map(func(x string) (string, error) { return "value:" + x, nil }).
					Build()

				// Chain them using ToFlow utility
				flows.ToFlow(ctx, from, to)

				// Feed single value to first flow
				go func() {
					from.In() <- 42
					close(from.In())
				}()

				return to, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			flow, ctx := c.subject()

			// When
			collector := helpers.NewCollector[string](ctx)
			flow.ToSink(collector)
			collected := collector.Items()

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}

func TestSourceToFlow_Chaining(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expected []string
		subject  func() (primitives.Flow[int, string], context.Context)
	}{
		"chains-slice-source-to-string-flow": {
			expected: []string{"num:1", "num:2", "num:3", "num:4", "num:5"},
			subject: func() (primitives.Flow[int, string], context.Context) {
				// Source: int slice
				source := sources.Slice(items).Build()

				// Flow: int -> string (adds prefix)
				flow := flows.
					Map(func(x int) (string, error) {
						return "num:" + strconv.Itoa(x), nil
					}).
					Build()

				// Chain source to flow using SourceToFlow utility
				flows.SourceToFlow(ctx, source, flow)

				return flow, ctx
			},
		},
		"chains-single-source-to-string-flow": {
			expected: []string{"value:42"},
			subject: func() (primitives.Flow[int, string], context.Context) {
				// Source: single int value
				source := sources.
					Single(func() (int, error) { return 42, nil }).
					Build()

				// Flow: int -> string
				flow := flows.
					Map(func(x int) (string, error) {
						return "value:" + strconv.Itoa(x), nil
					}).
					Build()

				// Chain source to flow using SourceToFlow utility
				flows.SourceToFlow(ctx, source, flow)

				return flow, ctx
			},
		},
		"chains-channel-source-to-string-flow": {
			expected: []string{"2", "4", "6", "8", "10"},
			subject: func() (primitives.Flow[int, string], context.Context) {
				// Source: channel with int values
				ch := make(chan int, 5)
				go func() {
					defer close(ch)
					for _, item := range items {
						ch <- item
					}
				}()
				source := sources.Channel(ch).Build()

				// Flow: int -> int (doubles)
				mapFlow := flows.
					Map(func(x int) (int, error) { return x * 2, nil }).
					Build()

				// Flow: int -> string
				flow := flows.
					Map(func(x int) (string, error) {
						return strconv.Itoa(x), nil
					}).
					Build()

				// Chain source to first flow
				flows.SourceToFlow(ctx, source, mapFlow)

				// Chain first flow to second flow
				flows.ToFlow(ctx, mapFlow, flow)

				return flow, ctx
			},
		},
		"chains-source-to-filter-to-string-flow": {
			expected: []string{"2", "4"},
			subject: func() (primitives.Flow[int, string], context.Context) {
				// Source: int slice
				source := sources.Slice(items).Build()

				// Flow: int -> int (filter even numbers)
				filter := flows.
					Filter(func(x int) (bool, error) { return x%2 == 0, nil }).
					Build()

				// Flow: int -> string
				flow := flows.
					Map(func(x int) (string, error) {
						return strconv.Itoa(x), nil
					}).
					Build()

				// Chain source to filter
				flows.SourceToFlow(ctx, source, filter)

				// Chain filter to map
				flows.ToFlow(ctx, filter, flow)

				return flow, ctx
			},
		},
		"chains-empty-slice-source": {
			expected: []string{},
			subject: func() (primitives.Flow[int, string], context.Context) {
				// Source: empty slice
				source := sources.Slice([]int{}).Build()

				// Flow: int -> string
				flow := flows.
					Map(func(x int) (string, error) {
						return "processed:" + strconv.Itoa(x), nil
					}).
					Build()

				// Chain source to flow using SourceToFlow utility
				flows.SourceToFlow(ctx, source, flow)

				return flow, ctx
			},
		},
		"chains-source-to-multiple-flows": {
			expected: []string{"2", "4", "6", "8", "10"},
			subject: func() (primitives.Flow[int, string], context.Context) {
				// Source: int slice
				source := sources.Slice(items).Build()

				// Flow: int -> int (doubles)
				doubleFlow := flows.
					Map(func(x int) (int, error) { return x * 2, nil }).
					Build()

				// Flow: int -> string
				flow := flows.
					Map(func(x int) (string, error) {
						return strconv.Itoa(x), nil
					}).
					Build()

				// Chain source to double flow
				flows.SourceToFlow(ctx, source, doubleFlow)

				// Chain double flow to string flow
				flows.ToFlow(ctx, doubleFlow, flow)

				return flow, ctx
			},
		},
		"chains-source-with-cancelled-context": {
			expected: []string{},
			subject: func() (primitives.Flow[int, string], context.Context) {
				// Source: int slice with cancelled context
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				source := sources.Slice(items).Context(ctx).Build()

				// Flow: int -> string
				flow := flows.
					Map(func(x int) (string, error) {
						return strconv.Itoa(x), nil
					}).
					Build()

				// Chain source to flow using SourceToFlow utility
				flows.SourceToFlow(ctx, source, flow)

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
			collector := helpers.NewCollector[string](ctx)
			flow.ToSink(collector)
			collected := collector.Items()

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}
