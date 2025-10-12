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

	sink := sinks.Channel(outputCh).Build()

	int2any := func(x int) (any, error) { return x, nil }
	onlyOdds := func(x int) (bool, error) { return x%2 != 0, nil }
	onlyEvens := func(x int) (bool, error) { return x%2 == 0, nil }

	source1 := sources.Slice([]int{10, 4, 5, 2, 9}).Build()
	filter1 := operators.Filter(onlyOdds).Build()
	source1.To(filter1)
	source1AsAny := operators.Map(int2any).Build()
	filter1.To(source1AsAny)

	source2 := sources.Slice([]int{8, 3, 6, 7, 1}).Build()
	filter2 := operators.Filter(onlyEvens).Build()
	source2.To(filter2)
	source2AsAny := operators.Map(int2any).Build()
	filter2.To(source2AsAny)

	merge := operators.
		Merge([]primitives.Out[any]{source1AsAny, source2AsAny}).
		BufferSize(10).
		Build()
	merge.To(sink)

	for v := range outputCh {
		fmt.Println("value:", v)
	}
}
