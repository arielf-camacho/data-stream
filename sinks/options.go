package sinks

import "context"

// WriterSinkOption is a function that can be used to configure a WriterSink.
type WriterSinkOption func(*WriterSink)

// WithBufferSize returns a WriterSinkOption that sets the buffer size for the
// WriterSink.
func WithBufferSize(bufferSize uint) WriterSinkOption {
	return func(s *WriterSink) {
		s.bufferSize = bufferSize
	}
}

// WithContext returns a WriterSinkOption that sets the context for the
// WriterSink.
func WithContext(ctx context.Context) WriterSinkOption {
	return func(s *WriterSink) {
		s.ctx = ctx
	}
}
