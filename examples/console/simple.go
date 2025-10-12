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

	source := slice.NewSliceSource([]byte{
		'1', '2', '3', '4', '5',
	})

	findNextCharacter := func(x byte) (byte, error) {
		return x + 1, nil
	}

	nextCharacter := operators.NewMapOperator(
		findNextCharacter,
		operators.WithParallelismForMap[byte, byte](4),
	)

	sink := sinks.NewChannelSink(outputCh)

	source.To(nextCharacter)
	nextCharacter.To(sink)

	for v := range outputCh {
		fmt.Println("value:", string(v))
	}
}
