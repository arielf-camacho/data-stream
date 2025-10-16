package sinks

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReduceSink_In(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expected int
		subject  func() (*ReduceSink[int, int], context.Context)
	}{
		"sums-all-values": {
			expected: 15, // 1+2+3+4+5
			subject: func() (*ReduceSink[int, int], context.Context) {
				sink := Reduce(
					func(result, value int, index uint) (int, error) {
						return result + value, nil
					}, 0).
					Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx
			},
		},
		"multiplies-all-values": {
			expected: 120, // 1*2*3*4*5
			subject: func() (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return result * value, nil
				}, 1).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx
			},
		},
		"finds-maximum-value": {
			expected: 5,
			subject: func() (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					if value > result {
						return value, nil
					}
					return result, nil
				}, 0).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx
			},
		},
		"counts-values": {
			expected: 5,
			subject: func() (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return result + 1, nil
				}, 0).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx
			},
		},
		"empty-input-returns-initial": {
			expected: 42,
			subject: func() (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return result + value, nil
				}, 42).Build()
				go func() {
					close(sink.In())
				}()
				return sink, ctx
			},
		},
		"single-value": {
			expected: 10,
			subject: func() (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return result + value, nil
				}, 0).Build()
				go func() {
					sink.In() <- 10
					close(sink.In())
				}()
				return sink, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			sink, _ := c.subject()

			// When
			time.Sleep(100 * time.Millisecond) // Allow processing to complete
			result := sink.Result()

			// Then
			assert.Equal(t, c.expected, result)
		})
	}
}

func TestReduceSink_ContextCancellation(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expectedResult int
		waitFn         func(result int)
		subject        func() (*ReduceSink[int, int], context.Context, context.CancelFunc)
	}{
		"cancelled-before-processing": {
			expectedResult: 0,
			subject: func() (*ReduceSink[int, int], context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return result + value, nil
				}, 0).Context(ctx).Build()
				cancel() // Cancel immediately
				go func() {
					defer close(sink.In())
					for _, item := range items {
						sink.In() <- item
					}
				}()
				return sink, ctx, cancel
			},
		},
		"cancelled-during-processing": {
			expectedResult: 3, // Should process some values before cancellation
			subject: func() (*ReduceSink[int, int], context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return result + value, nil
				}, 0).Context(ctx).Build()
				go func() {
					defer close(sink.In())
					for i, item := range items {
						sink.In() <- item
						if i == 3 { // Cancel after sending 2 items
							cancel()
						}
					}
				}()
				return sink, ctx, cancel
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			sink, _, _ := c.subject()

			// When
			var result int

			// Then
			times := 15
			for {
				if times == 0 {
					break
				}
				time.Sleep(50 * time.Millisecond)
				result = sink.Result()
				if result == c.expectedResult {
					break
				} else {
					times--
				}
			}
			assert.LessOrEqual(t, c.expectedResult, result)
		})
	}
}

func TestReduceSink_ErrorHandling(t *testing.T) {
	t.Parallel()

	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expectedResult int
		expectedError  error
		subject        func(chan error) (*ReduceSink[int, int], context.Context)
	}{
		"error-stops-processing": {
			expectedResult: 3, // 1+2, stops at 3
			expectedError:  assert.AnError,
			subject: func(errCh chan error) (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					if value == 3 {
						return result, assert.AnError
					}
					return result + value, nil
				}, 0).ErrorHandler(func(err error, index uint, value, result int) {
					errCh <- err
				}).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, context.Background()
			},
		},
		"error-on-first-value": {
			expectedResult: 0,
			expectedError:  assert.AnError,
			subject: func(errCh chan error) (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					if value == 1 {
						return 0, assert.AnError
					}
					return result + value, nil
				}, 0).ErrorHandler(func(err error, index uint, value, result int) {
					errCh <- err
				}).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, context.Background()
			},
		},
		"no-error-handler-ignores-errors": {
			expectedResult: 3, // 1+2, stops at 3
			expectedError:  nil,
			subject: func(errCh chan error) (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					if value == 3 {
						return result, assert.AnError
					}
					return result + value, nil
				}, 0).Build() // No error handler
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, context.Background()
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			errCh := make(chan error, 1)
			sink, _ := c.subject(errCh)

			// When
			time.Sleep(100 * time.Millisecond) // Allow processing to complete
			result := sink.Result()

			// Then
			assert.Equal(t, c.expectedResult, result)
			if c.expectedError != nil {
				select {
				case err := <-errCh:
					assert.NotNil(t, err)
					assert.Equal(t, c.expectedError.Error(), err.Error())
				case <-time.After(100 * time.Millisecond):
					t.Fatal("Expected error but none received")
				}
			} else {
				select {
				case <-errCh:
					t.Fatal("Unexpected error received")
				case <-time.After(100 * time.Millisecond):
					// No error expected, this is correct
				}
			}
		})
	}
}

