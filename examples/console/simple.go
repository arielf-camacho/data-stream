package main

import (
	"fmt"

	"github.com/arielf-camacho/data-stream/flows"
	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources"
)

// Demonstrates how to use the data-stream package to
// stream bytes to the console.

func main() {
	outputCh := make(chan byte)

	source := sources.
		Slice([]byte{'1', '2', '3', '4', '5'}).
		Build()

	nextCharacter := flows.
		Map(func(x byte) (byte, error) { return x + 1, nil }).
		Parallelism(4).
		Build()

	sink := sinks.Channel(outputCh).Build()

	source.ToFlow(nextCharacter).ToSink(sink)

	for v := range outputCh {
		fmt.Println("value:", v)
	}
}
