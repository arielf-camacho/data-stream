package main

import (
	"context"
	"fmt"
	"os"

	"github.com/arielf-camacho/data-stream/flows"
	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources"
)

func main() {
	outputCh := make(chan byte)
	go func() {
		for i := 1; i <= 5; i++ {
			outputCh <- byte('0' + i)
		}
		close(outputCh)
	}()

	source := sources.Channel(outputCh).Build()

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

	_ = sink.Wait()

	fmt.Println()
}
