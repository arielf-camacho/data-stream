# PassThroughFlow

The `PassThroughFlow` is a data forwarding operator that passes values from
the input stream to the output stream with optional type conversion. It's
useful for creating pipeline segments, type conversions, and as a building
block for more complex flows.

## Design

The `PassThroughFlow` implements the `primitives.Flow[IN, OUT]` interface and
provides a fluent builder API for configuration. It can either pass values
through unchanged (when IN == OUT) or convert between types using a custom
conversion function.

### Key Features

- **Type Conversion**: Can convert between different types using a custom function
- **Identity Pass-Through**: When no conversion is needed, acts as a simple
  pass-through
- **Error Handling**: Supports custom error handlers for conversion failures
- **Context Support**: Respects context cancellation
- **Flexible Conversion**: Any conversion function can be provided

## Graphical Representation

```
Input Stream: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->

    PassThroughFlow (with optional conversion)
         |
         v
Output Stream: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->
```

## Usage Examples

### Basic Pass-Through

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

    // Create a simple pass-through flow
    passThrough := flows.
        PassThrough[int, int]().
        Build()

    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()
    sink := sinks.Channel(outputCh).Build()

    // Connect the pipeline
    source.ToFlow(passThrough).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Passed through:", value)
    }
    // Output:
    // Passed through: 1
    // Passed through: 2
    // Passed through: 3
    // Passed through: 4
    // Passed through: 5
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
    converter := flows.
        PassThrough[int, string]().
        Convert(func(x int) (string, error) {
            return strconv.Itoa(x), nil
        }).
        Build()

    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(converter).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Converted:", value)
    }
    // Output:
    // Converted: 1
    // Converted: 2
    // Converted: 3
    // Converted: 4
    // Converted: 5
}
```

### Complex Type Conversion

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

type PersonSummary struct {
    Name        string
    Age         int
    IsAdult     bool
    AgeCategory string
}

func main() {
    outputCh := make(chan PersonSummary)

    // Convert Person to PersonSummary
    converter := flows.
        PassThrough[Person, PersonSummary]().
        Convert(func(p Person) (PersonSummary, error) {
            isAdult := p.Age >= 18
            var category string
            switch {
            case p.Age < 18:
                category = "minor"
            case p.Age < 65:
                category = "adult"
            default:
                category = "senior"
            }

            return PersonSummary{
                Name:        p.Name,
                Age:         p.Age,
                IsAdult:     isAdult,
                AgeCategory: category,
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

    source.ToFlow(converter).ToSink(sink)

    for summary := range outputCh {
        fmt.Printf("%s: %d years old, %s, adult: %t\n",
            summary.Name, summary.Age, summary.AgeCategory, summary.IsAdult)
    }
    // Output:
    // Alice: 25 years old, adult, adult: true
    // Bob: 17 years old, minor, adult: false
    // Charlie: 70 years old, senior, adult: true
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

    // Conversion that might fail
    converter := flows.
        PassThrough[int, int]().
        Convert(func(x int) (int, error) {
            if x == 0 {
                return 0, errors.New("cannot process zero")
            }
            return 100 / x, nil
        }).
        ErrorHandler(func(err error) {
            fmt.Printf("Conversion error: %v\n", err)
        }).
        Build()

    numbers := []int{10, 5, 0, 2, 1}
    source := sources.Slice(numbers).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(converter).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Result:", value)
    }
    // Output:
    // Result: 10
    // Result: 20
    // Conversion error: cannot process zero
    // (pipeline stops due to error)
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
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    outputCh := make(chan int)

    // Pass-through with context
    passThrough := flows.
        PassThrough[int, int]().
        Context(ctx).
        Build()

    // Create a large dataset
    largeData := make([]int, 1000)
    for i := range largeData {
        largeData[i] = i
    }

    source := sources.Slice(largeData).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(passThrough).ToSink(sink)

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

### With Custom Buffer Size

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

    // Pass-through with custom buffer size
    passThrough := flows.
        PassThrough[int, int]().
        BufferSize(100). // Larger buffer for better performance
        Build()

    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(passThrough).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Buffered pass-through:", value)
    }
}
```

