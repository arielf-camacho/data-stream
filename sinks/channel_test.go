package sinks

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/stretchr/testify/assert"
)

func TestChannelSink_In(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expected   []int
		subject    func() (*ChannelSink[int], chan int, context.Context)
		bufferSize uint
	}{
		"streams-all-values-to-channel": {
			expected:   []int{1, 2, 3, 4, 5},
			bufferSize: 1,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
		"streams-with-buffer": {
			expected:   []int{1, 2, 3, 4, 5},
			bufferSize: 10,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(10).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
		"streams-empty-input": {
			expected:   []int{},
			bufferSize: 1,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
		"streams-single-value": {
			expected:   []int{42},
			bufferSize: 1,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					sink.In() <- 42
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			_, out, _ := c.subject()

			// When
			var collected []int
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				for range c.expected {
					collected = append(collected, <-out)
				}
			}()

			wg.Wait()

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}

func TestChannelSink_ContextCancellation(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expectedCollected []int
		subject           func() (
			*ChannelSink[int], chan int, context.Context, context.CancelFunc,
		)
	}{
		"cancelled-before-processing": {
			expectedCollected: []int{},
			subject: func() (
				*ChannelSink[int], chan int, context.Context, context.CancelFunc,
			) {
				out := make(chan int, 10)
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				sink := Channel(out).Context(ctx).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx, cancel
			},
		},
		"cancelled-during-processing": {
			expectedCollected: []int{1, 2}, // Should collect some values before cancellation
			subject: func() (
				*ChannelSink[int], chan int, context.Context, context.CancelFunc,
			) {
				out := make(chan int, 10)
				ctx, cancel := context.WithCancel(context.Background())
				sink := Channel(out).Context(ctx).BufferSize(1).Build()
				go func() {
					for i, item := range items {
						sink.In() <- item
						if i == 1 {
							cancel()
						}
					}
					close(sink.In())
				}()
				return sink, out, ctx, cancel
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			_, out, ctx, _ := c.subject()

			// When
			var collected []int
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				collected = helpers.Collect(ctx, out)
			}()
			wg.Wait()

			// Then
			if len(c.expectedCollected) > 0 {
				assert.Subset(t, c.expectedCollected, collected)
			} else {
				assert.Empty(t, collected)
			}
		})
	}
}

func TestChannelSink_BuilderPattern(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3}

	cases := map[string]struct {
		expected   []int
		subject    func() (*ChannelSink[int], chan int, context.Context)
		bufferSize uint
	}{
		"default-context": {
			expected:   []int{1, 2, 3},
			bufferSize: 1,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
		"custom-context": {
			expected:   []int{1, 2, 3},
			bufferSize: 5,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				customCtx := context.Background()
				out := make(chan int, 10)
				sink := Channel(out).Context(customCtx).BufferSize(5).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, customCtx
			},
		},
		"zero-buffer-size": {
			expected:   []int{1, 2, 3},
			bufferSize: 0,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(0).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			_, out, _ := c.subject()

			// When
			var collected []int
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				for range c.expected {
					collected = append(collected, <-out)
				}
			}()

			wg.Wait()

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}

func TestChannelSink_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	cases := map[string]struct {
		expected   []int
		subject    func() (*ChannelSink[int], chan int, context.Context)
		concurrent bool
	}{
		"sequential-processing": {
			expected:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			concurrent: false,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 20)
				sink := Channel(out).BufferSize(10).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
		"concurrent-sending": {
			expected:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			concurrent: true,
			subject: func() (*ChannelSink[int], chan int, context.Context) {
				out := make(chan int, 20)
				sink := Channel(out).BufferSize(10).Build()
				var wg sync.WaitGroup
				for _, item := range items {
					wg.Add(1)
					go func(val int) {
						defer wg.Done()
						sink.In() <- val
					}(item)
				}
				go func() {
					wg.Wait()
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			_, out, _ := c.subject()

			// When
			var collected []int
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				for range c.expected {
					collected = append(collected, <-out)
				}
			}()

			wg.Wait()

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}

func TestChannelSink_TypeSafety(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		expected []string
		subject  func() (*ChannelSink[string], chan string, context.Context)
	}{
		"string-values": {
			expected: []string{"a", "b", "c"},
			subject: func() (*ChannelSink[string], chan string, context.Context) {
				out := make(chan string, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					sink.In() <- "a"
					sink.In() <- "b"
					sink.In() <- "c"
					close(sink.In())
				}()
				return sink, out, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			_, out, _ := c.subject()

			// When
			var collected []string
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				for range c.expected {
					collected = append(collected, <-out)
				}
			}()

			wg.Wait()

			// Then
			assert.ElementsMatch(t, c.expected, collected)
		})
	}
}

func TestChannelSink_Wait(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expectedCollected []int
		expectedError     error
		subject           func() (
			*ChannelSink[int], chan int, context.Context, context.CancelFunc,
		)
		waitFn   func(sink *ChannelSink[int]) error
		assertFn func(
			t *testing.T,
			collected []int,
			err error,
			expectedCollected []int,
			expectedError error,
		)
	}{
		"waits-for-completion": {
			expectedCollected: []int{1, 2, 3, 4, 5},
			expectedError:     nil,
			subject: func() (
				*ChannelSink[int], chan int, context.Context, context.CancelFunc,
			) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx, nil
			},
			waitFn: func(sink *ChannelSink[int]) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				collected []int,
				err error,
				expectedCollected []int,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.ElementsMatch(t, expectedCollected, collected)
			},
		},
		"waits-for-empty-input": {
			expectedCollected: []int{},
			expectedError:     nil,
			subject: func() (
				*ChannelSink[int],
				chan int,
				context.Context,
				context.CancelFunc,
			) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					close(sink.In())
				}()
				return sink, out, ctx, nil
			},
			waitFn: func(sink *ChannelSink[int]) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				collected []int,
				err error,
				expectedCollected []int,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.ElementsMatch(t, expectedCollected, collected)
			},
		},
		"waits-for-single-value": {
			expectedCollected: []int{42},
			expectedError:     nil,
			subject: func() (
				*ChannelSink[int], chan int, context.Context, context.CancelFunc,
			) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					sink.In() <- 42
					close(sink.In())
				}()
				return sink, out, ctx, nil
			},
			waitFn: func(sink *ChannelSink[int]) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				collected []int,
				err error,
				expectedCollected []int,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.ElementsMatch(t, expectedCollected, collected)
			},
		},
		"waits-with-buffer": {
			expectedCollected: []int{1, 2, 3, 4, 5},
			expectedError:     nil,
			subject: func() (
				*ChannelSink[int], chan int, context.Context, context.CancelFunc,
			) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(10).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx, nil
			},
			waitFn: func(sink *ChannelSink[int]) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				collected []int,
				err error,
				expectedCollected []int,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.ElementsMatch(t, expectedCollected, collected)
			},
		},
		"waits-until-cancelled-before-processing": {
			expectedCollected: []int{},
			expectedError:     nil,
			subject: func() (
				*ChannelSink[int], chan int, context.Context, context.CancelFunc,
			) {
				out := make(chan int, 10)
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				sink := Channel(out).Context(ctx).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx, cancel
			},
			waitFn: func(sink *ChannelSink[int]) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				collected []int,
				err error,
				expectedCollected []int,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.Empty(t, collected)
			},
		},
		"waits-until-cancelled-during-processing": {
			expectedCollected: []int{1, 2}, // Should collect some values before cancellation
			expectedError:     nil,
			subject: func() (
				*ChannelSink[int],
				chan int,
				context.Context,
				context.CancelFunc,
			) {
				out := make(chan int, 10)
				ctx, cancel := context.WithCancel(context.Background())
				sink := Channel(out).Context(ctx).BufferSize(1).Build()
				go func() {
					for i, item := range items {
						sink.In() <- item
						if i == 1 {
							cancel()
						}
					}
					close(sink.In())
				}()
				return sink, out, ctx, cancel
			},
			waitFn: func(sink *ChannelSink[int]) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				collected []int,
				err error,
				expectedCollected []int,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.Subset(t, expectedCollected, collected)
			},
		},
		"waits-blocks-until-completion": {
			expectedCollected: []int{1, 2, 3, 4, 5},
			expectedError:     nil,
			subject: func() (
				*ChannelSink[int],
				chan int,
				context.Context,
				context.CancelFunc,
			) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
						time.Sleep(10 * time.Millisecond) // Simulate some processing time
					}
					time.Sleep(10 * time.Millisecond)
					close(sink.In())
				}()
				return sink, out, ctx, nil
			},
			waitFn: func(sink *ChannelSink[int]) error {
				start := time.Now()
				err := sink.Wait()
				elapsed := time.Since(start)
				// Should take some time, but allow for very fast execution
				assert.GreaterOrEqual(t, elapsed, 0*time.Millisecond)
				return err
			},
			assertFn: func(
				t *testing.T,
				collected []int,
				err error,
				expectedCollected []int,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.ElementsMatch(t, expectedCollected, collected)
			},
		},
		"waits-multiple-calls": {
			expectedCollected: []int{1, 2, 3},
			expectedError:     nil,
			subject: func() (
				*ChannelSink[int], chan int, context.Context, context.CancelFunc,
			) {
				out := make(chan int, 10)
				sink := Channel(out).BufferSize(1).Build()
				go func() {
					for _, item := range []int{1, 2, 3} {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, out, ctx, nil
			},
			waitFn: func(sink *ChannelSink[int]) error {
				// Multiple Wait calls should all succeed
				err1 := sink.Wait()
				err2 := sink.Wait()
				err3 := sink.Wait()
				assert.NoError(t, err1)
				assert.NoError(t, err2)
				assert.NoError(t, err3)
				return err1
			},
			assertFn: func(
				t *testing.T,
				collected []int,
				err error,
				expectedCollected []int,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.ElementsMatch(t, expectedCollected, collected)
			},
		},
		"waits-concurrent-access": {
			expectedCollected: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			expectedError:     nil,
			subject: func() (
				*ChannelSink[int], chan int, context.Context, context.CancelFunc,
			) {
				out := make(chan int, 20)
				sink := Channel(out).BufferSize(10).Build()
				items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
				var wg sync.WaitGroup

				// Start multiple goroutines sending data concurrently
				for _, item := range items {
					wg.Add(1)
					go func(val int) {
						defer wg.Done()
						sink.In() <- val
					}(item)
				}

				// Start a goroutine to close the input channel when all data is sent
				go func() {
					wg.Wait()
					close(sink.In())
				}()
				return sink, out, ctx, nil
			},
			waitFn: func(sink *ChannelSink[int]) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				collected []int,
				err error,
				expectedCollected []int,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.ElementsMatch(t, expectedCollected, collected)
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			sink, out, ctx, _ := c.subject()

			// When
			var collected []int
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				if ctx.Done() != nil {
					collected = helpers.Collect(ctx, out)
				} else {
					for range c.expectedCollected {
						collected = append(collected, <-out)
					}
				}
			}()

			err := c.waitFn(sink)
			wg.Wait()

			// Then
			c.assertFn(t, collected, err, c.expectedCollected, c.expectedError)
		})
	}
}