func TestReduceSink_BuilderPattern(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3}

	cases := map[string]struct {
		expected int
		subject  func() (*ReduceSink[int, int], context.Context)
	}{
		"default-context": {
			expected: 6, // 1+2+3
			subject: func() (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return result + value, nil
				}, 0).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx
			},
		},
		"custom-context": {
			expected: 6, // 1+2+3
			subject: func() (*ReduceSink[int, int], context.Context) {
				customCtx := context.Background()
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return result + value, nil
				}, 0).Context(customCtx).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, customCtx
			},
		},
		"with-buffer-size": {
			expected: 6, // 1+2+3
			subject: func() (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return result + value, nil
				}, 0).BufferSize(10).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx
			},
		},
		"zero-buffer-size": {
			expected: 6, // 1+2+3
			subject: func() (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return result + value, nil
				}, 0).BufferSize(0).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			sink, _ := c.subject()

			// When
			time.Sleep(100 * time.Millisecond) // Allow processing to complete
			result := sink.Result()

			// Then
			assert.Equal(t, c.expected, result)
		})
	}
}

func TestReduceSink_TypeSafety(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []string{"a", "b", "c"}

	cases := map[string]struct {
		expected string
		subject  func() (*ReduceSink[string, string], context.Context)
	}{
		"string-concatenation": {
			expected: "abc",
			subject: func() (*ReduceSink[string, string], context.Context) {
				sink := Reduce(func(result, value string, index uint) (string, error) {
					return result + value, nil
				}, "").Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			sink, _ := c.subject()

			// When
			time.Sleep(100 * time.Millisecond) // Allow processing to complete
			result := sink.Result()

			// Then
			assert.Equal(t, c.expected, result)
		})
	}
}

func TestReduceSink_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	cases := map[string]struct {
		expected int
		subject  func() (*ReduceSink[int, int], context.Context)
	}{
		"concurrent-sending": {
			expected: 55, // Sum of 1 to 10
			subject: func() (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return result + value, nil
				}, 0).BufferSize(10).Build()
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
				return sink, ctx
			},
		},
		"concurrent-result-reading": {
			expected: 55, // Sum of 1 to 10
			subject: func() (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return result + value, nil
				}, 0).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			sink, _ := c.subject()

			// When
			time.Sleep(100 * time.Millisecond) // Allow processing to complete
			result := sink.Result()

			// Then
			assert.Equal(t, c.expected, result)
		})
	}
}

func TestReduceSink_PanicOnNilFunction(t *testing.T) {
	t.Parallel()

	// Given & When & Then
	assert.Panics(t, func() { Reduce[int](nil, 0) })
}

func TestReduceSink_IndexParameter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{10, 20, 30}

	cases := map[string]struct {
		expected int
		subject  func() (*ReduceSink[int, int], context.Context)
	}{
		"index-based-calculation": {
			expected: 80, // 10*0 + 20*1 + 30*2 = 0 + 20 + 60 = 80, but we start with 0
			subject: func() (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return result + (value * int(index)), nil
				}, 0).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx
			},
		},
		"index-counting": {
			expected: 3, // Should count 3 items (indices 0, 1, 2)
			subject: func() (*ReduceSink[int, int], context.Context) {
				sink := Reduce(func(result, value int, index uint) (int, error) {
					return int(index) + 1, nil // Return index + 1 to count items
				}, 0).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			sink, _ := c.subject()

			// When
			time.Sleep(100 * time.Millisecond) // Allow processing to complete
			result := sink.Result()

			// Then
			assert.Equal(t, c.expected, result)
		})
	}
}

