package sinks

import "context"

// ChannelSinkOption is a function that can be used to configure a WriterSink.
type ChannelSinkOption[T any] func(*ChannelSink[T])

// WithBufferSizeForChannel returns a ChannelSinkOption that sets the buffer
// size for the ChannelSink.
func WithBufferSizeForChannel[T any](bufferSize uint) ChannelSinkOption[T] {
	return func(s *ChannelSink[T]) {
		s.bufferSize = bufferSize
	}
}

// WithContextForChannel returns a ChannelSinkOption that sets the context for
// the ChannelSink.
func WithContextForChannel[T any](ctx context.Context) ChannelSinkOption[T] {
	return func(s *ChannelSink[T]) {
		s.ctx = ctx
	}
}
