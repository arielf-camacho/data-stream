# Chain Utilities

The chain utilities provide helper functions for chaining flows and sources
when type constraints prevent direct method chaining. These utilities are
essential when you need to connect flows with different input/output types or
when the compiler cannot infer the correct types for method chaining.

## Design

The chain utilities are standalone functions in the `flows` package that
facilitate type-safe chaining of components. They handle the channel streaming
between components and return the downstream component for further chaining.

### Key Features

- **Type-Safe Chaining**: Enables chaining flows with different type
  parameters
- **Source Integration**: Allows chaining sources to flows when direct
  chaining fails
- **Context Support**: All utilities respect context for cancellation
- **Flexible Composition**: Works with any components implementing the
  primitives interfaces

## Usage Examples

### Chaining Flows with Different Types

```go
package main

import (
    "context"
    "fmt"
    "strconv"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan string)

    // First flow: int -> string
    from := flows.
        Map(func(x int) (string, error) {
            return strconv.Itoa(x), nil
        }).
        Build()

    // Second flow: string -> string (adds prefix)
    to := flows.
        Map(func(x string) (string, error) {
            return "num:" + x, nil
        }).
        Build()

    // Chain them using ToFlow utility
    // from.ToFlow(to) would not work due to type constraints
    flows.ToFlow(context.Background(), from, to)

    // Feed input to first flow
    go func() {
        for i := 1; i <= 5; i++ {
            from.In() <- i
        }
        close(from.In())
    }()

    sink := sinks.Channel(outputCh).Build()
    to.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Result:", value)
    }
    // Output:
    // Result: num:1
    // Result: num:2
    // Result: num:3
    // Result: num:4
    // Result: num:5
}
```

### Chaining Source to Flow

```go
package main

import (
    "context"
    "fmt"
    "strconv"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan string)

    // Source: int slice
    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()

    // Flow: int -> string
    flow := flows.
        Map(func(x int) (string, error) {
            return "num:" + strconv.Itoa(x), nil
        }).
        Build()

    // Chain source to flow using SourceToFlow utility
    // source.ToFlow(flow) might not work in all cases
    flows.SourceToFlow(context.Background(), source, flow)

    sink := sinks.Channel(outputCh).Build()
    flow.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Processed:", value)
    }
    // Output:
    // Processed: num:1
    // Processed: num:2
    // Processed: num:3
    // Processed: num:4
    // Processed: num:5
}
```

### Complex Multi-Type Chain

```go
package main

import (
    "context"
    "fmt"
    "strconv"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan string)

    // Source: int slice
    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()

    // Flow 1: int -> int (doubles)
    doubleFlow := flows.
        Map(func(x int) (int, error) {
            return x * 2, nil
        }).
        Build()

    // Flow 2: int -> string
    stringFlow := flows.
        Map(func(x int) (string, error) {
            return strconv.Itoa(x), nil
        }).
        Build()

    // Flow 3: string -> string (adds prefix)
    prefixFlow := flows.
        Map(func(x string) (string, error) {
            return "doubled:" + x, nil
        }).
        Build()

    // Chain using utilities
    flows.SourceToFlow(context.Background(), source, doubleFlow)
    flows.ToFlow(context.Background(), doubleFlow, stringFlow)
    flows.ToFlow(context.Background(), stringFlow, prefixFlow)

    sink := sinks.Channel(outputCh).Build()
    prefixFlow.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Result:", value)
    }
    // Output:
    // Result: doubled:2
    // Result: doubled:4
    // Result: doubled:6
    // Result: doubled:8
    // Result: doubled:10
}
```

### Chaining with Filter in Between

```go
package main

import (
    "context"
    "fmt"
    "strconv"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan string)

    // Flow 1: int -> int (doubles)
    from := flows.
        Map(func(x int) (int, error) {
            return x * 2, nil
        }).
        Build()

    // Flow 2: int -> int (filter even numbers)
    filter := flows.
        Filter(func(x int) (bool, error) {
            return x%2 == 0, nil
        }).
        Build()

    // Flow 3: int -> string
    to := flows.
        Map(func(x int) (string, error) {
            return strconv.Itoa(x), nil
        }).
        Build()

    // Chain using utilities
    flows.ToFlow(context.Background(), from, filter)
    flows.ToFlow(context.Background(), filter, to)

    // Feed input to first flow
    go func() {
        for i := 1; i <= 5; i++ {
            from.In() <- i
        }
        close(from.In())
    }()

    sink := sinks.Channel(outputCh).Build()
    to.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Filtered:", value)
    }
    // Output:
    // Filtered: 2
    // Filtered: 4
    // Filtered: 6
    // Filtered: 8
    // Filtered: 10
}
```

### Chaining Channel Source

