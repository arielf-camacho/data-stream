# ChannelSource

The `ChannelSource` is a data source that emits values from a provided Go
channel until the channel is closed or the context is cancelled. It's useful
for integrating existing channel-based code with the data stream pipeline.

## Design

The `ChannelSource` implements the `primitives.Source[T]` interface and provides
a fluent builder API for configuration. It reads values from the provided
channel and forwards them to the output stream.

### Key Features

- **Channel Integration**: Seamlessly integrates with Go's channel-based
  concurrency model
- **Context Support**: Respects context cancellation
- **Error Handling**: Supports custom error handlers
- **Thread Safety**: Uses atomic operations to prevent multiple activations

## Graphical Representation

```
Input Channel: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->

    ChannelSource
         |
         v
Output Stream: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
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
    // Create a channel and populate it
    ch := make(chan int, 5)
    go func() {
        defer close(ch)
        for i := 1; i <= 5; i++ {
            ch <- i
        }
    }()

    // Create a channel source
    source := sources.Channel(ch).Build()

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
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    // Create a channel that will take longer than the timeout
    ch := make(chan int, 10)
    go func() {
        defer close(ch)
        for i := 1; i <= 100; i++ {
            ch <- i
            time.Sleep(10 * time.Millisecond)
        }
    }()

    source := sources.
        Channel(ch).
        Context(ctx).
        Build()

    count := 0
    for value := range source.Out() {
        count++
        fmt.Printf("Received: %d\n", value)
    }

    fmt.Printf("Total received before timeout: %d\n", count)
}
```

### With Error Handling

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    ch := make(chan string, 3)
    go func() {
        defer close(ch)
        ch <- "hello"
        ch <- "world"
        ch <- "golang"
    }()

    source := sources.
        Channel(ch).
        ErrorHandler(func(err error) {
            fmt.Printf("Error occurred: %v\n", err)
        }).
        Build()

    for value := range source.Out() {
        fmt.Println("Word:", value)
    }
    // Output:
    // Word: hello
    // Word: world
    // Word: golang
}
```

### Chaining with Flows and Sinks

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

    // Create a channel and populate it
    inputCh := make(chan int, 5)
    go func() {
        defer close(inputCh)
        for i := 1; i <= 5; i++ {
            inputCh <- i
        }
    }()

    // Create pipeline: ChannelSource -> MapFlow -> ChannelSink
    source := sources.Channel(inputCh).Build()

    mapper := flows.
        Map(func(x int) (string, error) {
            return fmt.Sprintf("Number: %d", x), nil
        }).
        Build()

    sink := sinks.Channel(outputCh).Build()

    // Connect the pipeline
    source.ToFlow(mapper).ToSink(sink)

    // Consume results
    for value := range outputCh {
        fmt.Println("Processed:", value)
    }
    // Output:
    // Processed: Number: 1
    // Processed: Number: 2
    // Processed: Number: 3
    // Processed: Number: 4
    // Processed: Number: 5
}
```

### Integration with Existing Channel Code

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

// Existing function that produces data via channel
func produceData(ch chan<- int) {
    defer close(ch)
    for i := 1; i <= 10; i++ {
        ch <- i * i // Send squares
    }
}

func main() {
    outputCh := make(chan int)

    // Create channel for existing function
    dataCh := make(chan int, 10)
    go produceData(dataCh)

    // Integrate with data stream
    source := sources.Channel(dataCh).Build()

    // Filter even numbers
    filter := flows.
        Filter(func(x int) (bool, error) {
            return x%2 == 0, nil
        }).
        Build()

    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(filter).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Even square:", value)
    }
    // Output:
    // Even square: 4
    // Even square: 16
    // Even square: 36
    // Even square: 64
    // Even square: 100
}
```

### Empty Channel Handling

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create an empty channel that's immediately closed
    ch := make(chan int)
    go func() {
        close(ch)
    }()

    source := sources.Channel(ch).Build()

    // This loop will not execute
    count := 0
    for value := range source.Out() {
        count++
        fmt.Println("This won't be printed:", value)
    }

    fmt.Printf("No values emitted from empty channel (count: %d)\n", count)
    // Output: No values emitted from empty channel (count: 0)
}
```

### Buffered vs Unbuffered Channels

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

    // Unbuffered channel - synchronous communication
    unbufferedCh := make(chan int)
    go func() {
        defer close(unbufferedCh)
        for i := 1; i <= 3; i++ {
            unbufferedCh <- i
        }
    }()
    unbufferedSource := sources.Channel(unbufferedCh).Build()

    // Buffered channel - asynchronous communication
    bufferedCh := make(chan int, 10)
    go func() {
        defer close(bufferedCh)
        for i := 4; i <= 6; i++ {
            bufferedCh <- i
        }
    }()
    bufferedSource := sources.Channel(bufferedCh).Build()

    // Process both sources
    mapper := flows.
        Map(func(x int) (string, error) {
            return fmt.Sprintf("Value: %d", x), nil
        }).
        Build()

    sink := sinks.Channel(outputCh).Build()

    // Use buffered source for better performance
    bufferedSource.ToFlow(mapper).ToSink(sink)

    for value := range outputCh {
        fmt.Println(value)
    }
}
```

## Builder Methods

### `Channel[T any](channel chan T) *ChannelSourceBuilder[T]`

Creates a new `ChannelSourceBuilder` with the provided channel.

### `Context(ctx context.Context) *ChannelSourceBuilder[T]`

Sets the context for cancellation and timeout handling.

### `ErrorHandler(handler func(error)) *ChannelSourceBuilder[T]`

Sets a custom error handler for channel operations.

### `Build() *ChannelSource[T]`

Creates and starts the `ChannelSource`.

## Performance Considerations

1. **Channel Buffer Size**: Use buffered channels for better performance when
   the producer might be faster than the consumer.

2. **Context Cancellation**: For long-running channels, consider using context
   cancellation to allow early termination.

3. **Channel Ownership**: The `ChannelSource` does not close the input channel.
   The caller is responsible for closing it when appropriate.

## Important Notes

1. **Single Use**: Each `ChannelSource` can only be used once. Attempting to
   connect it to multiple flows or sinks will panic.

2. **Channel Closure**: The source will stop reading when the input channel is
   closed. Make sure to close the channel when done sending values.

3. **Context Cancellation**: When the context is cancelled, the source stops
   reading from the input channel and closes the output channel.

4. **Channel Management**: The output channel is automatically closed when the
   input channel is closed or when the context is cancelled.

5. **Goroutine Safety**: The source uses atomic operations to ensure it can
   only be activated once.
