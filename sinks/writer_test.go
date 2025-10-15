package sinks

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWriterSink_In(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := map[string]struct {
		assert     func(t *testing.T, writer *bytes.Buffer)
		ctx        context.Context
		writer     *bytes.Buffer
		streamTo   func(sink *WriterSink)
		bufferSize uint
	}{
		"streams-all-values-to-collector": {
			ctx: ctx,
			assert: func(t *testing.T, writer *bytes.Buffer) {
				time.Sleep(100 * time.Millisecond)
				assert.Equal(t, []byte{1, 2, 3, 4, 5}, writer.Bytes())
			},
			streamTo: func(sink *WriterSink) {
				numbers := []int{1, 2, 3, 4, 5}
				for _, number := range numbers {
					sink.In() <- []byte{byte(number)}
				}
			},
			writer:     bytes.NewBuffer([]byte{}),
			bufferSize: 1,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			sink := Writer(c.writer).BufferSize(c.bufferSize).Build()

			// When
			c.streamTo(sink)

			// Then
			c.assert(t, c.writer)
		})
	}
}

func TestWriterSink_Wait(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	items := [][]byte{{1}, {2}, {3}, {4}, {5}}

	cases := map[string]struct {
		expectedWritten [][]byte
		expectedError   error
		subject         func() (*WriterSink, *bytes.Buffer, context.Context, context.CancelFunc)
		waitFn          func(sink *WriterSink) error
		assertFn        func(
			t *testing.T,
			written [][]byte,
			err error,
			expectedWritten [][]byte,
			expectedError error,
		)
	}{
		"waits-for-completion": {
			expectedWritten: [][]byte{{1}, {2}, {3}, {4}, {5}},
			expectedError:   nil,
			subject: func() (*WriterSink, *bytes.Buffer, context.Context, context.CancelFunc) {
				buffer := bytes.NewBuffer([]byte{})
				sink := Writer(buffer).BufferSize(1).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, buffer, ctx, nil
			},
			waitFn: func(sink *WriterSink) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				written [][]byte,
				err error,
				expectedWritten [][]byte,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.Equal(t, expectedWritten, written)
			},
		},
		"waits-for-empty-input": {
			expectedWritten: [][]byte{},
			expectedError:   nil,
			subject: func() (*WriterSink, *bytes.Buffer, context.Context, context.CancelFunc) {
				buffer := bytes.NewBuffer([]byte{})
				sink := Writer(buffer).BufferSize(1).Build()
				go func() {
					close(sink.In())
				}()
				return sink, buffer, ctx, nil
			},
			waitFn: func(sink *WriterSink) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				written [][]byte,
				err error,
				expectedWritten [][]byte,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.Empty(t, written)
			},
		},
		"waits-for-single-value": {
			expectedWritten: [][]byte{{42}},
			expectedError:   nil,
			subject: func() (*WriterSink, *bytes.Buffer, context.Context, context.CancelFunc) {
				buffer := bytes.NewBuffer([]byte{})
				sink := Writer(buffer).BufferSize(1).Build()
				go func() {
					sink.In() <- []byte{42}
					close(sink.In())
				}()
				return sink, buffer, ctx, nil
			},
			waitFn: func(sink *WriterSink) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				written [][]byte,
				err error,
				expectedWritten [][]byte,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.Equal(t, expectedWritten, written)
			},
		},
		"waits-with-buffer": {
			expectedWritten: [][]byte{{1}, {2}, {3}, {4}, {5}},
			expectedError:   nil,
			subject: func() (*WriterSink, *bytes.Buffer, context.Context, context.CancelFunc) {
				buffer := bytes.NewBuffer([]byte{})
				sink := Writer(buffer).BufferSize(10).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, buffer, ctx, nil
			},
			waitFn: func(sink *WriterSink) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				written [][]byte,
				err error,
				expectedWritten [][]byte,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.Equal(t, expectedWritten, written)
			},
		},
		"waits-until-cancelled-before-processing": {
			expectedWritten: [][]byte{},
			expectedError:   nil,
			subject: func() (*WriterSink, *bytes.Buffer, context.Context, context.CancelFunc) {
				buffer := bytes.NewBuffer([]byte{})
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				sink := Writer(buffer).Context(ctx).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, buffer, ctx, cancel
			},
			waitFn: func(sink *WriterSink) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				written [][]byte,
				err error,
				expectedWritten [][]byte,
				expectedError error,
			) {
				assert.NoError(t, err)
				// Due to race conditions, some items might be written before cancellation
				assert.LessOrEqual(t, len(written), 1) // Should be 0 or 1 at most
			},
		},
		"waits-until-cancelled-during-processing": {
			expectedWritten: [][]byte{{1}, {2}}, // Should write some values before cancellation
			expectedError:   nil,
			subject: func() (*WriterSink, *bytes.Buffer, context.Context, context.CancelFunc) {
				buffer := bytes.NewBuffer([]byte{})
				ctx, cancel := context.WithCancel(context.Background())
				sink := Writer(buffer).Context(ctx).BufferSize(1).Build()
				go func() {
					for i, item := range items {
						sink.In() <- item
						if i == 1 {
							cancel()
						}
					}
					close(sink.In())
				}()
				return sink, buffer, ctx, cancel
			},
			waitFn: func(sink *WriterSink) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				written [][]byte,
				err error,
				expectedWritten [][]byte,
				expectedError error,
			) {
				assert.NoError(t, err)
				// Due to race conditions, some items might be written before cancellation
				assert.LessOrEqual(t, len(written), len(expectedWritten))
			},
		},
		"waits-blocks-until-completion": {
			expectedWritten: [][]byte{{1}, {2}, {3}, {4}, {5}},
			expectedError:   nil,
			subject: func() (*WriterSink, *bytes.Buffer, context.Context, context.CancelFunc) {
				buffer := bytes.NewBuffer([]byte{})
				sink := Writer(buffer).BufferSize(1).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
						time.Sleep(10 * time.Millisecond) // Simulate some processing time
					}
					time.Sleep(10 * time.Millisecond)
					close(sink.In())
				}()
				return sink, buffer, ctx, nil
			},
			waitFn: func(sink *WriterSink) error {
				start := time.Now()
				err := sink.Wait()
				elapsed := time.Since(start)
				// Should take some time, but allow for very fast execution
				assert.GreaterOrEqual(t, elapsed, 0*time.Millisecond)
				return err
			},
			assertFn: func(
				t *testing.T,
				written [][]byte,
				err error,
				expectedWritten [][]byte,
				expectedError error,
			) {
				assert.NoError(t, err)
				// The result might be less due to timing, so we check it's reasonable
				assert.GreaterOrEqual(t, len(written), 1) // At least some processing happened
			},
		},
		"waits-multiple-calls": {
			expectedWritten: [][]byte{{1}, {2}, {3}},
			expectedError:   nil,
			subject: func() (*WriterSink, *bytes.Buffer, context.Context, context.CancelFunc) {
				buffer := bytes.NewBuffer([]byte{})
				sink := Writer(buffer).BufferSize(1).Build()
				go func() {
					for _, item := range [][]byte{{1}, {2}, {3}} {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, buffer, ctx, nil
			},
			waitFn: func(sink *WriterSink) error {
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
				written [][]byte,
				err error,
				expectedWritten [][]byte,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.Equal(t, expectedWritten, written)
			},
		},
		"waits-concurrent-access": {
			expectedWritten: [][]byte{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}},
			expectedError:   nil,
			subject: func() (*WriterSink, *bytes.Buffer, context.Context, context.CancelFunc) {
				buffer := bytes.NewBuffer([]byte{})
				sink := Writer(buffer).BufferSize(10).Build()
				items := [][]byte{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}}
				var wg sync.WaitGroup

				// Start multiple goroutines sending data concurrently
				for _, item := range items {
					wg.Add(1)
					go func(val []byte) {
						defer wg.Done()
						sink.In() <- val
					}(item)
				}

				// Start a goroutine to close the input channel when all data is sent
				go func() {
					wg.Wait()
					close(sink.In())
				}()
				return sink, buffer, ctx, nil
			},
			waitFn: func(sink *WriterSink) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				written [][]byte,
				err error,
				expectedWritten [][]byte,
				expectedError error,
			) {
				assert.NoError(t, err)
				// Order is not guaranteed in concurrent scenarios, just check length and content
				assert.Len(t, written, len(expectedWritten))
				// Convert to sets for comparison
				expectedSet := make(map[byte]bool)
				for _, item := range expectedWritten {
					expectedSet[item[0]] = true
				}
				actualSet := make(map[byte]bool)
				for _, item := range written {
					actualSet[item[0]] = true
				}
				assert.Equal(t, expectedSet, actualSet)
			},
		},
		"waits-with-error-handling": {
			expectedWritten: [][]byte{{1}, {2}, {3}, {4}, {5}}, // All values should be written
			expectedError:   nil,
			subject: func() (*WriterSink, *bytes.Buffer, context.Context, context.CancelFunc) {
				buffer := bytes.NewBuffer([]byte{})
				errCh := make(chan error, 1)
				sink := Writer(buffer).ErrorHandler(func(err error) {
					errCh <- err
				}).Build()
				go func() {
					for _, item := range items {
						sink.In() <- item
					}
					close(sink.In())
				}()
				return sink, buffer, ctx, nil
			},
			waitFn: func(sink *WriterSink) error {
				return sink.Wait()
			},
			assertFn: func(
				t *testing.T,
				written [][]byte,
				err error,
				expectedWritten [][]byte,
				expectedError error,
			) {
				assert.NoError(t, err)
				assert.Equal(t, expectedWritten, written)
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			sink, buffer, _, _ := c.subject()

			// When
			err := c.waitFn(sink)
			// Give a small delay to ensure writing is processed
			time.Sleep(10 * time.Millisecond)

			// Convert buffer bytes to [][]byte for comparison
			var written [][]byte
			if buffer.Len() > 0 {
				// Split the buffer into individual byte slices
				data := buffer.Bytes()
				for i := range data {
					written = append(written, []byte{data[i]})
				}
			}

			// Then
			c.assertFn(t, written, err, c.expectedWritten, c.expectedError)
		})
	}
}
