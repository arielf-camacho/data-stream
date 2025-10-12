package main

import (
	"fmt"

	"github.com/arielf-camacho/data-stream/operators"
	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources"
)

func main() {
	outputCh1 := make(chan string)

	sink1 := sinks.Channel(outputCh1).Build()

	source := sources.
		Single(func() (string, error) { return "Hello, world!", nil }).
		Build()

	map1 := operators.
		Map(func(x string) (string, error) { return x + "1", nil }).
		Build()
	map2 := operators.
		Map(func(x string) (string, error) { return x + "2", nil }).
		Build()
	map3 := operators.
		Map(func(x string) (string, error) { return x + "3", nil }).
		Build()

	spread := operators.Spread(source, map1, map2, map3).Build()

	merge := operators.Merge(spread.Outlets()...).Build()
	mapMerge := operators.
		Map(func(x string) (string, error) { return x + "merge", nil }).
		Build()

	merge.ToFlow(mapMerge).ToSink(sink1)

	for v := range outputCh1 {
		fmt.Println("value from sink1:", v)
	}
}
