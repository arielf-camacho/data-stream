package main

import (
	"fmt"

	"github.com/arielf-camacho/data-stream/operators"
	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources/slice"
)

// Demonstrates how to use the data-stream package to
// stream bytes to the console.

func main() {
	outputCh := make(chan byte)

	source := slice.
		Slice([]byte{'1', '2', '3', '4', '5'}).
		Build()

	nextCharacter := operators.
		Map(func(x byte) (byte, error) { return x + 1, nil }).
		Parallelism(4).
		Build()

	sink := sinks.Channel(outputCh).Build()

	source.To(nextCharacter)
	nextCharacter.To(sink)

	for v := range outputCh {
		fmt.Println("value:", string(v))
	}
}
