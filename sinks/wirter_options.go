package sinks

import "context"

// WriterSinkOption is a function that can be used to configure a WriterSink.
type WriterSinkOption func(*WriterSink)

// WithBufferSizeForWriter returns a WriterSinkOption that sets the buffer size for the
// WriterSink.
func WithBufferSizeForWriter(bufferSize uint) WriterSinkOption {
	return func(s *WriterSink) {
		s.bufferSize = bufferSize
	}
}

// WithContextForWriter returns a WriterSinkOption that sets the context for the
// WriterSink.
func WithContextForWriter(ctx context.Context) WriterSinkOption {
	return func(s *WriterSink) {
		s.ctx = ctx
	}
}
