# ChannelSink

The `ChannelSink` is a data sink that writes incoming values to a Go channel.
It's the most flexible sink type, allowing you to integrate the data stream
with any Go code that can read from channels.

## Design

The `ChannelSink` implements the `primitives.Sink[T]` interface and provides
a fluent builder API for configuration. It acts as a bridge between the data
stream and your application code by forwarding all received values to the
specified output channel.

### Key Features

- **Channel Integration**: Seamlessly integrates with Go's channel-based
  concurrency model
- **Buffer Control**: Configurable buffer size for the input channel
- **Context Support**: Respects context cancellation
- **Flexible Consumption**: The output channel can be consumed by any Go code

## Graphical Representation

```
Input Stream: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->

    ChannelSink
         |
         v
Output Channel: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create an output channel
    outputCh := make(chan int)

    // Create a channel sink
    sink := sinks.Channel(outputCh).Build()

    // Create a simple source
    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()

    // Connect source to sink
    source.ToSink(sink)

    // Consume values from the output channel
    for value := range outputCh {
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
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan string)

    sink := sinks.
        Channel(outputCh).
        BufferSize(10). // Larger buffer for better performance
        Build()

    source := sources.Slice([]string{"a", "b", "c", "d", "e"}).Build()
    source.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Letter:", value)
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
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    outputCh := make(chan int)

    sink := sinks.
        Channel(outputCh).
        Context(ctx).
        Build()

    // Create a large data source
    largeData := make([]int, 1000)
    for i := range largeData {
        largeData[i] = i
    }

    source := sources.Slice(largeData).Build()
    source.ToSink(sink)

    count := 0
    for value := range outputCh {
        count++
        if count%100 == 0 {
            fmt.Printf("Processed %d values\n", count)
        }
    }

    fmt.Printf("Total processed before timeout: %d\n", count)
}
```

### Multiple Consumers

```go
package main

import (
    "fmt"
    "sync"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan string)

    sink := sinks.Channel(outputCh).Build()

    source := sources.Slice([]string{"apple", "banana", "cherry"}).Build()
    source.ToSink(sink)

    var wg sync.WaitGroup
    wg.Add(2)

    // Consumer 1
    go func() {
        defer wg.Done()
        for value := range outputCh {
            fmt.Printf("Consumer 1: %s\n", value)
        }
    }()

    // Consumer 2
    go func() {
        defer wg.Done()
        for value := range outputCh {
            fmt.Printf("Consumer 2: %s\n", value)
        }
    }()

    wg.Wait()
}
```

### Integration with Existing Code

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

// Existing function that processes data from a channel
func processData(inputCh <-chan int) {
    for value := range inputCh {
        fmt.Printf("Processing: %d\n", value)
        // Your existing processing logic here
    }
}

func main() {
    outputCh := make(chan int)

    // Create a pipeline
    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()

    // Transform the data
    mapper := flows.
        Map(func(x int) (int, error) {
            return x * 2, nil
        }).
        Build()

    sink := sinks.Channel(outputCh).Build()

    // Connect the pipeline
    source.ToFlow(mapper).ToSink(sink)

    // Use existing processing function
    processData(outputCh)
}
```

### Error Handling in Pipeline

```go
package main

import (
    "errors"
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan int)

    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()

    // Map that might fail
    mapper := flows.
        Map(func(x int) (int, error) {
            if x == 3 {
                return 0, errors.New("error processing 3")
            }
            return x * 2, nil
        }).
        ErrorHandler(func(err error) {
            fmt.Printf("Map error: %v\n", err)
        }).
        Build()

    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(mapper).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Received:", value)
    }
    // Output:
    // Received: 2
    // Received: 4
    // Map error: error processing 3
    // (pipeline stops due to error)
}
```

### Buffered vs Unbuffered Channels

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Unbuffered channel - synchronous communication
    unbufferedCh := make(chan int)
    unbufferedSink := sinks.Channel(unbufferedCh).Build()

    // Buffered channel - asynchronous communication
    bufferedCh := make(chan int, 5)
    bufferedSink := sinks.Channel(bufferedCh).Build()

    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()

    // Use buffered sink for better performance
    source.ToSink(bufferedSink)

    for value := range bufferedCh {
        fmt.Println("Buffered:", value)
    }
}
```

## Builder Methods

### `Channel[T any](channel chan T) *ChannelSinkBuilder[T]`

Creates a new `ChannelSinkBuilder` with the provided output channel.

### `Context(ctx context.Context) *ChannelSinkBuilder[T]`

Sets the context for cancellation and timeout handling.

### `BufferSize(size uint) *ChannelSinkBuilder[T]`

Sets the buffer size for the input channel. This affects how many values can
be buffered before blocking the upstream flow.

### `Build() *ChannelSink`

Creates and starts the `ChannelSink`.

## Performance Considerations

1. **Buffer Size**: Larger buffer sizes can improve performance by reducing
   blocking between the sink and upstream flows.

2. **Channel Type**: Use buffered channels for better performance when the
   consumer might be slower than the producer.

3. **Consumer Speed**: If the consumer is slower than the producer, consider
   using a larger buffer or multiple consumers.

## Important Notes

1. **Channel Ownership**: The `ChannelSink` does not close the output channel.
   The caller is responsible for closing it when appropriate.

2. **Blocking Behavior**: If the output channel is full and unbuffered, the
   sink will block until a consumer reads from it.

3. **Context Cancellation**: When the context is cancelled, the sink stops
   forwarding values and the input channel is closed.

4. **Error Propagation**: Errors from upstream flows will cause the sink to
   stop forwarding values.

5. **Single Consumer**: By default, only one goroutine should read from the
   output channel. For multiple consumers, consider using a fan-out pattern
   with multiple sinks.
