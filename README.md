# Data Stream

A powerful and flexible Go library for building data streaming pipelines with a
clean, composable API. Data Stream provides a comprehensive set of primitives for
creating, transforming, and consuming data streams with built-in support for
concurrency, error handling, and context cancellation.

## Purpose

Data Stream is designed to simplify the creation of data processing pipelines in
Go applications. It provides a fluent, builder-based API that makes it easy to
compose complex data transformations while maintaining type safety and
performance.

## Design Principles

### 1. **Composability**

All components (sources, flows, sinks) can be easily chained together to create
complex data processing pipelines.

### 2. **Type Safety**

Full generic support ensures compile-time type checking throughout your
pipelines.

### 3. **Concurrency by Default**

Built-in support for parallel processing and concurrent operations where
appropriate.

### 4. **Context Awareness**

All components respect Go's context for cancellation and timeout handling.

### 5. **Error Handling**

Comprehensive error handling with customizable error handlers for robust
applications.

### 6. **Resource Management**

Automatic channel lifecycle management prevents resource leaks.

## Quick Start

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

    // Create a pipeline: source -> filter -> sink
    source := sources.Slice([]int{1, 2, 3, 4, 5}).Build()
    filter := flows.Filter(func(x int) (bool, error) {
        return x%2 == 0, nil
    }).Build()
    sink := sinks.Channel(outputCh).Build()

    // Connect the pipeline
    source.ToFlow(filter).ToSink(sink)

    // Consume results
    for v := range outputCh {
        fmt.Println("Even number:", v)
    }
}
```

## Makefile Commands

The following commands are essential for contributors:

### Development Setup

```bash
make init          # Install required tools and dependencies
make tidy          # Format code and clean dependencies
```

### Code Quality

```bash
make vet           # Run go vet for static analysis
make lint          # Run golangci-lint for comprehensive linting
```

### Testing

```bash
make test          # Run all tests
make test-no-cache # Run tests without cache
make coverage      # Run tests with coverage report
make coverage/html # Generate HTML coverage report
```

### Complete Workflow

```bash
make all           # Run tidy, vet, lint, and coverage
```

## Component Overview

### Sources

Data sources that emit values into the stream:

| Component        | Description                          | Documentation                                    |
| ---------------- | ------------------------------------ | ------------------------------------------------ |
| **SingleSource** | Emits a single value from a function | [docs/sources/single.md](docs/sources/single.md) |
| **SliceSource**  | Emits all values from a slice        | [docs/sources/slice.md](docs/sources/slice.md)   |

### Flows

Data transformation operators that process values:

| Component           | Description                             | Documentation                                            |
| ------------------- | --------------------------------------- | -------------------------------------------------------- |
| **FilterFlow**      | Filters values based on a predicate     | [docs/flows/filter.md](docs/flows/filter.md)             |
| **MapFlow**         | Transforms values using a function      | [docs/flows/map.md](docs/flows/map.md)                   |
| **MergeFlow**       | Merges multiple streams into one        | [docs/flows/merge.md](docs/flows/merge.md)               |
| **SplitFlow**       | Splits a stream based on a predicate    | [docs/flows/split.md](docs/flows/split.md)               |
| **SpreadFlow**      | Duplicates a stream to multiple outputs | [docs/flows/spread.md](docs/flows/spread.md)             |
| **PassThroughFlow** | Passes values through unchanged         | [docs/flows/pass-through.md](docs/flows/pass-through.md) |

### Sinks

Data consumers that receive values from the stream:

| Component       | Description                       | Documentation                                  |
| --------------- | --------------------------------- | ---------------------------------------------- |
| **ChannelSink** | Writes values to a Go channel     | [docs/sinks/channel.md](docs/sinks/channel.md) |
| **WriterSink**  | Writes byte data to an io.Writer  | [docs/sinks/writer.md](docs/sinks/writer.md)   |
| **ReduceSink**  | Reduces values to a single result | [docs/sinks/reduce.md](docs/sinks/reduce.md)   |

## Examples

Check out the [examples](examples/) directory for complete working examples:

- [Single Source](examples/single-source/) - Basic single value streaming
- [Filter](examples/filter/) - Filtering and merging streams
- [Merge](examples/merge/) - Merging multiple streams
- [Split](examples/split/) - Splitting streams based on conditions
- [Spread](examples/spread/) - Duplicating streams to multiple outputs
- [Console](examples/console/) - Shows how to output to the console

## Architecture

Data Stream is built around three core primitives:

- **Sources** (`primitives.Source[T]`) - Generate data
- **Flows** (`primitives.Flow[IN, OUT]`) - Transform data
- **Sinks** (`primitives.Sink[T]`) - Consume data

These primitives implement the `Inlet[T]` and `Outlet[T]` interfaces, which
provide the channel-based communication mechanism that powers the streaming
pipeline.

## Contributing

1. Run `make init` to set up your development environment
2. Make your changes
3. Run `make all` to ensure code quality
4. Submit a pull request

## License

This project is licensed under the MIT License.
