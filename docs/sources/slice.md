# SliceSource

The `SliceSource` is a data source that emits all values from a provided slice
in order. It's ideal for processing collections of data, batch operations, or
converting static data into a stream.

## Design

The `SliceSource` implements the `primitives.Source[T]` interface and provides
a fluent builder API for configuration. It iterates through the provided slice
and emits each value through a channel.

### Key Features

- **Sequential Emission**: Emits values in the exact order they appear in the
  slice
- **Buffer Control**: Configurable buffer size for the output channel
- **Context Support**: Respects context cancellation during emission
- **Memory Efficient**: Processes the slice without copying it

## Graphical Representation

```
Input Slice: [1, 2, 3, 4, 5]

    Slice iteration
         |
         v
    -- 1 -- 2 -- 3 -- 4 -- 5 -- | --> (channel closes)
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a slice source from a slice of integers
    numbers := []int{1, 2, 3, 4, 5}
    source := sources.Slice(numbers).Build()

    // Read all values
    for value := range source.Out() {
        fmt.Println("Received:", value)
    }
    // Output:
    // Received: 1
    // Received: 2
    // Received: 3
    // Received: 4
    // Received: 5
}
```

### With Custom Buffer Size

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    data := []string{"apple", "banana", "cherry", "date", "elderberry"}

    source := sources.
        Slice(data).
        BufferSize(10). // Set buffer size for better performance
        Build()

    for value := range source.Out() {
        fmt.Println("Fruit:", value)
    }
}
```

### With Context Cancellation

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a large slice
    largeSlice := make([]int, 1000)
    for i := range largeSlice {
        largeSlice[i] = i
    }

    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    source := sources.
        Slice(largeSlice).
        Context(ctx).
        Build()

    count := 0
    for value := range source.Out() {
        count++
        if count%100 == 0 {
            fmt.Printf("Processed %d values\n", count)
        }
    }

    fmt.Printf("Total processed: %d values\n", count)
}
```

### Processing Different Data Types

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sources"
)

type Person struct {
    Name string
    Age  int
}

func main() {
    people := []Person{
        {Name: "Alice", Age: 30},
        {Name: "Bob", Age: 25},
        {Name: "Charlie", Age: 35},
    }

    source := sources.Slice(people).Build()

    for person := range source.Out() {
        fmt.Printf("%s is %d years old\n", person.Name, person.Age)
    }
}
```

### Chaining with Transformations

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
    words := []string{"hello", "world", "golang", "streaming"}
    outputCh := make(chan string)

    // Create pipeline: SliceSource -> MapFlow -> ChannelSink
    source := sources.Slice(words).Build()

    // Transform to uppercase
    mapper := flows.
        Map(func(s string) (string, error) {
            return strings.ToUpper(s), nil
        }).
        Build()

    sink := sinks.Channel(outputCh).Build()

    // Connect the pipeline
    source.ToFlow(mapper).ToSink(sink)

    // Consume results
    for value := range outputCh {
        fmt.Println("Uppercase:", value)
    }
    // Output:
    // Uppercase: HELLO
    // Uppercase: WORLD
    // Uppercase: GOLANG
    // Uppercase: STREAMING
}
```

### Filtering and Processing

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    outputCh := make(chan int)

    source := sources.Slice(numbers).Build()

    // Filter even numbers
    filter := flows.
        Filter(func(x int) (bool, error) {
            return x%2 == 0, nil
        }).
        Build()

    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(filter).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Even number:", value)
    }
    // Output:
    // Even number: 2
    // Even number: 4
    // Even number: 6
    // Even number: 8
    // Even number: 10
}
```

### Empty Slice Handling

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Empty slice
    emptySlice := []int{}
    source := sources.Slice(emptySlice).Build()

    // This loop will not execute
    for value := range source.Out() {
        fmt.Println("This won't be printed:", value)
    }

    fmt.Println("No values emitted from empty slice")
}
```

## Builder Methods

### `Slice[T any](slice []T) *SliceSourceBuilder[T]`

Creates a new `SliceSourceBuilder` with the provided slice.

### `Context(ctx context.Context) *SliceSourceBuilder[T]`

Sets the context for cancellation and timeout handling.

### `BufferSize(size uint) *SliceSourceBuilder[T]`

Sets the buffer size for the output channel. Larger buffers can improve
performance for high-throughput scenarios.

### `Build() *SliceSource[T]`

Creates and starts the `SliceSource`.

## Performance Considerations

1. **Buffer Size**: For large slices or high-throughput scenarios, consider
   setting a larger buffer size to reduce blocking.

2. **Memory Usage**: The `SliceSource` doesn't copy the slice, so it's memory
   efficient. However, the slice should not be modified while the source is
   running.

3. **Context Cancellation**: For very large slices, consider using context
   cancellation to allow early termination.

## Important Notes

1. **Single Use**: Each `SliceSource` can only be used once. Attempting to
   connect it to multiple flows or sinks will panic.

2. **Slice Immutability**: The source reads from the slice during iteration.
   Modifying the slice while the source is running may lead to undefined
   behavior.

3. **Empty Slices**: Empty slices will result in no values being emitted, and
   the channel will close immediately.

4. **Channel Management**: The output channel is automatically closed when all
   values have been emitted or when the context is cancelled.

5. **Order Preservation**: Values are emitted in the exact order they appear in
   the slice.
