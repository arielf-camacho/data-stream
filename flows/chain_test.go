package flows_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/arielf-camacho/data-stream/flows"
	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
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

// func TestToFlow_ContextCancellation(t *testing.T) {
// 	t.Parallel()

// 	items := []int{1, 2, 3, 4, 5}

// 	cases := map[string]struct {
// 		expectedCollected []string
// 		subject           func() (primitives.Flow[string, string], context.Context, context.CancelFunc)
// 	}{
// 		"cancelled-before-processing": {
// 			expectedCollected: []string{},
// 			subject: func() (primitives.Flow[string, string], context.Context, context.CancelFunc) {
// 				ctx, cancel := context.WithCancel(context.Background())

// 				// First flow: int -> string
// 				from := flows.Map(func(x int) (string, error) {
// 					return strconv.Itoa(x), nil
// 				}).Build()

// 				// Second flow: string -> string
// 				to := flows.Map(func(x string) (string, error) {
// 					return "processed:" + x, nil
// 				}).Build()

// 				// Chain them using ToFlow utility
// 				flows.ToFlow(ctx, from, to)

// 				// Cancel immediately
// 				cancel()

// 				// Feed input to first flow
// 				go func() {
// 					for _, item := range items {
// 						from.In() <- item
// 					}
// 					close(from.In())
// 				}()

// 				return to, ctx, cancel
// 			},
// 		},
// 		"cancelled-during-processing": {
// 			expectedCollected: []string{"1", "2"}, // Should collect some values before cancellation
// 			subject: func() (primitives.Flow[string, string], context.Context, context.CancelFunc) {
// 				ctx, cancel := context.WithCancel(context.Background())

// 				// First flow: int -> string
// 				from := flows.Map(func(x int) (string, error) {
// 					return strconv.Itoa(x), nil
// 				}).Build()

// 				// Second flow: string -> string
// 				to := flows.Map(func(x string) (string, error) {
// 					return "processed:" + x, nil
// 				}).Build()

// 				// Chain them using ToFlow utility
// 				flows.ToFlow(ctx, from, to)

// 				// Feed input to first flow with cancellation
// 				go func() {
// 					for i, item := range items {
// 						from.In() <- item
// 						if i == 1 { // Cancel after sending 2 items
// 							cancel()
// 						}
// 					}
// 					close(from.In())
// 				}()

// 				return to, ctx, cancel
// 			},
// 		},
// 	}

// 	for name, c := range cases {
// 		t.Run(name, func(t *testing.T) {
// 			t.Parallel()

// 			// Given
// 			flow, ctx, _ := c.subject()

// 			// When
// 			var collected []string
// 			var wg sync.WaitGroup
// 			wg.Add(1)

// 			go func() {
// 				defer wg.Done()
// 				collected = helpers.Collect(ctx, flow.Out())
// 			}()

// 			// Wait for processing to complete or timeout
// 			select {
// 			case <-time.After(1 * time.Second):
// 				// Ensure cleanup
// 			case <-ctx.Done():
// 			}

// 			wg.Wait()

// 			// Then
// 			if len(c.expectedCollected) > 0 {
// 				assert.Subset(t, c.expectedCollected, collected)
// 			} else {
// 				assert.Empty(t, collected)
// 			}
// 		})
// 	}
// }

// func TestToFlow_TypeSafety(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()
// 	items := []int{1, 2, 3}

// 	cases := map[string]struct {
// 		expected []float64
// 		subject  func() (primitives.Flow[float64, float64], context.Context)
// 	}{
// 		"int-to-string-to-float64": {
// 			expected: []float64{1.0, 2.0, 3.0},
// 			subject: func() (primitives.Flow[float64, float64], context.Context) {
// 				// First flow: int -> string
// 				from := flows.Map(func(x int) (string, error) {
// 					return strconv.Itoa(x), nil
// 				}).Build()

