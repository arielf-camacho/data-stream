package primitives

// Sink represents any object that can receive data through a channel of type T.
type Sink[T any] interface {
	In[T]
}
