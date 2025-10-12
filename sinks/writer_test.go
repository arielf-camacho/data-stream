package sinks

import (
	"bytes"
	"context"
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
