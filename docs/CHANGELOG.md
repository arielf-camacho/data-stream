# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2025-01-09

### Added

- SingleSink sink node to receive of a single value from upstream and wait for it

## [1.1.0] - 2024-12-15

### Added

- **Waitable behavior to sinks**: All sinks now support waiting for completion
  - `ChannelSink` can be waited on to ensure all values are processed
  - `WriterSink` can be waited on to ensure all writes are completed
  - `ReduceSink` can be waited on to get the final reduced value
- Enhanced sink interface with completion tracking
- Improved error handling in sink operations
- Comprehensive documentation for all components
- CHANGELOG.md for tracking project changes

### Changed

- Sink builders now return waitable sink implementations
- Sink completion is now tracked automatically

## [1.0.0] - 2025-10-13

### Added

- **Initial core components**:
  - **Sources**:
    - `SingleSource`: Emits a single value from a function
    - `SliceSource`: Emits all values from a slice
  - **Flows**:
    - `FilterFlow`: Filters values based on a predicate function
    - `MapFlow`: Transforms values using a function with optional parallelism
    - `MergeFlow`: Merges multiple streams into a single output stream
    - `SplitFlow`: Splits a stream into two based on a predicate
    - `SpreadFlow`: Duplicates a stream to multiple outputs
    - `PassThroughFlow`: Passes values through with optional type conversion
  - **Sinks**:
    - `ChannelSink`: Writes values to a Go channel
    - `WriterSink`: Writes byte data to an io.Writer
- **Core primitives**:
  - `Source[T]`: Interface for data sources
  - `Flow[IN, OUT]`: Interface for data transformations
  - `Sink[T]`: Interface for data consumers
  - `Inlet[T]` and `Outlet[T]`: Base interfaces for channel communication
- **Builder pattern**: Fluent API for configuring all components
- **Context support**: All components respect Go's context for cancellation
- **Error handling**: Customizable error handlers for robust applications
- **Type safety**: Full generic support with compile-time type checking
- **Concurrency**: Built-in support for parallel processing where appropriate
- **Resource management**: Automatic channel lifecycle management
- **Examples**: Complete working examples for all components
- **Documentation**: Comprehensive documentation with usage examples
- **Testing**: Full test coverage for all components
- **Makefile**: Development tools and commands for contributors

### Technical Details

- Built with Go 1.24.4
- Uses generics for type safety
- Implements channel-based communication
- Supports context cancellation throughout the pipeline
- Provides fluent builder APIs for easy configuration
- Includes comprehensive error handling mechanisms
