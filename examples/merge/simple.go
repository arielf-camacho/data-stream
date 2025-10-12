package main

import (
	"fmt"

	"github.com/arielf-camacho/data-stream/operators"
	"github.com/arielf-camacho/data-stream/primitives"
	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources"
)

func main() {
	outputCh := make(chan any)

	byte2Any := func(x byte) (any, error) {
		return x, nil
	}

	source1 := sources.Slice([]byte{'1', '2', '3', '4', '5'}).Build()
	mapSource1 := operators.Map(byte2Any).Build()
	source1.To(mapSource1)

	source2 := sources.Slice([]byte{'6', '7', '8', '9'}).Build()
	mapSource2 := operators.Map(byte2Any).Build()
	source2.To(mapSource2)

	sink := sinks.Channel(outputCh).Build()

	merge := operators.
		Merge([]primitives.Out[any]{mapSource1, mapSource2}).
		BufferSize(10).
		Build()
	merge.To(sink)

	for v := range outputCh {
		fmt.Println("value:", v)
	}
}
