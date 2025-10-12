package sinks

import (
	"context"
	"io"

	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = primitives.Sink[[]byte](&WriterSink{})

// WriterSink is a sink that writes the values to a
// io.WriterSink.
//
// Graphically, the WriterSink looks like this:
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
// -- WriterSink --
// -> 1 -- 2 -- 3 -- 4 -- 5 -- |
type WriterSink struct {
	ctx        context.Context
	bufferSize uint
	in         chan []byte
	writer     io.Writer
}

// WriterSinkBuilder is a fluent builder for WriterSink.
type WriterSinkBuilder struct {
	writer     io.Writer
	ctx        context.Context
	bufferSize uint
}

// Writer creates a new WriterSinkBuilder for building a WriterSink.
func Writer(w io.Writer) *WriterSinkBuilder {
	return &WriterSinkBuilder{
		writer: w,
		ctx:    context.Background(),
	}
}

// Context sets the context for the WriterSink.
func (b *WriterSinkBuilder) Context(ctx context.Context) *WriterSinkBuilder {
	b.ctx = ctx
	return b
}

// BufferSize sets the buffer size for the WriterSink input channel.
func (b *WriterSinkBuilder) BufferSize(size uint) *WriterSinkBuilder {
	b.bufferSize = size
	return b
}

// Build creates and starts the WriterSink.
func (b *WriterSinkBuilder) Build() *WriterSink {
	writer := &WriterSink{
		writer:     b.writer,
		ctx:        b.ctx,
		bufferSize: b.bufferSize,
	}

	writer.in = make(chan []byte, writer.bufferSize)

	go writer.start()

	return writer
}

func (w *WriterSink) In() chan<- []byte {
	return w.in
}

func (w *WriterSink) start() {
	for v := range w.in {
		select {
		case <-w.ctx.Done():
			return
		default:
			w.writer.Write(v)
		}
	}
}