```go
package main

import (
    "context"
    "fmt"
    "strconv"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan string)

    // Create channel source
    ch := make(chan int, 5)
    go func() {
        defer close(ch)
        for i := 1; i <= 5; i++ {
            ch <- i
        }
    }()
    source := sources.Channel(ch).Build()

    // Flow: int -> int (doubles)
    mapFlow := flows.
        Map(func(x int) (int, error) {
            return x * 2, nil
        }).
        Build()

    // Flow: int -> string
    stringFlow := flows.
        Map(func(x int) (string, error) {
            return strconv.Itoa(x), nil
        }).
        Build()

    // Chain using utilities
    flows.SourceToFlow(context.Background(), source, mapFlow)
    flows.ToFlow(context.Background(), mapFlow, stringFlow)

    sink := sinks.Channel(outputCh).Build()
    stringFlow.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Result:", value)
    }
    // Output:
    // Result: 2
    // Result: 4
    // Result: 6
    // Result: 8
    // Result: 10
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

    source := sources.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).Build()

    // Flow with delay
    slowFlow := flows.
        Map(func(x int) (int, error) {
            time.Sleep(100 * time.Millisecond)
            return x * 2, nil
        }).
        Build()

    // Chain with context
    flows.SourceToFlow(ctx, source, slowFlow)

    sink := sinks.Channel(outputCh).Build()
    slowFlow.ToSink(sink)

    count := 0
    for value := range outputCh {
        count++
        fmt.Printf("Processed %d values, current: %d\n", count, value)
    }

    fmt.Printf("Total processed before timeout: %d\n", count)
}
```

### Chaining Return Values

```go
package main

import (
    "context"
    "fmt"
    "strconv"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    outputCh := make(chan string)

    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()

    // Chain and use return value for further chaining
    flow1 := flows.
        Map(func(x int) (string, error) {
            return strconv.Itoa(x), nil
        }).
        Build()

    flow2 := flows.
        Map(func(x string) (string, error) {
            return "value:" + x, nil
        }).
        Build()

    // ToFlow returns the downstream flow, allowing method chaining
    flows.SourceToFlow(context.Background(), source, flow1).
        ToFlow(flow2)

    sink := sinks.Channel(outputCh).Build()
    flow2.ToSink(sink)

    for value := range outputCh {
        fmt.Println("Result:", value)
    }
    // Output:
    // Result: value:1
    // Result: value:2
    // Result: value:3
    // Result: value:4
    // Result: value:5
}
```

## Function Reference

### `ToFlow[IN, OUT, NEXT any](ctx context.Context, from primitives.Flow[IN, OUT], to primitives.Flow[OUT, NEXT]) primitives.Flow[OUT, NEXT]`

Connects two flows where the output type of the first flow matches the input
type of the second flow. This is useful when direct method chaining fails due
to type constraints.

**Parameters:**

- `ctx`: Context for cancellation and timeout handling
- `from`: The upstream flow (output type OUT)
- `to`: The downstream flow (input type OUT)

**Returns:** The downstream flow for further chaining

**Example:**

```go
from := flows.Map(func(x int) (string, error) { return strconv.Itoa(x), nil }).Build()
to := flows.Map(func(x string) (string, error) { return "num:" + x, nil }).Build()
flows.ToFlow(context.Background(), from, to)
```

### `SourceToFlow[IN, OUT any](ctx context.Context, source primitives.Source[IN], flow primitives.Flow[IN, OUT]) primitives.Flow[IN, OUT]`

Connects a source to a flow where the source output type matches the flow input
type. This is useful when direct method chaining fails due to type constraints.

**Parameters:**

- `ctx`: Context for cancellation and timeout handling
- `source`: The source (output type IN)
- `flow`: The flow (input type IN)

**Returns:** The flow for further chaining

**Example:**

```go
source := sources.Slice([]int{1, 2, 3}).Build()
flow := flows.Map(func(x int) (string, error) { return strconv.Itoa(x), nil }).Build()
flows.SourceToFlow(context.Background(), source, flow)
```

## Performance Considerations

1. **Context Usage**: Always provide a context when using these utilities to
   enable proper cancellation and timeout handling.

2. **Type Safety**: These utilities maintain full type safety at compile time,
   ensuring that incompatible types cannot be chained.

3. **Method Chaining**: When possible, prefer direct method chaining (e.g.,
   `source.ToFlow(flow)`) as it's more idiomatic. Use these utilities only when
   type constraints prevent direct chaining.

## Important Notes

1. **Type Compatibility**: The utilities enforce type compatibility at compile
   time. The output type of the upstream component must match the input type
   of the downstream component.

2. **Context Propagation**: The context is used for the streaming operation
   between components. Make sure to use appropriate contexts for cancellation
   and timeout handling.

3. **Return Values**: Both functions return the downstream component, allowing
   for further chaining if needed.

4. **Single Use**: The components being chained still follow their single-use
   semantics. Each component can only be connected once.

5. **When to Use**: Use these utilities when:
   - Direct method chaining fails due to type inference issues
   - You need explicit type parameters for clarity
   - You're chaining components with complex type relationships
