# SingleSink

The `SingleSink` is a data sink that gathers the expected SINGLE value from
upstream. It immediately receives the value and closes the input channel, so
please make sure that the upstream will emit ONLY ONE value up to this point.
If the previous condition is not met, and a second value is emitted, that
second value will be taken instead of the first one, and that's not what we
want. This sink also implements the `WaitableSink` interface, so you can wait
for the sink to finish receiving the value. Until the upstream closes, this
sink will still be waiting for the value, and the `Wait()` method will block
until that value is received.

## Design

The `SingleSink` implements both the `primitives.Sink[T]` and
`primitives.WaitableSink[T]` interfaces and provides a fluent builder API for
configuration. It's designed to capture a single value from a stream, making
it perfect for scenarios where you expect exactly one result from your
pipeline.

### Key Features

- **Single Value Capture**: Receives and stores exactly one value from the
  stream
- **Waitable**: Implements `WaitableSink` interface for synchronization
- **Thread-Safe**: Safe concurrent access to the result via `Result()` method
- **Context Support**: Respects context cancellation
- **Immediate Processing**: Receives the value immediately and closes the
  input channel

## Graphical Representation

```
Input Stream: -- 1 --------------------------- | -->

    SingleSink
         |
         v
Result: 1
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
    // Create a single sink
    sink := sinks.Single[int]().Build()

    // Create a source that emits a single value
    source := sources.Single(func() (int, error) {
        return 42, nil
    }).Build()

    // Connect source to sink
    source.ToSink(sink)

    // Wait for the value to be received
    _ = sink.Wait()

    // Get the result
    result := sink.Result()
    fmt.Printf("Result: %d\n", result)
    // Output: Result: 42
}
```

### With Pipeline Processing

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    source := sources.Single(func() (int, error) {
        return 1000, nil
    }).Build()

    // Transform the value
    sumFlow := flows.
        Map(func(x int) (int, error) {
            return x + 1, nil
        }).
        Build()

    sink := sinks.Single[int]().Build()

    // Connect the pipeline
    source.ToFlow(sumFlow).ToSink(sink)

    // Wait for processing to complete
    _ = sink.Wait()

    fmt.Println(sink.Result())
    // Output: 1001
}
```

### With Filter to Get First Match

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a source with multiple values
    source := sources.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).Build()

    // Filter to get the first even number
    filter := flows.Filter(func(x int) (bool, error) {
        return x%2 == 0, nil
    }).Build()

    sink := sinks.Single[int]().Build()

    // Connect the pipeline
    source.ToFlow(filter).ToSink(sink)

    // Wait for the first match
    _ = sink.Wait()

    firstEven := sink.Result()
    fmt.Printf("First even number: %d\n", firstEven)
    // Output: First even number: 2
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
    ctx, cancel := context.WithTimeout(
        context.Background(),
        100*time.Millisecond,
    )
    defer cancel()

    // Create a single sink with context
    sink := sinks.Single[int]().Context(ctx).Build()

    // Create a source that might take longer
    source := sources.Single(func() (int, error) {
        time.Sleep(200 * time.Millisecond)
        return 42, nil
    }).Build()

    source.ToSink(sink)

    // Wait will return when context is cancelled
    _ = sink.Wait()

    result := sink.Result()
    fmt.Printf("Result: %d\n", result)
    // Output: Result: 0 (zero value if context cancelled before value received)
}
```

### Getting Result Without Waiting

```go
package main

import (
    "fmt"
    "time"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    sink := sinks.Single[string]().Build()

    source := sources.Single(func() (string, error) {
        return "Hello, World!", nil
    }).Build()

    source.ToSink(sink)

    // You can check the result without waiting
    // (though it might be the zero value if not received yet)
    time.Sleep(50 * time.Millisecond)
    result := sink.Result()
    fmt.Printf("Result: %s\n", result)
    // Output: Result: Hello, World!
}
```

### Thread-Safe Concurrent Access

