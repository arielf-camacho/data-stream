package primitives

// Out represents an any object that can send data through a channel of type T.
type Out[T any] interface {
	// Out returns a channel from which a stream of values can be read. This
	// function will always return the same non-nil channel. Don't need to close
	// the channel as it's automatically closed when the source has emitted all
	// values.
	Out() <-chan T
}
