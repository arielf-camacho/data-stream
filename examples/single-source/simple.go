package main

import (
	"fmt"

	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources"
)

func main() {
	outputCh := make(chan any)

	sink := sinks.Channel(outputCh).Build()

	source := sources.Single(func() (any, error) {
		return "Hello, world!", nil
	}).Build()

	source.To(sink)

	for v := range outputCh {
		fmt.Println("value:", v)
	}
}