// 				// Second flow: string -> float64
// 				to := flows.Map(func(x string) (float64, error) {
// 					return strconv.ParseFloat(x, 64)
// 				}).Build()

// 				// Chain them using ToFlow utility
// 				flows.ToFlow(ctx, from, to)

// 				// Feed input to first flow
// 				go func() {
// 					for _, item := range items {
// 						from.In() <- item
// 					}
// 					close(from.In())
// 				}()

// 				return to, ctx
// 			},
// 		},
// 	}

// 	for name, c := range cases {
// 		t.Run(name, func(t *testing.T) {
// 			t.Parallel()

// 			// Given
// 			flow, ctx := c.subject()

// 			// When
// 			collected := helpers.Collect(ctx, flow.Out())

// 			// Then
// 			assert.ElementsMatch(t, c.expected, collected)
// 		})
// 	}
// }

// func TestToFlow_WithSource(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()
// 	items := []int{1, 2, 3, 4, 5}

// 	cases := map[string]struct {
// 		expected []string
// 		subject  func() (primitives.Flow[string, string], context.Context)
// 	}{
// 		"source-to-map-to-filter": {
// 			expected: []string{"4", "8"},
// 			subject: func() (primitives.Flow[string, string], context.Context) {
// 				// Source: provides int values
// 				source := sources.Slice(items).Build()

// 				// First flow: int -> int (doubles)
// 				from := flows.Map(func(x int) (int, error) {
// 					return x * 2, nil
// 				}).Build()

// 				// Second flow: int -> string (filters even numbers)
// 				filter := flows.Filter(func(x int) (bool, error) {
// 					return x%2 == 0, nil
// 				}).Build()

// 				to := flows.Map(func(x int) (string, error) {
// 					return strconv.Itoa(x), nil
// 				}).Build()

// 				// Chain filter to map
// 				flows.ToFlow(ctx, filter, to)

// 				// Connect source to first flow
// 				source.ToFlow(from)

// 				// Chain first flow to filter
// 				flows.ToFlow(ctx, from, filter)

// 				return to, ctx
// 			},
// 		},
// 	}

// 	for name, c := range cases {
// 		t.Run(name, func(t *testing.T) {
// 			t.Parallel()

// 			// Given
// 			flow, ctx := c.subject()

// 			// When
// 			collected := helpers.Collect(ctx, flow.Out())

// 			// Then
// 			assert.ElementsMatch(t, c.expected, collected)
// 		})
// 	}
// }

// func TestToFlow_ConcurrentAccess(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()
// 	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

// 	cases := map[string]struct {
// 		expected []string
// 		subject  func() (primitives.Flow[string, string], context.Context)
// 	}{
// 		"concurrent-processing": {
// 			expected: []string{"2", "4", "6", "8", "10", "12", "14", "16", "18", "20"},
// 			subject: func() (primitives.Flow[string, string], context.Context) {
// 				// First flow: int -> int (doubles) with parallelism
// 				from := flows.Map(func(x int) (int, error) {
// 					return x * 2, nil
// 				}).Parallelism(3).Build()

// 				// Second flow: int -> string
// 				to := flows.Map(func(x int) (string, error) {
// 					return strconv.Itoa(x), nil
// 				}).Build()

// 				// Chain them using ToFlow utility
// 				flows.ToFlow(ctx, from, to)

// 				// Feed input to first flow
// 				go func() {
// 					for _, item := range items {
// 						from.In() <- item
// 					}
// 					close(from.In())
// 				}()

// 				return to, ctx
// 			},
// 		},
// 	}

// 	for name, c := range cases {
// 		t.Run(name, func(t *testing.T) {
// 			t.Parallel()

// 			// Given
// 			flow, ctx := c.subject()

// 			// When
// 			collected := helpers.Collect(ctx, flow.Out())

// 			// Then
// 			assert.ElementsMatch(t, c.expected, collected)
// 		})
// 	}
// }
