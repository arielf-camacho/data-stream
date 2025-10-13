# SpreadFlow

The `SpreadFlow` is a data distribution operator that duplicates a single input
stream to multiple output streams. Each value from the input stream is sent to
all output streams, making it useful for fan-out patterns and parallel
processing.

## Design

The `SpreadFlow` provides a fluent builder API for configuration. It reads
from a single input stream and forwards each value to all provided output
streams sequentially.

### Key Features

- **Fan-Out Pattern**: Distributes one input stream to multiple output streams
- **Sequential Distribution**: Sends each value to all outputs in sequence
- **Context Support**: Respects context cancellation
- **Flexible Output Count**: Can spread to any number of output streams
- **Blocking Behavior**: Slower consumers can block the entire spread operation

## Graphical Representation

```
Input Stream: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->

    SpreadFlow (3 outputs)
         |
         v
Output 1: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
Output 2: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
Output 3: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
```

## Usage Examples

### Basic Spreading

```go
package main

import (
    "fmt"
    "sync"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create output channels
    output1 := make(chan int)
    output2 := make(chan int)
    output3 := make(chan int)

    // Create sinks
    sink1 := sinks.Channel(output1).Build()
    sink2 := sinks.Channel(output2).Build()
    sink3 := sinks.Channel(output3).Build()

    // Create flows for the spread outputs
    flow1 := flows.PassThrough[int, int]().Build()
    flow2 := flows.PassThrough[int, int]().Build()
    flow3 := flows.PassThrough[int, int]().Build()

    // Create the spread flow
    spread := flows.
        Spread(sources.Slice([]int{1, 2, 3, 4, 5}).Build(),
            flow1, flow2, flow3).
        Build()

    // Connect flows to sinks
    flow1.ToSink(sink1)
    flow2.ToSink(sink2)
    flow3.ToSink(sink3)

    var wg sync.WaitGroup
    wg.Add(3)

    // Consumer 1
    go func() {
        defer wg.Done()
        for value := range output1 {
            fmt.Printf("Consumer 1: %d\n", value)
        }
    }()

    // Consumer 2
    go func() {
        defer wg.Done()
        for value := range output2 {
            fmt.Printf("Consumer 2: %d\n", value)
        }
    }()

    // Consumer 3
    go func() {
        defer wg.Done()
        for value := range output3 {
            fmt.Printf("Consumer 3: %d\n", value)
        }
    }()

    wg.Wait()
    // Output (order may vary):
    // Consumer 1: 1
    // Consumer 2: 1
    // Consumer 3: 1
    // Consumer 1: 2
    // Consumer 2: 2
    // Consumer 3: 2
    // ... (continues for all values)
}
```

### Spreading with Different Processing

```go
package main

import (
    "fmt"
    "strings"
    "sync"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create output channels
    upperCh := make(chan string)
    lowerCh := make(chan string)
    lengthCh := make(chan int)

    // Create sinks
    upperSink := sinks.Channel(upperCh).Build()
    lowerSink := sinks.Channel(lowerCh).Build()
    lengthSink := sinks.Channel(lengthCh).Build()

    // Create different processing flows
    upperFlow := flows.
        Map(func(s string) (string, error) {
            return strings.ToUpper(s), nil
        }).
        Build()

    lowerFlow := flows.
        Map(func(s string) (string, error) {
            return strings.ToLower(s), nil
        }).
        Build()

    lengthFlow := flows.
        Map(func(s string) (int, error) {
            return len(s), nil
        }).
        Build()

    // Create the spread flow
    words := []string{"Hello", "World", "Golang", "Streaming"}
    spread := flows.
        Spread(sources.Slice(words).Build(),
            upperFlow, lowerFlow, lengthFlow).
        Build()

    // Connect flows to sinks
    upperFlow.ToSink(upperSink)
    lowerFlow.ToSink(lowerSink)
    lengthFlow.ToSink(lengthSink)

    var wg sync.WaitGroup
    wg.Add(3)

    go func() {
        defer wg.Done()
        for value := range upperCh {
            fmt.Printf("Uppercase: %s\n", value)
        }
    }()

    go func() {
        defer wg.Done()
        for value := range lowerCh {
            fmt.Printf("Lowercase: %s\n", value)
        }
    }()

    go func() {
        defer wg.Done()
        for value := range lengthCh {
            fmt.Printf("Length: %d\n", value)
        }
    }()

    wg.Wait()
    // Output:
    // Uppercase: HELLO
    // Lowercase: hello
    // Length: 5
    // Uppercase: WORLD
    // Lowercase: world
    // Length: 5
    // Uppercase: GOLANG
    // Lowercase: golang
    // Length: 6
    // Uppercase: STREAMING
    // Lowercase: streaming
    // Length: 9
}
```

