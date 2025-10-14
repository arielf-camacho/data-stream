package main

import (
	"fmt"
	"time"

	"github.com/arielf-camacho/data-stream/flows"
	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources"
)

func main() {
	// Create a pipeline: source -> filter -> reduce
	source := sources.
		Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).
		Build()

	// Filter even numbers
	filter := flows.
		Filter(func(x int) (bool, error) { return x%2 == 0, nil }).
		Build()

	// Sum the filtered values
	sink := sinks.
		Reduce(func(result, value int, index uint) (int, error) {
			return result + value, nil
		}, 0).
		Build()

	// Connect the pipeline
	source.ToFlow(filter).ToSink(sink)

	// Get the sum of even numbers
	time.Sleep(100 * time.Millisecond)
	sum := sink.Result()
	fmt.Printf("Sum of even numbers: %d\n", sum)
	// Output: Sum of even numbers: 30 (2+4+6+8+10)
}
