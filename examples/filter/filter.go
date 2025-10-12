package main

import (
	"fmt"

	"github.com/arielf-camacho/data-stream/operators"
	"github.com/arielf-camacho/data-stream/primitives"
	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources/slice"
)

func main() {
	outputCh := make(chan any)

	sink := sinks.NewChannelSink(outputCh)

	int2any := func(x int) (any, error) { return x, nil }
	onlyOdds := func(x int) (bool, error) { return x%2 != 0, nil }
	onlyEvens := func(x int) (bool, error) { return x%2 == 0, nil }

	source1 := slice.NewSliceSource([]int{10, 4, 5, 2, 9})
	filter1 := operators.NewFilterOperator(onlyOdds)
	source1.To(filter1)
	source1AsAny := operators.NewMapOperator(int2any)
	filter1.To(source1AsAny)

	source2 := slice.NewSliceSource([]int{8, 3, 6, 7, 1})
	filter2 := operators.NewFilterOperator(onlyEvens)
	source2.To(filter2)
	source2AsAny := operators.NewMapOperator(int2any)
	filter2.To(source2AsAny)

	merge := operators.NewMergeOperator(
		[]primitives.Out[any]{source1AsAny, source2AsAny},
		operators.WithBufferSizeForMerge(10),
	)
	merge.To(sink)

	for v := range outputCh {
		fmt.Println("value:", v)
	}
}