### Spreading with Buffering

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create output channels
    fastCh := make(chan int)
    slowCh := make(chan int)

    // Create sinks
    fastSink := sinks.Channel(fastCh).Build()
    slowSink := sinks.Channel(slowCh).Build()

    // Create flows with different buffer sizes
    fastFlow := flows.
        PassThrough[int, int]().
        BufferSize(10). // Larger buffer for fast consumer
        Build()

    slowFlow := flows.
        PassThrough[int, int]().
        BufferSize(2). // Smaller buffer for slow consumer
        Build()

    // Create the spread flow
    numbers := make([]int, 100)
    for i := range numbers {
        numbers[i] = i
    }

    spread := flows.
        Spread(sources.Slice(numbers).Build(),
            fastFlow, slowFlow).
        Build()

    // Connect flows to sinks
    fastFlow.ToSink(fastSink)
    slowFlow.ToSink(slowSink)

    var wg sync.WaitGroup
    wg.Add(2)

    // Fast consumer
    go func() {
        defer wg.Done()
        for value := range fastCh {
            fmt.Printf("Fast: %d\n", value)
        }
    }()

    // Slow consumer
    go func() {
        defer wg.Done()
        for value := range slowCh {
            time.Sleep(100 * time.Millisecond) // Simulate slow processing
            fmt.Printf("Slow: %d\n", value)
        }
    }()

    wg.Wait()
}
```

### Spreading with Context Cancellation

```go
package main

import (
    "context"
    "fmt"
    "sync"
    "time"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // Create output channels
    output1 := make(chan int)
    output2 := make(chan int)

    // Create sinks
    sink1 := sinks.Channel(output1).Build()
    sink2 := sinks.Channel(output2).Build()

    // Create flows
    flow1 := flows.PassThrough[int, int]().Build()
    flow2 := flows.PassThrough[int, int]().Build()

    // Create the spread flow with context
    largeData := make([]int, 1000)
    for i := range largeData {
        largeData[i] = i
    }

    spread := flows.
        Spread(sources.Slice(largeData).Build(),
            flow1, flow2).
        Context(ctx).
        Build()

    // Connect flows to sinks
    flow1.ToSink(sink1)
    flow2.ToSink(sink2)

    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        count := 0
        for value := range output1 {
            count++
            if count%100 == 0 {
                fmt.Printf("Output 1 processed: %d\n", count)
            }
        }
        fmt.Printf("Output 1 total: %d\n", count)
    }()

    go func() {
        defer wg.Done()
        count := 0
        for value := range output2 {
            count++
            if count%100 == 0 {
                fmt.Printf("Output 2 processed: %d\n", count)
            }
        }
        fmt.Printf("Output 2 total: %d\n", count)
    }()

    wg.Wait()
}
```

### Complex Pipeline with Spread

```go
package main

