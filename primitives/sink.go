package primitives

// Sink represents any object that can receive data through a channel of type T.
type Sink[T any] interface {
	Inlet[T]
}

// WaitableSink represents any object that can wait for the sink to finish
// processing all values, blocking the current goroutine.
type WaitableSink[T any] interface {
	Sink[T]

	// Wait waits for the sink to finish processing all values. An error is
	// returned if the wait is somehow impossible according to the own sink's
	// implementation.
	Wait() error
}
