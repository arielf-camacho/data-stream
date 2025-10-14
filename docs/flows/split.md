# SplitFlow

The `SplitFlow` is a data routing operator that splits a single input stream
into two output streams based on a predicate function. Values that match the
predicate go to one stream, while non-matching values go to another stream.

## Design

The `SplitFlow` provides a fluent builder API for configuration. It evaluates
each incoming value against a predicate function and routes it to either the
matching or non-matching output stream.

### Key Features

- **Conditional Routing**: Routes values based on a predicate function
- **Two Output Streams**: Provides separate streams for matching and
  non-matching values
- **Error Handling**: Supports custom error handlers for predicate failures
- **Context Support**: Respects context cancellation
- **Type Preservation**: Maintains the same type for all streams

## Graphical Representation

```
Input Stream: -- 1 -- 2 -- 3 -- 4 -- 5 -- | -->

    SplitFlow (predicate: x % 2 == 0)
         |
         v
    Matching Stream: ------- 2 ------- 4 -- | -->
    NonMatching Stream: -- 1 ------- 3 ------- 5 -- | -->
```

## Usage Examples

### Basic Splitting

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
    // Create output channels for both streams
    evenCh := make(chan int)
    oddCh := make(chan int)

    // Create sinks
    evenSink := sinks.Channel(evenCh).Build()
    oddSink := sinks.Channel(oddCh).Build()

    // Create the split flow
    split := flows.
        Split(sources.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}).Build(),
            func(x int) (bool, error) {
                return x%2 == 0, nil
            }).
        Build()

    // Connect the split outputs to sinks
    split.Matching().ToSink(evenSink)
    split.NonMatching().ToSink(oddSink)

    var wg sync.WaitGroup
    wg.Add(2)

    // Consumer for even numbers
    go func() {
        defer wg.Done()
        for value := range evenCh {
            fmt.Printf("Even: %d\n", value)
        }
    }()

    // Consumer for odd numbers
    go func() {
        defer wg.Done()
        for value := range oddCh {
            fmt.Printf("Odd: %d\n", value)
        }
    }()

    wg.Wait()
    // Output:
    // Even: 2
    // Odd: 1
    // Even: 4
    // Odd: 3
    // Even: 6
    // Odd: 5
    // Even: 8
    // Odd: 7
    // Even: 10
    // Odd: 9
}
```

### String Splitting

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
    longWordsCh := make(chan string)
    shortWordsCh := make(chan string)

    longWordsSink := sinks.Channel(longWordsCh).Build()
    shortWordsSink := sinks.Channel(shortWordsCh).Build()

    words := []string{"apple", "cat", "elephant", "dog", "butterfly", "ant"}

    split := flows.
        Split(sources.Slice(words).Build(),
            func(s string) (bool, error) {
                return len(s) > 4, nil
            }).
        Build()

    split.Matching().ToSink(longWordsSink)
    split.NonMatching().ToSink(shortWordsSink)

    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        for word := range longWordsCh {
            fmt.Printf("Long word: %s\n", word)
        }
    }()

    go func() {
        defer wg.Done()
        for word := range shortWordsCh {
            fmt.Printf("Short word: %s\n", word)
        }
    }()

    wg.Wait()
    // Output:
    // Long word: apple
    // Short word: cat
    // Long word: elephant
    // Short word: dog
    // Long word: butterfly
    // Short word: ant
}
```

### Complex Data Splitting

```go
package main

import (
    "fmt"
    "sync"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

type Product struct {
    Name  string
    Price float64
    Category string
}

func main() {
    expensiveCh := make(chan Product)
    cheapCh := make(chan Product)

    expensiveSink := sinks.Channel(expensiveCh).Build()
    cheapSink := sinks.Channel(cheapCh).Build()

    products := []Product{
        {Name: "Laptop", Price: 999.99, Category: "Electronics"},
        {Name: "Book", Price: 19.99, Category: "Education"},
        {Name: "Phone", Price: 699.99, Category: "Electronics"},
        {Name: "Pen", Price: 2.99, Category: "Office"},
        {Name: "Tablet", Price: 499.99, Category: "Electronics"},
    }

    split := flows.
        Split(sources.Slice(products).Build(),
            func(p Product) (bool, error) {
                return p.Price > 100.0, nil
            }).
        Build()

    split.Matching().ToSink(expensiveSink)
    split.NonMatching().ToSink(cheapSink)

    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        for product := range expensiveCh {
            fmt.Printf("Expensive: %s ($%.2f)\n", product.Name, product.Price)
        }
    }()

    go func() {
        defer wg.Done()
        for product := range cheapCh {
            fmt.Printf("Cheap: %s ($%.2f)\n", product.Name, product.Price)
        }
    }()

    wg.Wait()
    // Output:
    // Expensive: Laptop ($999.99)
    // Cheap: Book ($19.99)
    // Expensive: Phone ($699.99)
    // Cheap: Pen ($2.99)
    // Expensive: Tablet ($499.99)
}
```

### With Error Handling