import (
    "fmt"
    "sync"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create output channels
    evenCh := make(chan int)
    oddCh := make(chan int)
    squaredCh := make(chan int)

    // Create sinks
    evenSink := sinks.Channel(evenCh).Build()
    oddSink := sinks.Channel(oddCh).Build()
    squaredSink := sinks.Channel(squaredCh).Build()

    // Create processing flows
    evenFlow := flows.
        Filter(func(x int) (bool, error) {
            return x%2 == 0, nil
        }).
        Build()

    oddFlow := flows.
        Filter(func(x int) (bool, error) {
            return x%2 != 0, nil
        }).
        Build()

    squaredFlow := flows.
        Map(func(x int) (int, error) {
            return x * x, nil
        }).
        Build()

    // Create the spread flow
    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    spread := flows.
        Spread(sources.Slice(numbers).Build(),
            evenFlow, oddFlow, squaredFlow).
        Build()

    // Connect flows to sinks
    evenFlow.ToSink(evenSink)
    oddFlow.ToSink(oddSink)
    squaredFlow.ToSink(squaredSink)

    var wg sync.WaitGroup
    wg.Add(3)

    go func() {
        defer wg.Done()
        for value := range evenCh {
            fmt.Printf("Even: %d\n", value)
        }
    }()

    go func() {
        defer wg.Done()
        for value := range oddCh {
            fmt.Printf("Odd: %d\n", value)
        }
    }()

    go func() {
        defer wg.Done()
        for value := range squaredCh {
            fmt.Printf("Squared: %d\n", value)
        }
    }()

    wg.Wait()
    // Output:
    // Even: 2
    // Odd: 1
    // Squared: 1
    // Even: 4
    // Odd: 3
    // Squared: 4
    // Even: 6
    // Odd: 5
    // Squared: 9
    // Even: 8
    // Odd: 7
    // Squared: 16
    // Even: 10
    // Odd: 9
    // Squared: 25
    // Squared: 36
    // Squared: 49
    // Squared: 64
    // Squared: 81
    // Squared: 100
}
```

### Spreading to Merge

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

    // Create processing flows
    upperFlow := flows.
        Map(func(s string) (string, error) {
            return "UPPER: " + s, nil
        }).
        Build()

    lowerFlow := flows.
        Map(func(s string) (string, error) {
            return "lower: " + s, nil
        }).
        Build()

    // Create the spread flow
    words := []string{"Hello", "World", "Golang"}
    spread := flows.
        Spread(sources.Slice(words).Build(),
            upperFlow, lowerFlow).
        Build()

    // Merge the spread outputs
    merge := flows.
        Merge(spread.Outlets()...).
        Build()

    sink := sinks.Channel(outputCh).Build()
    merge.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Merged:", value)
    }
    // Output (order may vary):
    // Merged: UPPER: Hello
    // Merged: lower: Hello
    // Merged: UPPER: World
    // Merged: lower: World
    // Merged: UPPER: Golang
    // Merged: lower: Golang
}
```

## Builder Methods

### `Spread[T any](in primitives.Outlet[T], out ...primitives.Flow[T, T]) *SpreadBuilder[T]`

Creates a new `SpreadBuilder` with the provided input stream and output flows.

### `Context(ctx context.Context) *SpreadBuilder[T]`

Sets the context for cancellation and timeout handling.

### `Build() *SpreadFlow[T]`

Creates and starts the `SpreadFlow`.

## Access Methods

### `Outlets() []primitives.Outlet[T]`

Returns all the output outlets of the spread flow.

## Performance Considerations

1. **Blocking Behavior**: The spread operation is sequential - each value is
   sent to all outputs in order. If one output is slow, it will block the
   entire spread operation.

2. **Buffer Sizes**: Use larger buffer sizes on output flows to reduce blocking
   when consumers have different processing speeds.

3. **Output Count**: More outputs mean more sequential operations per input
   value, which can impact performance.

## Important Notes

1. **Sequential Distribution**: Values are sent to outputs sequentially, not
   concurrently. This means slower consumers can block faster ones.

2. **All Outputs Required**: The spread requires at least one output flow to be
   provided.

3. **Context Cancellation**: When the context is cancelled, the spread stops
   processing and closes all output flows.

4. **Type Consistency**: All output flows must accept the same type `T` as the
   input stream.

5. **Single Use**: Each `SpreadFlow` can only be used once. The input stream
   is consumed by the spread operation.
