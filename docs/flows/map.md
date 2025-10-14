# MapFlow

The `MapFlow` is a data transformation operator that transforms each value in
the input stream using a provided function. It's one of the most fundamental
and powerful operators in functional programming, allowing you to convert data
from one type to another or apply transformations.

## Design

The `MapFlow` implements the `primitives.Flow[IN, OUT]` interface and provides
a fluent builder API for configuration. It applies a transformation function to
each incoming value and emits the result to the output stream.

### Key Features

- **Type Transformation**: Can convert between different types (IN -> OUT)
- **Parallel Processing**: Supports configurable parallelism for CPU-intensive
  transformations
- **Error Handling**: Supports custom error handlers for transformation failures
- **Context Support**: Respects context cancellation
- **Flexible Mapping**: Any function that takes IN and returns OUT can be used

## Graphical Representation

```
Input Stream: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->

    MapFlow (f(x) = x * 2)
         |
         v
Output Stream: -- 2 -- 4 -- 6 -- 8 -- 10 -- | -->
```

## Usage Examples

### Basic Transformation

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

    // Double each number
    mapper := flows.
        Map(func(x int) (int, error) {
            return x * 2, nil
        }).
        Build()

    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()
    sink := sinks.Channel(outputCh).Build()

    // Connect the pipeline
    source.ToFlow(mapper).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Doubled:", value)
    }
    // Output:
    // Doubled: 2
    // Doubled: 4
    // Doubled: 6
    // Doubled: 8
    // Doubled: 10
}
```

### Type Conversion

```go
package main

import (
    "fmt"
    "strconv"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan string)

    // Convert integers to strings
    mapper := flows.
        Map(func(x int) (string, error) {
            return strconv.Itoa(x), nil
        }).
        Build()

    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(mapper).ToSink(sink)

    for value := range outputCh {
        fmt.Println("String:", value)
    }
    // Output:
    // String: 1
    // String: 2
    // String: 3
    // String: 4
    // String: 5
}
```

### String Processing

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

    // Convert to uppercase and add prefix
    mapper := flows.
        Map(func(s string) (string, error) {
            return "PREFIX: " + strings.ToUpper(s), nil
        }).
        Build()

    words := []string{"hello", "world", "golang"}
    source := sources.Slice(words).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(mapper).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Processed:", value)
    }
    // Output:
    // Processed: PREFIX: HELLO
    // Processed: PREFIX: WORLD
    // Processed: PREFIX: GOLANG
}
```

### With Parallel Processing

```go
package main

import (
    "fmt"
    "time"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan int)

    // CPU-intensive transformation with parallelism
    mapper := flows.
        Map(func(x int) (int, error) {
            // Simulate CPU-intensive work
            time.Sleep(100 * time.Millisecond)
            return x * x, nil // Square the number
        }).
        Parallelism(4). // Use 4 parallel workers
        Build()

    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8}
    source := sources.Slice(numbers).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(mapper).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Squared:", value)
    }
    // Output (order may vary due to parallelism):
    // Squared: 1
    // Squared: 4
    // Squared: 9
    // Squared: 16
    // Squared: 25
    // Squared: 36
    // Squared: 49
    // Squared: 64
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

    // Transformation that might fail
    mapper := flows.
        Map(func(x int) (int, error) {
            if x == 0 {
                return 0, errors.New("division by zero")
            }
            return 100 / x, nil
        }).
        ErrorHandler(func(err error) {
            fmt.Printf("Map error: %v\n", err)
        }).
        Build()

    numbers := []int{10, 5, 0, 2, 1}
    source := sources.Slice(numbers).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(mapper).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Result:", value)
    }
    // Output:
    // Result: 10
    // Result: 20
    // Map error: division by zero
    // (pipeline stops due to error)
}
```

### Complex Data Transformation

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

type PersonInfo struct {
    Name        string
    Age         int
    IsAdult     bool
    AgeGroup    string
}