```go
package main

import (
    "fmt"
    "sync"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    sink := sinks.Single[int]().Build()

    source := sources.Single(func() (int, error) {
        return 100, nil
    }).Build()

    source.ToSink(sink)

    _ = sink.Wait()

    // Multiple goroutines can safely access Result()
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            result := sink.Result()
            fmt.Printf("Goroutine result: %d\n", result)
        }()
    }
    wg.Wait()
    // All goroutines will safely read the same value: 100
}
```

### With Reduce to Get Final Result

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a source with multiple values
    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()

    // Use reduce to sum all values
    reduceFlow := flows.Reduce(func(result, value int, index uint) (
        int, error,
    ) {
        return result + value, nil
    }, 0).Build()

    // Get the single final result
    sink := sinks.Single[int]().Build()

    source.ToFlow(reduceFlow).ToSink(sink)

    _ = sink.Wait()

    sum := sink.Result()
    fmt.Printf("Sum: %d\n", sum)
    // Output: Sum: 15
}
```

### Handling Zero Values

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    sink := sinks.Single[int]().Build()

    // Send zero value
    source := sources.Single(func() (int, error) {
        return 0, nil
    }).Build()

    source.ToSink(sink)

    _ = sink.Wait()

    result := sink.Result()
    fmt.Printf("Result: %d\n", result)
    // Output: Result: 0

    // Note: Zero value is a valid result, so you need to use Wait()
    // to distinguish between "not received yet" and "received zero value"
}
```

## Builder Methods

### `Single[T any]() *SingleSinkBuilder[T]`

Creates a new `SingleSinkBuilder` for building a `SingleSink`.

### `Context(ctx context.Context) *SingleSinkBuilder[T]`

Sets the context for cancellation and timeout handling.

### `Build() *SingleSink[T]`

Creates and starts the `SingleSink`.

## Methods

### `In() chan<- T`

Returns the input channel where values should be sent.

### `Wait() error`

Waits for the `SingleSink` to finish receiving the value. This method blocks
until the upstream closes or the context is cancelled. Returns `nil` when
complete.

### `Result() T`

Returns the received value. This method is thread-safe and can be called
concurrently. If no value has been received yet, it returns the zero value
of type `T`.

## Performance Considerations

1. **Single Value Only**: This sink is designed for exactly one value. If
   multiple values are sent, only the last one is kept, which may not be the
   intended behavior.

2. **Thread Safety**: The `Result()` method is thread-safe but involves a
   read lock. For high-frequency access, consider caching the result after
   `Wait()` completes.

3. **Memory Usage**: The sink holds only one value in memory, making it very
   memory-efficient.

4. **Blocking Behavior**: The `Wait()` method blocks until a value is received
   or the context is cancelled. Use context timeouts to prevent indefinite
   blocking.

## Important Notes

1. **Single Value Expectation**: This sink expects exactly ONE value from
   upstream. If multiple values are sent, only the last one is kept, and the
   first value(s) will be lost.

2. **Wait for Completion**: Always call `Wait()` before accessing `Result()`
   if you need to ensure the value has been received. Without waiting, you
   might get the zero value.

3. **Zero Value Handling**: The zero value is a valid result. Use `Wait()` to
   distinguish between "not received yet" (zero value before wait completes)
   and "received zero value" (zero value after wait completes).

4. **Context Cancellation**: When the context is cancelled, processing stops
   immediately. If no value was received before cancellation, `Result()` will
   return the zero value.

5. **Thread Safety**: The `Result()` method is thread-safe and can be called
   from multiple goroutines concurrently.

6. **Channel Closure**: The sink closes the input channel after receiving the
   value, so make sure upstream components handle channel closure gracefully.

7. **Empty Stream**: If no values are sent and the channel is closed, the
   result will be the zero value of type `T`.

8. **WaitableSink Interface**: This sink implements `WaitableSink`, making it
   suitable for synchronization points in your pipeline where you need to wait
   for a single result.
