package primitives

// Outlet represents any object that can send data through a channel of type T.
type Outlet[T any] interface {
	// Out returns a channel from which a stream of values can be read. This
	// function will always return the same non-nil channel. The channel is closed
	// when the outlet has emitted all values.
	Out() <-chan T
}
