package main

import (
	"fmt"

	"github.com/arielf-camacho/data-stream/flows"
	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources"
)

func main() {
	outputCh := make(chan byte)

	source1 := sources.Slice([]byte{'1', '2', '3', '4', '5'}).Build()
	source2 := sources.Slice([]byte{'6', '7', '8', '9'}).Build()

	sink := sinks.Channel(outputCh).Build()

	merge := flows.Merge(source1, source2).BufferSize(10).Build()
	merge.ToSink(sink)

	for v := range outputCh {
		fmt.Println("value:", v)
	}
}
