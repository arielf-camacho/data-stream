package sinks

import (
	"context"
	"io"

	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = primitives.Sink[[]byte](&WriterSink{})

// WriterSink is a sink that writes the values to a io.WriterSink.
//
// Graphically, the WriterSink looks like this:
//
// -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
// -- WriterSink --
// -> 1 -- 2 -- 3 -- 4 -- 5 -- |
type WriterSink struct {
	primitives.Sink[[]byte]

	ctx        context.Context
	bufferSize uint
	in         chan []byte
	writer     io.Writer
}

// NewWriterSink returns a new WriterSink given the io.Writer to write to.
func NewWriterSink(w io.Writer, opts ...WriterSinkOption) *WriterSink {
	writer := &WriterSink{
		writer: w,
		ctx:    context.Background(),
	}

	for _, opt := range opts {
		opt(writer)
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
