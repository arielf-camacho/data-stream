package primitives

// Operator represents any object that can receive data through a channel of
// type IN and send data through a channel of type OUT.
type Operator[IN any, OUT any] interface {
	Source[IN]
	Sink[OUT]
}
