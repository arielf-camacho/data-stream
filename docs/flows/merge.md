# MergeFlow

The `MergeFlow` is a data combination operator that merges multiple input
streams into a single output stream. It's essential for combining data from
different sources or parallel processing branches.

## Design

The `MergeFlow` implements the `primitives.Outlet[T]` interface and provides
a fluent builder API for configuration. It reads from multiple input streams
concurrently and forwards all values to a single output stream.

### Key Features

- **Multiple Input Streams**: Can merge any number of input streams
- **Concurrent Processing**: Reads from all inputs simultaneously using goroutines
- **Order Preservation**: Values are emitted as they arrive from any input
- **Context Support**: Respects context cancellation
- **Flexible Input**: Accepts any type that implements `primitives.Outlet[T]`

## Graphical Representation

```
Input Stream 1: -- 1 ------- 3 ------- 5 -- | -->
Input Stream 2: ------- 2 ------- 4 ------- | -->
Input Stream 3: ------------ 6 ------------ | -->

    MergeFlow
         |
         v
Output Stream: -- 1 -- 2 -- 3 -- 4 -- 6 -- 5 -- | -->
```

## Usage Examples

### Basic Merging

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan int)

    // Create multiple sources
    source1 := sources.Slice([]int{1, 3, 5}).Build()
    source2 := sources.Slice([]int{2, 4, 6}).Build()

    // Merge the sources
    merge := flows.
        Merge(source1, source2).
        Build()

    sink := sinks.Channel(outputCh).Build()
    merge.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Merged value:", value)
    }
    // Output (order may vary):
    // Merged value: 1
    // Merged value: 2
    // Merged value: 3
    // Merged value: 4
    // Merged value: 5
    // Merged value: 6
}
```

### Merging Different Data Types

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan string)

    // Create sources with different data
    source1 := sources.Slice([]string{"Hello", "World"}).Build()
    source2 := sources.Slice([]string{"Golang", "Streaming"}).Build()
    source3 := sources.Slice([]string{"Data", "Processing"}).Build()

    // Merge all three sources
    merge := flows.
        Merge(source1, source2, source3).
        Build()

    sink := sinks.Channel(outputCh).Build()
    merge.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Word:", value)
    }
    // Output (order may vary):
    // Word: Hello
    // Word: Golang
    // Word: Data
    // Word: World
    // Word: Streaming
    // Word: Processing
}
```

### Merging with Buffer Size

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan int)

    // Create sources with different sizes
    source1 := sources.Slice([]int{1, 2, 3}).Build()
    source2 := sources.Slice([]int{4, 5, 6, 7, 8}).Build()
    source3 := sources.Slice([]int{9, 10}).Build()

    // Merge with custom buffer size
    merge := flows.
        Merge(source1, source2, source3).
        BufferSize(20). // Larger buffer for better performance
        Build()

    sink := sinks.Channel(outputCh).Build()
    merge.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Value:", value)
    }
}
```

### Merging Filtered Streams

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan int)

    // Create a source with mixed data
    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    source := sources.Slice(numbers).Build()

    // Create filters for different conditions
    evenFilter := flows.
        Filter(func(x int) (bool, error) {
            return x%2 == 0, nil
        }).
        Build()

    oddFilter := flows.
        Filter(func(x int) (bool, error) {
            return x%2 != 0, nil
        }).
        Build()

    // Connect source to both filters
    source.ToFlow(evenFilter)
    source.ToFlow(oddFilter)

    // Merge the filtered streams
    merge := flows.
        Merge(evenFilter, oddFilter).
        Build()

    sink := sinks.Channel(outputCh).Build()
    merge.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Filtered value:", value)
    }
    // Output (order may vary):
    // Filtered value: 1
    // Filtered value: 2
    // Filtered value: 3
    // Filtered value: 4
    // ... (all numbers 1-10)
}
```

