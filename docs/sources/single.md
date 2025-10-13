# SingleSource

The `SingleSource` is a data source that emits exactly one value from a
user-provided function. It's useful for generating single values, API calls,
or any operation that produces one result.

## Design

The `SingleSource` implements the `primitives.Source[T]` interface and provides
a fluent builder API for configuration. It uses a goroutine to execute the
provided function and emit the result through a channel.

### Key Features

- **Single Value Emission**: Emits exactly one value and then closes
- **Error Handling**: Supports custom error handlers for function failures
- **Context Support**: Respects context cancellation
- **Thread Safety**: Uses atomic operations to prevent multiple activations

## Graphical Representation

```
Input Function: f() -> (T, error)

    f() execution
         |
         v
    [Single Value]
         |
         v
    -- value -- | --> (channel closes)
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
    // Create a single source that returns a string
    source := sources.
        Single(func() (string, error) {
            return "Hello, World!", nil
        }).
        Build()

    // Read the value
    for value := range source.Out() {
        fmt.Println("Received:", value)
    }
}
```

### With Error Handling

```go
package main

import (
    "errors"
    "fmt"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    source := sources.
        Single(func() (int, error) {
            // Simulate an error
            return 0, errors.New("something went wrong")
        }).
        ErrorHandler(func(err error) {
            fmt.Printf("Error occurred: %v\n", err)
        }).
        Build()

    // The source will emit no values due to the error
    for value := range source.Out() {
        fmt.Println("This won't be printed:", value)
    }
}
```

### With Context

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    source := sources.
        Single(func() (string, error) {
            // Simulate a long-running operation
            time.Sleep(3 * time.Second)
            return "This won't be reached", nil
        }).
        Context(ctx).
        Build()

    // The context will cancel before the function completes
    for value := range source.Out() {
        fmt.Println("Received:", value)
    }
}
```

### API Call Example

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    source := sources.
        Single(func() (string, error) {
            resp, err := http.Get("https://api.example.com/data")
            if err != nil {
                return "", err
            }
            defer resp.Body.Close()

            // Process response and return data
            return "API response data", nil
        }).
        ErrorHandler(func(err error) {
            fmt.Printf("API call failed: %v\n", err)
        }).
        Build()

    for value := range source.Out() {
        fmt.Println("API data:", value)
    }
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

    // Create pipeline: SingleSource -> MapFlow -> ChannelSink
    source := sources.
        Single(func() (string, error) {
            return "hello", nil
        }).
        Build()

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
        fmt.Println("Processed:", value) // Output: "HELLO"
    }
}
```

## Builder Methods

### `Single[T any](get func() (T, error)) *SingleSourceBuilder[T]`

Creates a new `SingleSourceBuilder` with the provided function.

### `Context(ctx context.Context) *SingleSourceBuilder[T]`

Sets the context for cancellation and timeout handling.

### `ErrorHandler(handler func(error)) *SingleSourceBuilder[T]`

Sets a custom error handler for function execution errors.

### `Build() *SingleSource[T]`

Creates and starts the `SingleSource`.

## Important Notes

1. **Single Use**: Each `SingleSource` can only be used once. Attempting to
   connect it to multiple flows or sinks will panic.

2. **Error Handling**: If the provided function returns an error, the source
   will not emit any values and will close immediately.

3. **Context Cancellation**: The source respects context cancellation and will
   stop execution if the context is cancelled.

4. **Channel Management**: The output channel is automatically closed when the
   source completes or encounters an error.

5. **Goroutine Safety**: The source uses atomic operations to ensure it can
   only be activated once.
