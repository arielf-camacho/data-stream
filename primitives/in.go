package primitives

// In represents any object that can receive data through a channel of type
// T.
type In[T any] interface {
	In() chan<- T
}
