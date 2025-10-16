package helpers

import (
	"context"
)

// Collect collects the values from the given channel into a slice and returns
// it. If the context is done, the function returns the collected values so far.
func Collect[T any](ctx context.Context, source <-chan T) []T {
	var result []T
	for {
		select {
		case <-ctx.Done():
			return result
		case v, ok := <-source:
			if !ok {
				return result
			}
			result = append(result, v)
		}
	}
}

// Drain drains the given channel until it is closed.
func Drain[T any](source <-chan T) {
	for range source {
	}
}

// StreamToOptions are the options for the StreamTo function.
type StreamToOptions struct {
	SkipClosingOutputChannel bool
}

// StreamToOption is a function that can be used to configure the StreamTo
// function.
type StreamToOption func(opts *StreamToOptions)

// WithSkipClosingOutputChannel sets the skip closing output channel option for
// the StreamTo function.
func WithSkipClosingOutputChannel(skip bool) StreamToOption {
	return func(opts *StreamToOptions) {
		opts.SkipClosingOutputChannel = skip
	}
}

// StreamTo streams the values from the input channel to the output channel, and
// optionally skips closing the output channel. Optionally sets the done channel
// to be notified when the stream is done. The done channel will be closed when
// the stream is done.
func StreamTo[T any](
	ctx context.Context,
	from <-chan T,
	to chan<- T,
	opts ...StreamToOption,
) {
	options := &StreamToOptions{}
	for _, opt := range opts {
		opt(options)
	}

	go func() {
		if !options.SkipClosingOutputChannel {
			defer close(to)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-from:
				if !ok {
					return
				}

				select {
				case <-ctx.Done():
					return
				case to <- v:
				}
			}
		}
	}()
}