func main() {
    outputCh := make(chan PersonInfo)

    // Transform Person to PersonInfo
    mapper := flows.
        Map(func(p Person) (PersonInfo, error) {
            isAdult := p.Age >= 18
            var ageGroup string
            switch {
            case p.Age < 18:
                ageGroup = "minor"
            case p.Age < 65:
                ageGroup = "adult"
            default:
                ageGroup = "senior"
            }

            return PersonInfo{
                Name:     p.Name,
                Age:      p.Age,
                IsAdult:  isAdult,
                AgeGroup: ageGroup,
            }, nil
        }).
        Build()

    people := []Person{
        {Name: "Alice", Age: 25},
        {Name: "Bob", Age: 17},
        {Name: "Charlie", Age: 70},
    }

    source := sources.Slice(people).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(mapper).ToSink(sink)

    for info := range outputCh {
        fmt.Printf("%s: %d years old, %s, adult: %t\n",
            info.Name, info.Age, info.AgeGroup, info.IsAdult)
    }
    // Output:
    // Alice: 25 years old, adult, adult: true
    // Bob: 17 years old, minor, adult: false
    // Charlie: 70 years old, senior, adult: true
}
```

### JSON Processing

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    outputCh := make(chan []byte)

    // Convert User to JSON bytes
    mapper := flows.
        Map(func(u User) ([]byte, error) {
            return json.Marshal(u)
        }).
        ErrorHandler(func(err error) {
            fmt.Printf("JSON marshal error: %v\n", err)
        }).
        Build()

    users := []User{
        {ID: 1, Name: "Alice"},
        {ID: 2, Name: "Bob"},
        {ID: 3, Name: "Charlie"},
    }

    source := sources.Slice(users).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(mapper).ToSink(sink)

    for jsonData := range outputCh {
        fmt.Println("JSON:", string(jsonData))
    }
    // Output:
    // JSON: {"id":1,"name":"Alice"}
    // JSON: {"id":2,"name":"Bob"}
    // JSON: {"id":3,"name":"Charlie"}
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
    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
    defer cancel()

    outputCh := make(chan int)

    mapper := flows.
        Map(func(x int) (int, error) {
            // Simulate some processing time
            time.Sleep(100 * time.Millisecond)
            return x * 2, nil
        }).
        Context(ctx).
        Build()

    // Create a large dataset
    largeData := make([]int, 100)
    for i := range largeData {
        largeData[i] = i
    }

    source := sources.Slice(largeData).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(mapper).ToSink(sink)

    count := 0
    for value := range outputCh {
        count++
        fmt.Printf("Processed %d values, current: %d\n", count, value)
    }

    fmt.Printf("Total processed before timeout: %d\n", count)
}
```

## Builder Methods

### `Map[IN, OUT any](fn func(IN) (OUT, error)) *MapBuilder[IN, OUT]`

Creates a new `MapBuilder` with the provided transformation function.

### `Context(ctx context.Context) *MapBuilder[IN, OUT]`

Sets the context for cancellation and timeout handling.

### `Parallelism(p uint) *MapBuilder[IN, OUT]`

Sets the parallelism level for concurrent processing. Default is 1 (sequential).

### `ErrorHandler(handler func(error)) *MapBuilder[IN, OUT]`

Sets a custom error handler for transformation failures.

### `BufferSize(size uint) *MapBuilder[IN, OUT]`

Sets the buffer size for the input and output channels.

### `Build() *MapFlow[IN, OUT]`

Creates and starts the `MapFlow`.

## Performance Considerations

1. **Parallelism**: Use parallelism for CPU-intensive transformations. The
   optimal level depends on your CPU cores and the nature of the work.

2. **Function Complexity**: Keep transformation functions efficient. Complex
   operations can become bottlenecks in the pipeline.

3. **Buffer Size**: Larger buffer sizes can improve performance by reducing
   blocking between flows.

4. **Error Handling**: Always provide error handlers for production code to
   handle transformation failures gracefully.

## Important Notes

1. **Single Use**: Each `MapFlow` can only be used once. Attempting to connect
   it to multiple downstream flows will panic.

2. **Error Handling**: If the transformation function returns an error, the
   map will stop processing and close the output channel.

3. **Context Cancellation**: The map respects context cancellation and will
   stop processing if the context is cancelled.

4. **Type Safety**: The map provides compile-time type safety between input
   and output types.

5. **Order Preservation**: With parallelism > 1, the order of output values
   may not match the input order. Use sequential processing if order matters.
