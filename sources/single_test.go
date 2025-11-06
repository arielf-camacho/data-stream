package sources_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
	"github.com/arielf-camacho/data-stream/sources"
)

func TestSingleSource_Out(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		expected     []int
		expectedErr  error
		subject      func(errCh chan error) primitives.Source[int]
		expectErrSet bool
	}{
		"emits-single-value": {
			expected: []int{42},
			subject: func(errCh chan error) primitives.Source[int] {
				return sources.
					Single(func() (int, error) { return 42, nil }).
					Build()
			},
		},
		"no-value-on-error": {
			expected:     nil,
			expectedErr:  assert.AnError,
			expectErrSet: true,
			subject: func(errCh chan error) primitives.Source[int] {
				return sources.
					Single(func() (int, error) { return 0, assert.AnError }).
					ErrorHandler(func(err error) { errCh <- err }).
					Build()
			},
		},
		"error-without-handler-does-not-panic": {
			expected: nil,
			subject: func(errCh chan error) primitives.Source[int] {
				return sources.
					Single(func() (int, error) { return 0, assert.AnError }).
					Build()
			},
		},
		"cancelled-context": {
			expected: nil,
			subject: func(errCh chan error) primitives.Source[int] {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				return sources.
					Single(func() (int, error) { return 42, nil }).
					Context(ctx).
					Build()
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			errCh := make(chan error, 1)
			source := c.subject(errCh)

			// When
			var receivedErr error
			var collected []int

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				collected = helpers.Collect(ctx, source.Out())
			}()

			if c.expectErrSet {
				select {
				case receivedErr = <-errCh:
				case <-time.After(100 * time.Millisecond):
				}
			}

			wg.Wait()

			// Then
			assert.Equal(t, c.expected, collected)
			if c.expectErrSet {
				assert.NotNil(t, receivedErr)
				assert.Equal(t, c.expectedErr.Error(), receivedErr.Error())
			}
		})
	}
}

func TestSingleSource_ToFlow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		expected     []int
		expectedErr  error
		subject      func(errCh chan error) *sources.SingleSource[int]
		expectErrSet bool
	}{
		"streams-single-value-to-collector": {
			expected: []int{42},
			subject: func(errCh chan error) *sources.SingleSource[int] {
				return sources.
					Single(func() (int, error) { return 42, nil }).
					Build()
			},
		},
		"no-value-on-error": {
			expected:     nil,
			expectedErr:  assert.AnError,
			expectErrSet: true,
			subject: func(errCh chan error) *sources.SingleSource[int] {
				return sources.
					Single(func() (int, error) { return 0, assert.AnError }).
					ErrorHandler(func(err error) { errCh <- err }).
					Build()
			},
		},
		"cancelled-context": {
			expected: nil,
			subject: func(errCh chan error) *sources.SingleSource[int] {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				return sources.
					Single(func() (int, error) { return 42, nil }).
					Context(ctx).
					Build()
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			errCh := make(chan error, 1)
			source := c.subject(errCh)
			collector := helpers.NewCollector[int](ctx)

			// When
			source.ToSink(collector)

			var receivedErr error
			if c.expectErrSet {
				select {
				case receivedErr = <-errCh:
				case <-time.After(100 * time.Millisecond):
				}
			}

			// Then
			assert.Equal(t, c.expected, collector.Items())
			if c.expectErrSet {
				assert.NotNil(t, receivedErr)
				assert.Equal(t, c.expectedErr.Error(), receivedErr.Error())
			}
		})
	}
}

func TestSingleSource_ToSink(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := map[string]struct {
		expected     []int
		expectedErr  error
		subject      func(errCh chan error) *sources.SingleSource[int]
		expectErrSet bool
	}{
		"streams-single-value-to-sink": {
			expected: []int{42},
			subject: func(errCh chan error) *sources.SingleSource[int] {
				return sources.
					Single(func() (int, error) { return 42, nil }).
					Build()
			},
		},
		"no-value-on-error": {
			expected:     nil,
			expectedErr:  assert.AnError,
			expectErrSet: true,
			subject: func(errCh chan error) *sources.SingleSource[int] {
				return sources.
					Single(func() (int, error) { return 0, assert.AnError }).
					ErrorHandler(func(err error) { errCh <- err }).
					Build()
			},
		},
		"cancelled-context": {
			expected: nil,
			subject: func(errCh chan error) *sources.SingleSource[int] {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				return sources.
					Single(func() (int, error) { return 42, nil }).
					Context(ctx).
					Build()
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Given
			errCh := make(chan error, 1)
			source := c.subject(errCh)
			collector := helpers.NewCollector[int](ctx)

			// When
			source.ToSink(collector)

			var receivedErr error
			if c.expectErrSet {
				select {
				case receivedErr = <-errCh:
				case <-time.After(100 * time.Millisecond):
				}
			}

			// Then
			assert.Equal(t, c.expected, collector.Items())
			if c.expectErrSet {
				assert.NotNil(t, receivedErr)
				assert.Equal(t, c.expectedErr.Error(), receivedErr.Error())
			}
		})
	}

	t.Run("panics-on-multiple-ToSink-calls", func(t *testing.T) {
		t.Parallel()

		// Given
		source := sources.
			Single(func() (int, error) { return 42, nil }).
			Build()
		collector1 := helpers.NewCollector[int](ctx)
		collector2 := helpers.NewCollector[int](ctx)

		// When
		source.ToSink(collector1)

		// Then
		assert.Panics(t, func() { source.ToSink(collector2) })
	})
}
