package primitives

// Flow represents any object that can receive data through a channel of
// type IN and send data through a channel of type OUT.
type Flow[IN any, OUT any] interface {
	Inlet[IN]
	Outlet[OUT]

	// ToFlow passes the values from this flow to the given In object. The
	// channel provided by the In object, even though it's owned by it, will be
	// closed when the flow has emitted all values automatically. Writers of In
	// objects should not close the channel manually.
	ToFlow(in Flow[OUT, OUT]) Flow[OUT, OUT]

	// ToSink passes the values from this flow to the given Sink. The sink's input
	// channel will be closed when the flow has emitted all values automatically.
	ToSink(in Sink[OUT])
}
