# WriterSink

The `WriterSink` is a data sink that writes incoming byte data to an
`io.Writer`. It's perfect for writing stream data to files, network
connections, or any other output destination that implements the `io.Writer`
interface.

## Design

The `WriterSink` implements the `primitives.Sink[[]byte]` interface and
provides a fluent builder API for configuration. It forwards all received
byte slices to the specified writer, making it ideal for file I/O, network
operations, and other output scenarios.

### Key Features

- **io.Writer Integration**: Works with any type that implements `io.Writer`
- **Error Handling**: Supports custom error handlers for write failures
- **Buffer Control**: Configurable buffer size for the input channel
- **Context Support**: Respects context cancellation
- **Byte Stream Processing**: Specifically designed for byte data

## Graphical Representation

```
Input Stream: -- [bytes1] -- [bytes2] -- [bytes3] -- | -->

    WriterSink
         |
         v
    io.Writer: -- bytes1 -- bytes2 -- bytes3 -- | -->
```

## Usage Examples

### Writing to a File

```go
package main

import (
    "os"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create or open a file for writing
    file, err := os.Create("output.txt")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Create a writer sink
    sink := sinks.Writer(file).Build()

    // Create a source with byte data
    data := [][]byte{
        []byte("Hello, "),
        []byte("World!\n"),
        []byte("This is a "),
        []byte("data stream.\n"),
    }

    source := sources.Slice(data).Build()
    source.ToSink(sink)

    // The data is now written to output.txt
}
```

### Writing to Standard Output

```go
package main

import (
    "os"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a writer sink that writes to stdout
    sink := sinks.Writer(os.Stdout).Build()

    // Create a source with byte data
    messages := [][]byte{
        []byte("Line 1\n"),
        []byte("Line 2\n"),
        []byte("Line 3\n"),
    }

    source := sources.Slice(messages).Build()
    source.ToSink(sink)

    // Output:
    // Line 1
    // Line 2
    // Line 3
}
```

### With Error Handling

```go
package main

import (
    "fmt"
    "os"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Create a file that might not be writable
    file, err := os.Create("/readonly/output.txt")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    sink := sinks.
        Writer(file).
        ErrorHandler(func(err error) {
            fmt.Printf("Write error: %v\n", err)
        }).
        Build()

    data := [][]byte{[]byte("This might fail")}
    source := sources.Slice(data).Build()
    source.ToSink(sink)
}
```

### With Custom Buffer Size

```go
package main

import (
    "os"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    file, err := os.Create("large_output.txt")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    sink := sinks.
        Writer(file).
        BufferSize(1000). // Larger buffer for better performance
        Build()

    // Create large amounts of data
    largeData := make([][]byte, 1000)
    for i := range largeData {
        largeData[i] = []byte(fmt.Sprintf("Line %d\n", i))
    }

    source := sources.Slice(largeData).Build()
    source.ToSink(sink)
}
```

### Writing to Network Connection

```go
package main

import (
    "fmt"
    "net"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    // Connect to a server
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    sink := sinks.
        Writer(conn).
        ErrorHandler(func(err error) {
            fmt.Printf("Network write error: %v\n", err)
        }).
        Build()

    // Send data over the network
    messages := [][]byte{
        []byte("GET / HTTP/1.1\r\n"),
        []byte("Host: localhost\r\n"),
        []byte("\r\n"),
    }

    source := sources.Slice(messages).Build()
    source.ToSink(sink)
}
```

### Processing Text Data

```go
package main

import (
    "fmt"
    "os"
    "strings"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    file, err := os.Create("processed.txt")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    // Create a pipeline that processes text
    textLines := []string{
        "hello world",
        "golang streaming",
        "data processing",
    }

    source := sources.Slice(textLines).Build()

    // Convert strings to bytes and add newlines
    mapper := flows.
        Map(func(s string) ([]byte, error) {
            return []byte(strings.ToUpper(s) + "\n"), nil
        }).
        Build()

    sink := sinks.Writer(file).Build()

    // Connect the pipeline
    source.ToFlow(mapper).ToSink(sink)

    // The processed text is now in processed.txt
}
```

### Writing JSON Data

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "github.com/arielf-camacho/data-stream/flows"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    file, err := os.Create("people.json")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    people := []Person{
        {Name: "Alice", Age: 30},
        {Name: "Bob", Age: 25},
        {Name: "Charlie", Age: 35},
    }

    source := sources.Slice(people).Build()

    // Convert to JSON bytes
    jsonMapper := flows.
        Map(func(p Person) ([]byte, error) {
            jsonData, err := json.Marshal(p)
            if err != nil {
                return nil, err
            }
            return append(jsonData, '\n'), nil
        }).
        ErrorHandler(func(err error) {
            fmt.Printf("JSON marshal error: %v\n", err)
        }).
        Build()

    sink := sinks.Writer(file).Build()

    source.ToFlow(jsonMapper).ToSink(sink)
}
```

### With Context Cancellation

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    "github.com/arielf-camacho/data-stream/sinks"
    "github.com/arielf-camacho/data-stream/sources"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    file, err := os.Create("timeout_test.txt")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    sink := sinks.
        Writer(file).
        Context(ctx).
        ErrorHandler(func(err error) {
            fmt.Printf("Context error: %v\n", err)
        }).
        Build()

    // Create a large amount of data
    largeData := make([][]byte, 10000)
    for i := range largeData {
        largeData[i] = []byte(fmt.Sprintf("Line %d\n", i))
    }

    source := sources.Slice(largeData).Build()
    source.ToSink(sink)

    // The sink will stop writing when the context times out
}
```

## Builder Methods

### `Writer(w io.Writer) *WriterSinkBuilder`

Creates a new `WriterSinkBuilder` with the provided `io.Writer`.

### `Context(ctx context.Context) *WriterSinkBuilder`

Sets the context for cancellation and timeout handling.

### `BufferSize(size uint) *WriterSinkBuilder`

Sets the buffer size for the input channel. Larger buffers can improve
performance for high-throughput scenarios.

### `ErrorHandler(handler func(error)) *WriterSinkBuilder`

Sets a custom error handler for write failures.

### `Build() *WriterSink`

Creates and starts the `WriterSink`.

## Performance Considerations

1. **Buffer Size**: For high-throughput scenarios, use a larger buffer size to
   reduce blocking between the sink and upstream flows.

2. **Writer Performance**: The performance of the `WriterSink` depends on the
   underlying writer. File writers, network writers, and in-memory writers
   have different performance characteristics.

3. **Error Handling**: Always provide an error handler for production code to
   handle write failures gracefully.

## Important Notes

1. **Byte Data Only**: The `WriterSink` only accepts `[]byte` data. Use a
   `MapFlow` to convert other types to bytes.

2. **Writer Responsibility**: The sink does not close the writer. The caller
   is responsible for closing it when appropriate.

3. **Error Handling**: Write errors will cause the sink to stop processing.
   Always provide an error handler for robust applications.

4. **Context Cancellation**: When the context is cancelled, the sink stops
   writing and the input channel is closed.

5. **Blocking Behavior**: If the underlying writer blocks (e.g., network
   congestion), the sink will block until the write completes.
