package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/arielf-camacho/data-stream/flows"
	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources"
)

// Demonstrates how to use the data-stream package to
// stream bytes to the console.

func main() {
	source := sources.
		Slice([]byte{'1', '2', '3', '4', '5'}).
		Build()

	nextCharacter := flows.
		Map(func(x byte) (byte, error) { return x + 1, nil }).
		Parallelism(5).
		Build()

	toBytes := flows.
		Map(func(x byte) ([]byte, error) { return []byte{x}, nil }).
		Build()

	sink := sinks.Writer(os.Stdout).Build()

	source.ToFlow(nextCharacter)
	flows.ToFlow(context.Background(), nextCharacter, toBytes)
	toBytes.ToSink(sink)

	time.Sleep(300 * time.Millisecond)
	fmt.Println()
}
