package main

import (
	"fmt"

	"github.com/arielf-camacho/data-stream/flows"
	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources"
)

func main() {
	outputCh := make(chan int)

	sink := sinks.Channel(outputCh).Build()

	onlyOdds := func(x int) (bool, error) { return x%2 != 0, nil }
	source1 := sources.Slice([]int{10, 4, 5, 2, 9}).Build()
	filter1 := flows.Filter(onlyOdds).Build()
	source1.ToFlow(filter1)

	onlyEvens := func(x int) (bool, error) { return x%2 == 0, nil }
	source2 := sources.Slice([]int{8, 3, 6, 7, 1}).Build()
	filter2 := flows.Filter(onlyEvens).Build()
	source2.ToFlow(filter2)

	merge := flows.
		Merge(filter1, filter2).
		BufferSize(10).
		Build()

	merge.ToSink(sink)

	for v := range outputCh {
		fmt.Println("value:", v)
	}
}