### Merging with Context Cancellation

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    outputCh := make(chan int)

    // Create sources with different processing times
    source1 := sources.Slice([]int{1, 2, 3, 4, 5}).Build()
    source2 := sources.Slice([]int{6, 7, 8, 9, 10}).Build()

    // Add delays to simulate different processing speeds
    slowMapper1 := flows.
        Map(func(x int) (int, error) {
            time.Sleep(500 * time.Millisecond)
            return x, nil
        }).
        Build()

    slowMapper2 := flows.
        Map(func(x int) (int, error) {
            time.Sleep(300 * time.Millisecond)
            return x, nil
        }).
        Build()

    source1.ToFlow(slowMapper1)
    source2.ToFlow(slowMapper2)

    merge := flows.
        Merge(slowMapper1, slowMapper2).
        Context(ctx).
        Build()

    sink := sinks.Channel(outputCh).Build()
    merge.ToSink(sink)

    count := 0
    for value := range outputCh {
        count++
        fmt.Printf("Received %d: %d\n", count, value)
    }

    fmt.Printf("Total received before timeout: %d\n", count)
}
```

### Merging Single Sources

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan string)

    // Create single sources
    source1 := sources.
        Single(func() (string, error) {
            return "First value", nil
        }).
        Build()

    source2 := sources.
        Single(func() (string, error) {
            return "Second value", nil
        }).
        Build()

    source3 := sources.
        Single(func() (string, error) {
            return "Third value", nil
        }).
        Build()

    // Merge single sources
    merge := flows.
        Merge(source1, source2, source3).
        Build()

    sink := sinks.Channel(outputCh).Build()
    merge.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Single value:", value)
    }
    // Output (order may vary):
    // Single value: First value
    // Single value: Second value
    // Single value: Third value
}
```

### Complex Pipeline with Merge

```go
package main

import (
    "fmt"
    "strings"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan string)

    // Create different data sources
    fruits := []string{"apple", "banana", "cherry"}
    vegetables := []string{"carrot", "broccoli", "spinach"}
    meats := []string{"chicken", "beef", "fish"}

    fruitSource := sources.Slice(fruits).Build()
    vegetableSource := sources.Slice(vegetables).Build()
    meatSource := sources.Slice(meats).Build()

    // Process each category differently
    fruitProcessor := flows.
        Map(func(s string) (string, error) {
            return "Fruit: " + strings.Title(s), nil
        }).
        Build()

    vegetableProcessor := flows.
        Map(func(s string) (string, error) {
            return "Vegetable: " + strings.ToUpper(s), nil
        }).
        Build()

    meatProcessor := flows.
        Map(func(s string) (string, error) {
            return "Meat: " + strings.ToLower(s), nil
        }).
        Build()

    // Connect sources to processors
    fruitSource.ToFlow(fruitProcessor)
    vegetableSource.ToFlow(vegetableProcessor)
    meatSource.ToFlow(meatProcessor)

    // Merge all processed streams
    merge := flows.
        Merge(fruitProcessor, vegetableProcessor, meatProcessor).
        Build()

    sink := sinks.Channel(outputCh).Build()
    merge.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Processed:", value)
    }
    // Output (order may vary):
    // Processed: Fruit: Apple
    // Processed: Vegetable: CARROT
    // Processed: Meat: chicken
    // Processed: Fruit: Banana
    // Processed: Vegetable: BROCCOLI
    // Processed: Meat: beef
    // Processed: Fruit: Cherry
    // Processed: Vegetable: SPINACH
    // Processed: Meat: fish
}
```

## Builder Methods

### `Merge[T any](from ...primitives.Outlet[T]) *MergeBuilder[T]`

Creates a new `MergeBuilder` with the provided input streams.

### `Context(ctx context.Context) *MergeBuilder[T]`

Sets the context for cancellation and timeout handling.

### `BufferSize(size uint) *MergeBuilder[T]`

Sets the buffer size for the output channel.

### `Build() *MergeFlow[T]`

Creates and starts the `MergeFlow`.

## Performance Considerations

1. **Buffer Size**: Larger buffer sizes can improve performance by reducing
   blocking between the merge and downstream flows.

2. **Input Stream Count**: More input streams mean more goroutines. Consider
   the overhead of managing many concurrent readers.

3. **Stream Speed Differences**: If input streams have very different speeds,
   faster streams may be blocked by slower ones.

## Important Notes

1. **Order Non-Deterministic**: The order of values in the output stream is
   not guaranteed and depends on the timing of values arriving from input
   streams.

2. **All Streams Must Close**: The merge will only close its output when all
   input streams have closed.

3. **Context Cancellation**: When the context is cancelled, the merge stops
   reading from all input streams and closes the output.

4. **Type Consistency**: All input streams must emit the same type `T`.

5. **Single Use**: Each `MergeFlow` can only be used once. Attempting to
   connect it to multiple downstream flows will panic.
