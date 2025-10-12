package primitives

// Source is an interface that represents a source of data. It yields a channel
// from which a stream of values can be read and passed downstream for further
// processing.
type Source[T any] interface {
	Out[T]

	// To passes the values from the source to the given In object. The channel
	// provided by the In object, even though it's owned by it, will be closed
	// when the source has emitted all values automatically. Writers of In objects
	// should not close the channel manually.
	To(in In[T])
}
