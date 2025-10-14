package main

import (
	"fmt"
	"sync"

	"github.com/arielf-camacho/data-stream/flows"
	"github.com/arielf-camacho/data-stream/sinks"
	"github.com/arielf-camacho/data-stream/sources"
)

func main() {
	outputCh1 := make(chan int)
	outputCh2 := make(chan int)

	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	source := sources.Slice(numbers).Build()

	sink1 := sinks.Channel(outputCh1).Build()
	sink2 := sinks.Channel(outputCh2).Build()

	split := flows.
		Split(source, func(x int) (bool, error) { return x%2 == 0, nil }).
		Build()

	split.Matching().ToSink(sink1)
	split.NonMatching().ToSink(sink2)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for v := range outputCh1 {
			fmt.Println("Even number:", v)
		}
	}()

	go func() {
		defer wg.Done()
		for v := range outputCh2 {
			fmt.Println("Odd number:", v)
		}
	}()

	wg.Wait()
}