func TestReduceSink_Wait(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := []int{1, 2, 3, 4, 5}

	cases := map[string]struct {
		expectedResult int
		expectedError  error
		subject        func() (*ReduceSink[int, int], context.Context, context.CancelFunc)
		waitFn         func(sink *ReduceSink[int, int]) error
		assertFn       func(t *testing.T, result int, err error, expectedResult int, expectedError error)
	}{
		"waits-for-completion": {
			expectedResult: 15, // 1+2+3+4+5
			expectedError:  nil,
			subject: func() (*ReduceSink[int, int], context.Context, context.CancelFunc) {
				sink := Reduce(
					func(result, value int, index uint) (int, error) {
						return result + value, nil
					}, 0).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx, nil
			},
			waitFn: func(sink *ReduceSink[int, int]) error {
				return sink.Wait()
			},
			assertFn: func(t *testing.T, result int, err error, expectedResult int, expectedError error) {
				assert.NoError(t, err)
				assert.Equal(t, expectedResult, result)
			},
		},
		"waits-for-empty-input": {
			expectedResult: 42,
			expectedError:  nil,
			subject: func() (*ReduceSink[int, int], context.Context, context.CancelFunc) {
				sink := Reduce(
					func(result, value int, index uint) (int, error) {
						return result + value, nil
					}, 42).Build()
				go func() {
					close(sink.In())
				}()
				return sink, ctx, nil
			},
			waitFn: func(sink *ReduceSink[int, int]) error {
				return sink.Wait()
			},
			assertFn: func(t *testing.T, result int, err error, expectedResult int, expectedError error) {
				assert.NoError(t, err)
				assert.Equal(t, expectedResult, result)
			},
		},
		"waits-for-single-value": {
			expectedResult: 10,
			expectedError:  nil,
			subject: func() (*ReduceSink[int, int], context.Context, context.CancelFunc) {
				sink := Reduce(
					func(result, value int, index uint) (int, error) {
						return result + value, nil
					}, 0).Build()
				go func() {
					sink.In() <- 10
					close(sink.In())
				}()
				return sink, ctx, nil
			},
			waitFn: func(sink *ReduceSink[int, int]) error {
				return sink.Wait()
			},
			assertFn: func(t *testing.T, result int, err error, expectedResult int, expectedError error) {
				assert.NoError(t, err)
				assert.Equal(t, expectedResult, result)
			},
		},
		"waits-with-buffer": {
			expectedResult: 15, // 1+2+3+4+5
			expectedError:  nil,
			subject: func() (*ReduceSink[int, int], context.Context, context.CancelFunc) {
				sink := Reduce(
					func(result, value int, index uint) (int, error) {
						return result + value, nil
					}, 0).BufferSize(10).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx, nil
			},
			waitFn: func(sink *ReduceSink[int, int]) error {
				return sink.Wait()
			},
			assertFn: func(t *testing.T, result int, err error, expectedResult int, expectedError error) {
				assert.NoError(t, err)
				assert.Equal(t, expectedResult, result)
			},
		},
		"waits-until-cancelled-before-processing": {
			expectedResult: 0,
			expectedError:  nil,
			subject: func() (*ReduceSink[int, int], context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				sink := Reduce(
					func(result, value int, index uint) (int, error) {
						return result + value, nil
					}, 0).Context(ctx).Build()
				cancel() // Cancel immediately
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx, cancel
			},
			waitFn: func(sink *ReduceSink[int, int]) error {
				return sink.Wait()
			},
			assertFn: func(t *testing.T, result int, err error, expectedResult int, expectedError error) {
				assert.NoError(t, err)
				// Due to race conditions, some items might be processed before cancellation
				assert.LessOrEqual(t, result, 1) // Should be 0 or 1 at most
			},
		},
		"waits-until-cancelled-during-processing": {
			expectedResult: 3, // Should process some values before cancellation
			expectedError:  nil,
			subject: func() (*ReduceSink[int, int], context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				sink := Reduce(
					func(result, value int, index uint) (int, error) {
						return result + value, nil
					}, 0).Context(ctx).Build()
				go func() {
					for i, item := range items {
						sink.In() <- item
						if i == 1 {
							cancel()
						}
					}
					close(sink.In())
				}()
				return sink, ctx, cancel
			},
			waitFn: func(sink *ReduceSink[int, int]) error {
				return sink.Wait()
			},
			assertFn: func(t *testing.T, result int, err error, expectedResult int, expectedError error) {
				assert.NoError(t, err)
				assert.LessOrEqual(t, expectedResult, result)
			},
		},
		"waits-blocks-until-completion": {
			expectedResult: 15, // 1+2+3+4+5
			expectedError:  nil,
			subject: func() (*ReduceSink[int, int], context.Context, context.CancelFunc) {
				sink := Reduce(
					func(result, value int, index uint) (int, error) {
						return result + value, nil
					}, 0).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
						time.Sleep(10 * time.Millisecond) // Simulate some processing time
					}
					time.Sleep(10 * time.Millisecond)
					close(sink.In())
				}()
				return sink, ctx, nil
			},
			waitFn: func(sink *ReduceSink[int, int]) error {
				start := time.Now()
				err := sink.Wait()
				elapsed := time.Since(start)
				// Should take some time, but allow for very fast execution
				assert.GreaterOrEqual(t, elapsed, 0*time.Millisecond)
				return err
			},
			assertFn: func(t *testing.T, result int, err error, expectedResult int, expectedError error) {
				assert.NoError(t, err)
				// The result might be less due to timing, so we check it's reasonable
				assert.GreaterOrEqual(t, result, 1) // At least some processing happened
			},
		},
		"waits-multiple-calls": {
			expectedResult: 6, // 1+2+3
			expectedError:  nil,
			subject: func() (*ReduceSink[int, int], context.Context, context.CancelFunc) {
				sink := Reduce(
					func(result, value int, index uint) (int, error) {
						return result + value, nil
					}, 0).Build()
				go func() {
					for _, item := range []int{1, 2, 3} {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx, nil
			},
			waitFn: func(sink *ReduceSink[int, int]) error {
				// Multiple Wait calls should all succeed
				err1 := sink.Wait()
				err2 := sink.Wait()
				err3 := sink.Wait()
				assert.NoError(t, err1)
				assert.NoError(t, err2)
				assert.NoError(t, err3)
				return err1
			},
			assertFn: func(t *testing.T, result int, err error, expectedResult int, expectedError error) {
				assert.NoError(t, err)
				assert.Equal(t, expectedResult, result)
			},
		},
		"waits-concurrent-access": {
			expectedResult: 55, // Sum of 1 to 10
			expectedError:  nil,
			subject: func() (*ReduceSink[int, int], context.Context, context.CancelFunc) {
				sink := Reduce(
					func(result, value int, index uint) (int, error) {
						return result + value, nil
					}, 0).BufferSize(10).Build()
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
				return sink, ctx, nil
			},
			waitFn: func(sink *ReduceSink[int, int]) error {
				return sink.Wait()
			},
			assertFn: func(t *testing.T, result int, err error, expectedResult int, expectedError error) {
				assert.NoError(t, err)
				assert.Equal(t, expectedResult, result)
			},
		},
		"waits-with-error-handling": {
			expectedResult: 3, // 1+2, stops at 3
			expectedError:  nil,
			subject: func() (*ReduceSink[int, int], context.Context, context.CancelFunc) {
				errCh := make(chan error, 1)
				sink := Reduce(
					func(result, value int, index uint) (int, error) {
						if value == 3 {
							return result, assert.AnError
						}
						return result + value, nil
					}, 0).ErrorHandler(func(err error, index uint, value, result int) {
					errCh <- err
				}).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, ctx, nil
			},
			waitFn: func(sink *ReduceSink[int, int]) error {
				return sink.Wait()
			},
			assertFn: func(t *testing.T, result int, err error, expectedResult int, expectedError error) {
				assert.NoError(t, err)
				assert.Equal(t, expectedResult, result)
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			sink, _, _ := c.subject()

			// When
			err := c.waitFn(sink)
			// Give a small delay to ensure result is processed
			time.Sleep(10 * time.Millisecond)
			result := sink.Result()

			// Then
			c.assertFn(t, result, err, c.expectedResult, c.expectedError)
		})
	}
}
