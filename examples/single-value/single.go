package main

import (
	"fmt"

	"github.com/arielf-camacho/data-stream/flows"
	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources"
)

func main() {
	source := sources.Single(func() (int, error) { return 1000, nil }).Build()

	sumFlow := flows.
		Map(func(x int) (int, error) { return x + 1, nil }).
		Build()

	sink := sinks.Single[int]().Build()

	source.ToFlow(sumFlow).ToSink(sink)

	_ = sink.Wait()

	fmt.Println(sink.Result())
}
