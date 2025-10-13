# FilterFlow

The `FilterFlow` is a data transformation operator that filters values from
the input stream based on a predicate function. Only values for which the
predicate returns `true` are passed through to the output stream.

## Design

The `FilterFlow` implements the `primitives.Flow[T, T]` interface and provides
a fluent builder API for configuration. It evaluates each incoming value
against a predicate function and only forwards values that pass the test.

### Key Features

- **Predicate-Based Filtering**: Uses a function to determine which values to
  pass through
- **Error Handling**: Supports custom error handlers for predicate failures
- **Context Support**: Respects context cancellation
- **Type Preservation**: Maintains the same type for input and output

## Graphical Representation

```
Input Stream: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->

    FilterFlow (predicate: x > 2)
         |
         v
Output Stream: ------------ 3 -- 4 -- 5 -- | -->
```

## Usage Examples

### Basic Filtering

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

    // Create a filter that only passes even numbers
    filter := flows.
        Filter(func(x int) (bool, error) {
            return x%2 == 0, nil
        }).
        Build()

    source := sources.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).Build()
    sink := sinks.Channel(outputCh).Build()

    // Connect the pipeline
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

### String Filtering

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

    // Filter strings that start with "A"
    filter := flows.
        Filter(func(s string) (bool, error) {
            return strings.HasPrefix(s, "A"), nil
        }).
        Build()

    words := []string{"Apple", "Banana", "Apricot", "Cherry", "Avocado"}
    source := sources.Slice(words).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(filter).ToSink(sink)

    for value := range outputCh {
        fmt.Println("A-word:", value)
    }
    // Output:
    // A-word: Apple
    // A-word: Apricot
    // A-word: Avocado
}
```

### Complex Filtering Logic

```go
package main

import (
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

type Person struct {
    Name string
    Age  int
}

func main() {
    outputCh := make(chan Person)

    // Filter adults (age >= 18)
    filter := flows.
        Filter(func(p Person) (bool, error) {
            return p.Age >= 18, nil
        }).
        Build()

    people := []Person{
        {Name: "Alice", Age: 25},
        {Name: "Bob", Age: 17},
        {Name: "Charlie", Age: 30},
        {Name: "Diana", Age: 16},
        {Name: "Eve", Age: 22},
    }

    source := sources.Slice(people).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(filter).ToSink(sink)

    for person := range outputCh {
        fmt.Printf("Adult: %s (age %d)\n", person.Name, person.Age)
    }
    // Output:
    // Adult: Alice (age 25)
    // Adult: Charlie (age 30)
    // Adult: Eve (age 22)
}
```

### With Error Handling

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

    // Filter that might fail
    filter := flows.
        Filter(func(x int) (bool, error) {
            if x < 0 {
                return false, errors.New("negative numbers not allowed")
            }
            return x%2 == 0, nil
        }).
        ErrorHandler(func(err error) {
            fmt.Printf("Filter error: %v\n", err)
        }).
        Build()

    numbers := []int{1, 2, -3, 4, 5}
    source := sources.Slice(numbers).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(filter).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Filtered value:", value)
    }
    // Output:
    // Filtered value: 2
    // Filtered value: 4
    // Filter error: negative numbers not allowed
    // (pipeline stops due to error)
}
```

### Multiple Filters in Sequence

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

    // First filter: even numbers
    evenFilter := flows.
        Filter(func(x int) (bool, error) {
            return x%2 == 0, nil
        }).
        Build()

    // Second filter: greater than 5
    greaterThan5Filter := flows.
        Filter(func(x int) (bool, error) {
            return x > 5, nil
        }).
        Build()

    source := sources.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).Build()
    sink := sinks.Channel(outputCh).Build()

    // Chain the filters
    source.ToFlow(evenFilter).ToFlow(greaterThan5Filter).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Even and > 5:", value)
    }
    // Output:
    // Even and > 5: 6
    // Even and > 5: 8
    // Even and > 5: 10
}
```

### With Context Cancellation

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
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    outputCh := make(chan int)

    filter := flows.
        Filter(func(x int) (bool, error) {
            // Simulate some processing time
            time.Sleep(50 * time.Millisecond)
            return x%2 == 0, nil
        }).
        Context(ctx).
        Build()

    // Create a large dataset
    largeData := make([]int, 1000)
    for i := range largeData {
        largeData[i] = i
    }

    source := sources.Slice(largeData).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(filter).ToSink(sink)

    count := 0
    for value := range outputCh {
        count++
        fmt.Printf("Processed %d values, current: %d\n", count, value)
    }

    fmt.Printf("Total processed before timeout: %d\n", count)
}
```

### Custom Buffer Size

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

    filter := flows.
        Filter(func(x int) (bool, error) {
            return x > 0, nil
        }).
        BufferSize(100). // Larger buffer for better performance
        Build()

    source := sources.Slice([]int{-2, -1, 0, 1, 2, 3, 4, 5}).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(filter).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Positive number:", value)
    }
    // Output:
    // Positive number: 1
    // Positive number: 2
    // Positive number: 3
    // Positive number: 4
    // Positive number: 5
}
```

## Builder Methods

### `Filter[T any](predicate func(T) (bool, error)) *FilterBuilder[T]`

Creates a new `FilterBuilder` with the provided predicate function.

### `Context(ctx context.Context) *FilterBuilder[T]`

Sets the context for cancellation and timeout handling.

### `ErrorHandler(handler func(error)) *FilterBuilder[T]`

Sets a custom error handler for predicate execution errors.

### `BufferSize(size uint) *FilterBuilder[T]`

Sets the buffer size for the input and output channels.

### `Build() *FilterFlow[T]`

Creates and starts the `FilterFlow`.

## Performance Considerations

1. **Predicate Complexity**: Keep predicate functions simple and fast. Complex
   operations in predicates can slow down the entire pipeline.

2. **Buffer Size**: For high-throughput scenarios, use larger buffer sizes to
   reduce blocking between flows.

3. **Early Termination**: Consider using context cancellation for long-running
   filters that might need to be stopped early.

## Important Notes

1. **Single Use**: Each `FilterFlow` can only be used once. Attempting to
   connect it to multiple downstream flows will panic.

2. **Error Handling**: If the predicate function returns an error, the filter
   will stop processing and close the output channel.

3. **Context Cancellation**: The filter respects context cancellation and will
   stop processing if the context is cancelled.

4. **Type Safety**: The filter maintains the same type for input and output,
   ensuring type safety throughout the pipeline.

5. **Order Preservation**: The filter preserves the order of values that pass
   the predicate test.