### JSON Conversion

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
    jsonConverter := flows.
        PassThrough[User, []byte]().
        Convert(func(u User) ([]byte, error) {
            return json.Marshal(u)
        }).
        ErrorHandler(func(err error) {
            fmt.Printf("JSON conversion error: %v\n", err)
        }).
        Build()

    users := []User{
        {ID: 1, Name: "Alice"},
        {ID: 2, Name: "Bob"},
        {ID: 3, Name: "Charlie"},
    }

    source := sources.Slice(users).Build()
    sink := sinks.Channel(outputCh).Build()

    source.ToFlow(jsonConverter).ToSink(sink)

    for jsonData := range outputCh {
        fmt.Println("JSON:", string(jsonData))
    }
    // Output:
    // JSON: {"id":1,"name":"Alice"}
    // JSON: {"id":2,"name":"Bob"}
    // JSON: {"id":3,"name":"Charlie"}
}
```

### Chaining Multiple Pass-Throughs

```go
package main

import (
    "fmt"
    "strconv"
    "strings"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan string)

    // Chain multiple conversions
    intToString := flows.
        PassThrough[int, string]().
        Convert(func(x int) (string, error) {
            return strconv.Itoa(x), nil
        }).
        Build()

    stringToUpper := flows.
        PassThrough[string, string]().
        Convert(func(s string) (string, error) {
            return strings.ToUpper(s), nil
        }).
        Build()

    addPrefix := flows.
        PassThrough[string, string]().
        Convert(func(s string) (string, error) {
            return "NUMBER: " + s, nil
        }).
        Build()

    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()
    sink := sinks.Channel(outputCh).Build()

    // Chain the conversions
    source.ToFlow(intToString).ToFlow(stringToUpper).ToFlow(addPrefix).ToSink(sink)

    for value := range outputCh {
        fmt.Println("Final result:", value)
    }
    // Output:
    // Final result: NUMBER: 1
    // Final result: NUMBER: 2
    // Final result: NUMBER: 3
    // Final result: NUMBER: 4
    // Final result: NUMBER: 5
}
```

## Builder Methods

### `PassThrough[IN any, OUT any]() *PassThroughBuilder[IN, OUT]`

Creates a new `PassThroughBuilder` with default type conversion.

### `Context(ctx context.Context) *PassThroughBuilder[IN, OUT]`

Sets the context for cancellation and timeout handling.

### `BufferSize(size uint) *PassThroughBuilder[IN, OUT]`

Sets the buffer size for the input and output channels.

### `Convert(convert func(IN) (OUT, error)) *PassThroughBuilder[IN, OUT]`

Sets a custom conversion function. If not provided, uses default type assertion.

### `ErrorHandler(handler func(error)) *PassThroughBuilder[IN, OUT]`

Sets a custom error handler for conversion failures.

### `Build() *PassThroughFlow[IN, OUT]`

Creates and starts the `PassThroughFlow`.

## Performance Considerations

1. **Conversion Complexity**: Keep conversion functions simple and fast. Complex
   operations can become bottlenecks in the pipeline.

2. **Buffer Size**: Larger buffer sizes can improve performance by reducing
   blocking between flows.

3. **Type Assertions**: The default conversion uses type assertions, which are
   fast but will panic if types are incompatible.

## Important Notes

1. **Single Use**: Each `PassThroughFlow` can only be used once. Attempting to
   connect it to multiple downstream flows will panic.

2. **Error Handling**: If the conversion function returns an error, the flow
   will stop processing and close the output channel.

3. **Context Cancellation**: The flow respects context cancellation and will
   stop processing if the context is cancelled.

4. **Type Safety**: The flow provides compile-time type safety between input
   and output types.

5. **Default Conversion**: When no custom conversion is provided, the flow uses
   type assertion (`any(v).(OUT)`), which will panic if the types are
   incompatible.
