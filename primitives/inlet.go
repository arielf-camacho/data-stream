package primitives

// Inlet represents any object that can receive data through a channel of type
// T.
type Inlet[T any] interface {
	// In returns a channel from which a stream of values can be read. This
	// function will always return the same non-nil channel. Don't need to close
	// the channel as it's automatically closed when the source has emitted all
	// values.
	In() chan<- T
}
