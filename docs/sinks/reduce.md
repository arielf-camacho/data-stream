# ReduceSink

The `ReduceSink` is a data sink that reduces incoming values to a single result using a custom reduce function. It's perfect for aggregating data streams into summary statistics, totals, or any other single computed value.

## Design

The `ReduceSink` implements the `primitives.Sink[T]` interface and provides a fluent builder API for configuration. It processes all incoming values sequentially, applying a reduce function that accumulates a result, and provides thread-safe access to the final result.

### Key Features

- **Accumulation**: Reduces a stream of values to a single result
- **Thread-Safe**: Safe concurrent access to the result via `Result()` method
- **Index Tracking**: The reduce function receives the current index for each value
- **Error Handling**: Custom error handlers for reduce function failures
- **Context Support**: Respects context cancellation
- **Initial Value**: Configurable initial value for the reduction

## Graphical Representation

```
Input Stream: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->

    ReduceSink f(result, value, index) = result + value
         |
         v
Final Result: 15 (sum of all values)
```

## Usage Examples

### Basic Sum Reduction

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a reduce sink that sums all values
    sink := sinks.Reduce(func(result, value int, index uint) (int, error) {
        return result + value, nil
    }, 0).Build()

    // Create a source with numbers
    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()

    // Connect source to sink
    source.ToSink(sink)

    // Get the final result
    sum := sink.Result()
    fmt.Printf("Sum: %d\n", sum)
    // Output: Sum: 15
}
```

### Product Reduction

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a reduce sink that multiplies all values
    sink := sinks.Reduce(func(result, value int, index uint) (int, error) {
        return result * value, nil
    }, 1).Build() // Start with 1 for multiplication

    source := sources.Slice([]int{2, 3, 4}).Build()
    source.ToSink(sink)

    product := sink.Result()
    fmt.Printf("Product: %d\n", product)
    // Output: Product: 24
}
```

### Finding Maximum Value

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a reduce sink that finds the maximum value
    sink := sinks.Reduce(func(result, value int, index uint) (int, error) {
        if value > result {
            return value, nil
        }
        return result, nil
    }, 0).Build() // Start with 0

    source := sources.Slice([]int{5, 2, 8, 1, 9, 3}).Build()
    source.ToSink(sink)

    max := sink.Result()
    fmt.Printf("Maximum: %d\n", max)
    // Output: Maximum: 9
}
```

### String Concatenation

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a reduce sink that concatenates strings
    sink := sinks.Reduce(func(result, value string, index uint) (string, error) {
        return result + value, nil
    }, "").Build() // Start with empty string

    source := sources.Slice([]string{"Hello", " ", "World", "!"}).Build()
    source.ToSink(sink)

    message := sink.Result()
    fmt.Printf("Message: %s\n", message)
    // Output: Message: Hello World!
}
```

### Counting Values

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a reduce sink that counts values
    sink := sinks.Reduce(func(result, value int, index uint) (int, error) {
        return result + 1, nil
    }, 0).Build()

    source := sources.Slice([]int{10, 20, 30, 40, 50}).Build()
    source.ToSink(sink)

    count := sink.Result()
    fmt.Printf("Count: %d\n", count)
    // Output: Count: 5
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
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    // Create a reduce sink with context
    sink := sinks.Reduce(func(result, value int, index uint) (int, error) {
        return result + value, nil
    }, 0).Context(ctx).Build()

    // Create a large data source
    largeData := make([]int, 1000)
    for i := range largeData {
        largeData[i] = i
    }

    source := sources.Slice(largeData).Build()
    source.ToSink(sink)

    // Wait a bit for processing
    time.Sleep(200 * time.Millisecond)

    result := sink.Result()
    fmt.Printf("Partial sum before timeout: %d\n", result)
}
```

### With Error Handling

```go
package main

import (
    "errors"
    "fmt"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a reduce sink with error handling
    sink := sinks.Reduce(func(result, value int, index uint) (int, error) {
        if value < 0 {
            return 0, errors.New("negative values not allowed")
        }
        return result + value, nil
    }, 0).ErrorHandler(func(err error, index uint, value, result int) {
        fmt.Printf("Error at index %d: %v (value: %d, result so far: %d)\n",
            index, err, value, result)
    }).Build()

    source := sources.Slice([]int{1, 2, -3, 4, 5}).Build()
    source.ToSink(sink)

    result := sink.Result()
    fmt.Printf("Sum before error: %d\n", result)
    // Output:
    // Error at index 2: negative values not allowed (value: -3, result so far: 3)
    // Sum before error: 3
}
```

### Complex Aggregation

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

type Stats struct {
    Sum   int
    Count int
    Max   int
    Min   int
}

func main() {
    // Create a reduce sink that calculates statistics
    sink := sinks.Reduce(func(result Stats, value int, index uint) (Stats, error) {
        result.Sum += value
        result.Count++

        if result.Count == 1 {
            result.Max = value
            result.Min = value
        } else {
            if value > result.Max {
                result.Max = value
            }
            if value < result.Min {
                result.Min = value
            }
        }

        return result, nil
    }, Stats{}).Build()

    source := sources.Slice([]int{5, 2, 8, 1, 9, 3, 7, 4, 6}).Build()
    source.ToSink(sink)

    stats := sink.Result()
    fmt.Printf("Statistics:\n")
    fmt.Printf("  Sum: %d\n", stats.Sum)
    fmt.Printf("  Count: %d\n", stats.Count)
    fmt.Printf("  Max: %d\n", stats.Max)
    fmt.Printf("  Min: %d\n", stats.Min)
    fmt.Printf("  Average: %.2f\n", float64(stats.Sum)/float64(stats.Count))
    // Output:
    // Statistics:
    //   Sum: 45
    //   Count: 9
    //   Max: 9
    //   Min: 1
    //   Average: 5.00
}
```

