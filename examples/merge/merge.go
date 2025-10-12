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

	byte2Any := func(x byte) any {
		return x
	}

	source1 := slice.NewSliceSource([]byte{
		'1', '2', '3', '4', '5',
	})
	mapSource1 := operators.NewMapOperator(byte2Any)
	source1.To(mapSource1)

	source2 := slice.NewSliceSource([]byte{
		'6', '7', '8', '9',
	})
	mapSource2 := operators.NewMapOperator(byte2Any)
	source2.To(mapSource2)

	sink := sinks.NewChannelSink(outputCh)

	merge := operators.NewMergeOperator(
		[]primitives.Out[any]{
			mapSource1,
			mapSource2,
		},
		operators.WithBufferSizeForMerge(10),
	)
	merge.To(sink)

	for v := range outputCh {
		fmt.Println("value:", v)
	}
}