```go
package main

import (
    "errors"
    "fmt"
    "sync"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    validCh := make(chan int)
    invalidCh := make(chan int)

    validSink := sinks.Channel(validCh).Build()
    invalidSink := sinks.Channel(invalidCh).Build()

    numbers := []int{1, 2, -3, 4, 5, -6, 7}

    split := flows.
        Split(sources.Slice(numbers).Build(),
            func(x int) (bool, error) {
                if x < 0 {
                    return false, errors.New("negative number found")
                }
                return x%2 == 0, nil
            }).
        ErrorHandler(func(err error) {
            fmt.Printf("Split error: %v\n", err)
        }).
        Build()

    split.Matching().ToSink(validSink)
    split.NonMatching().ToSink(invalidSink)

    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        for value := range validCh {
            fmt.Printf("Valid even: %d\n", value)
        }
    }()

    go func() {
        defer wg.Done()
        for value := range invalidCh {
            fmt.Printf("Invalid/odd: %d\n", value)
        }
    }()

    wg.Wait()
    // Output:
    // Valid even: 2
    // Invalid/odd: 1
    // Split error: negative number found
    // (pipeline stops due to error)
}
```

### With Context Cancellation

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
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    highCh := make(chan int)
    lowCh := make(chan int)

    highSink := sinks.Channel(highCh).Build()
    lowSink := sinks.Channel(lowCh).Build()

    // Create a large dataset
    largeData := make([]int, 1000)
    for i := range largeData {
        largeData[i] = i
    }

    split := flows.
        Split(sources.Slice(largeData).Build(),
            func(x int) (bool, error) {
                // Simulate some processing time
                time.Sleep(10 * time.Millisecond)
                return x > 500, nil
            }).
        Context(ctx).
        Build()

    split.Matching().ToSink(highSink)
    split.NonMatching().ToSink(lowSink)

    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        count := 0
        for value := range highCh {
            count++
            if count%50 == 0 {
                fmt.Printf("High values processed: %d\n", count)
            }
        }
        fmt.Printf("Total high values: %d\n", count)
    }()

    go func() {
        defer wg.Done()
        count := 0
        for value := range lowCh {
            count++
            if count%50 == 0 {
                fmt.Printf("Low values processed: %d\n", count)
            }
        }
        fmt.Printf("Total low values: %d\n", count)
    }()

    wg.Wait()
}
```

### Chaining with Other Flows

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
    uppercaseCh := make(chan string)
    lowercaseCh := make(chan string)

    uppercaseSink := sinks.Channel(uppercaseCh).Build()
    lowercaseSink := sinks.Channel(lowercaseCh).Build()

    words := []string{"Apple", "banana", "Cherry", "date", "Elderberry"}

    split := flows.
        Split(sources.Slice(words).Build(),
            func(s string) (bool, error) {
                return strings.ToUpper(s) == s, nil
            }).
        Build()

    // Process uppercase words (add prefix)
    uppercaseProcessor := flows.
        Map(func(s string) (string, error) {
            return "UPPER: " + s, nil
        }).
        Build()

    // Process lowercase words (convert to uppercase)
    lowercaseProcessor := flows.
        Map(func(s string) (string, error) {
            return "LOWER: " + strings.ToUpper(s), nil
        }).
        Build()

    // Connect the pipeline
    split.Matching().ToFlow(uppercaseProcessor).ToSink(uppercaseSink)
    split.NonMatching().ToFlow(lowercaseProcessor).ToSink(lowercaseSink)

    var wg sync.WaitGroup
    wg.Add(2)

    go func() {
        defer wg.Done()
        for value := range uppercaseCh {
            fmt.Println("Uppercase result:", value)
        }
    }()

    go func() {
        defer wg.Done()
        for value := range lowercaseCh {
            fmt.Println("Lowercase result:", value)
        }
    }()

    wg.Wait()
    // Output:
    // Uppercase result: UPPER: Apple
    // Lowercase result: LOWER: BANANA
    // Uppercase result: UPPER: Cherry
    // Lowercase result: LOWER: DATE
    // Uppercase result: UPPER: Elderberry
}
```

## Builder Methods

### `Split[T any](in primitives.Outlet[T], predicate func(T) (bool, error)) *SplitBuilder[T]`

Creates a new `SplitBuilder` with the provided input stream and predicate.

### `Context(ctx context.Context) *SplitBuilder[T]`

Sets the context for cancellation and timeout handling.

### `BufferSize(size uint) *SplitBuilder[T]`

Sets the buffer size for the internal flows.

### `ErrorHandler(handler func(error)) *SplitBuilder[T]`

Sets a custom error handler for predicate execution errors.

### `Build() *SplitFlow[T]`

Creates and starts the `SplitFlow`.

## Access Methods

### `Matching() primitives.Flow[T, T]`

Returns the flow for values that match the predicate.

### `NonMatching() primitives.Flow[T, T]`

Returns the flow for values that don't match the predicate.

## Performance Considerations

1. **Predicate Complexity**: Keep predicate functions simple and fast. Complex
   operations can slow down the entire pipeline.

2. **Buffer Size**: Larger buffer sizes can improve performance by reducing
   blocking between flows.

3. **Consumer Speed**: If one output stream is consumed much slower than the
   other, it can block the entire split operation.

## Important Notes

1. **Two Output Streams**: The split always produces exactly two output streams

   - one for matching values and one for non-matching values.

2. **Error Handling**: If the predicate function returns an error, the split
   will stop processing and close both output streams.

3. **Context Cancellation**: When the context is cancelled, the split stops
   processing and closes both output streams.

4. **Type Preservation**: All streams maintain the same type `T`.

5. **Order Preservation**: The order of values in each output stream matches
   the order they were processed, but the two streams may have different
   relative timing.