### Using Index Parameter

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a reduce sink that uses the index parameter
    sink := sinks.Reduce(func(result, value int, index uint) (int, error) {
        // Weight each value by its position (0-based index)
        return result + (value * int(index)), nil
    }, 0).Build()

    source := sources.Slice([]int{10, 20, 30, 40}).Build()
    source.ToSink(sink)

    weightedSum := sink.Result()
    fmt.Printf("Weighted sum: %d\n", weightedSum)
    // Output: Weighted sum: 140 (10*0 + 20*1 + 30*2 + 40*3 = 0 + 20 + 60 + 120)
}
```

### With Buffer Size

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a reduce sink with custom buffer size
    sink := sinks.Reduce(func(result, value int, index uint) (int, error) {
        return result + value, nil
    }, 0).BufferSize(100).Build() // Larger buffer for better performance

    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()
    source.ToSink(sink)

    sum := sink.Result()
    fmt.Printf("Sum: %d\n", sum)
    // Output: Sum: 15
}
```

### Integration with Pipeline

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a pipeline: source -> filter -> reduce
    source := sources.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).Build()

    // Filter even numbers
    filter := flows.Filter(func(x int) (bool, error) {
        return x%2 == 0, nil
    }).Build()

    // Sum the filtered values
    sink := sinks.Reduce(func(result, value int, index uint) (int, error) {
        return result + value, nil
    }, 0).Build()

    // Connect the pipeline
    source.ToFlow(filter).ToSink(sink)

    // Get the sum of even numbers
    sum := sink.Result()
    fmt.Printf("Sum of even numbers: %d\n", sum)
    // Output: Sum of even numbers: 30 (2+4+6+8+10)
}
```

## Builder Methods

### `Reduce[IN, OUT any](fn func(result OUT, value IN, index uint) (OUT, error), initial OUT) *ReduceSinkBuilder[IN, OUT]`

Creates a new `ReduceSinkBuilder` with the provided reduce function and initial value.

**Parameters:**

- `fn`: The reduce function that takes the current result, incoming value, and index, returning the new result and any error
- `initial`: The initial value for the reduction

### `Context(ctx context.Context) *ReduceSinkBuilder[IN, OUT]`

Sets the context for cancellation and timeout handling.

### `BufferSize(size uint) *ReduceSinkBuilder[IN, OUT]`

Sets the buffer size for the input channel. This affects how many values can be buffered before blocking the upstream flow.

### `ErrorHandler(handler func(error, uint, IN, OUT)) *ReduceSinkBuilder[IN, OUT]`

Sets the error handler that will be called if the reduce function returns an error. The handler receives the error, index, value that caused the error, and the result so far.

### `Build() *ReduceSink[IN, OUT]`

Creates and starts the `ReduceSink`.

## Methods

### `In() chan<- IN`

Returns the input channel where values should be sent.

### `Result() OUT`

Returns the current result of the reduction. This method is thread-safe and can be called concurrently.

## Performance Considerations

1. **Buffer Size**: Larger buffer sizes can improve performance by reducing blocking between the sink and upstream flows.

2. **Reduce Function**: Keep the reduce function simple and fast, as it's called for every incoming value.

3. **Result Access**: The `Result()` method is thread-safe but involves a read lock. For high-frequency access, consider caching the result.

4. **Memory Usage**: The sink holds the accumulated result in memory, so be mindful of the result type's memory footprint.

## Important Notes

1. **Sequential Processing**: Values are processed sequentially in the order they arrive. The reduce function is called once per value.

2. **Index Parameter**: The index parameter starts at 0 and increments for each value processed.

3. **Error Handling**: If the reduce function returns an error, processing stops and the error handler is called (if provided).

4. **Context Cancellation**: When the context is cancelled, processing stops immediately and the current result is preserved.

5. **Thread Safety**: The `Result()` method is thread-safe and can be called from multiple goroutines.

6. **Initial Value**: The initial value is used as the starting point for the reduction. Choose an appropriate value (e.g., 0 for sum, 1 for product, empty string for concatenation).

7. **Empty Stream**: If no values are processed, the result will be the initial value.

8. **Single Result**: The sink produces a single result value, not a stream. Use `Result()` to access the final accumulated value.
