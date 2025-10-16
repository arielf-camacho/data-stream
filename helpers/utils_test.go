package helpers_test

import (
	"context"
	"testing"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/stretchr/testify/assert"
)

func TestStreamTo(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expected []int
		subject  func() chan int
	}{
		"streams-all-values": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() chan int {
				in := make(chan int, 10)
				out := make(chan int, 10)
				go func() {
					for _, item := range items {
						in <- item
					}
					close(in)
				}()
				helpers.StreamTo(context.Background(), in, out)
				return out
			},
		},
		"streams-empty-input": {
			expected: nil,
			subject: func() chan int {
				in := make(chan int)
				out := make(chan int, 10)
				close(in)
				helpers.StreamTo(context.Background(), in, out)
				return out
			},
		},
		"streams-single-value": {
			expected: []int{42},
			subject: func() chan int {
				in := make(chan int, 1)
				out := make(chan int, 10)
				go func() {
					in <- 42
					close(in)
				}()
				helpers.StreamTo(context.Background(), in, out)
				return out
			},
		},
		"streams-with-buffer": {
			expected: []int{1, 2, 3, 4, 5},
			subject: func() chan int {
				in := make(chan int, 10)
				out := make(chan int, 10)
				go func() {
					for _, item := range items {
						in <- item
					}
					close(in)
				}()
				helpers.StreamTo(context.Background(), in, out)
				return out
			},
		},
		"cancelled-context": {
			expected: nil,
			subject: func() chan int {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				in := make(chan int, 10)
				out := make(chan int, 10)
				go func() {
					for _, item := range items {
						in <- item
					}
					close(in)
				}()
				helpers.StreamTo(ctx, in, out)
				return out
			},
		},
		"cancelled-context-during-processing": {
			expected: []int{1},
			subject: func() chan int {
				ctx, cancel := context.WithCancel(context.Background())
				in := make(chan int)
				out := make(chan int)
				helpers.StreamTo(ctx, in, out)
				go func() {
					defer close(in)

					for i, item := range items {
						in <- item
						if i == 1 {
							cancel()
							break
						}
					}
				}()
				return out
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			out := c.subject()

			// When
			collected := helpers.Collect(context.Background(), out)

			// Then
			assert.GreaterOrEqual(t, len(collected), len(c.expected))
			for _, item := range c.expected {
				assert.Contains(t, collected, item)
			}
		})
	}
}
