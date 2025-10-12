package main

import (
	"os"
	"time"

	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources/slice"
)

// Simple demonstrates how to use the data-stream package to stream bytes to the
// console.
func Simple() {
	sink := sinks.NewWriterSink(os.Stdout)
	source := slice.NewSliceSource([][]byte{
		[]byte("1"),
		[]byte("2"),
		[]byte("3"),
		[]byte("4"),
		[]byte("5"),
		[]byte("\n"),
	})

	source.To(sink)

	time.Sleep(100 * time.Millisecond)
}

func main() {
	Simple()
}
