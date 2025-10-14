package flows

import (
	"context"
	"sync/atomic"

	"github.com/arielf-camacho/data-stream/primitives"
)

var _ = primitives.Flow[int, any](&PassThroughFlow[int, any]{})

// PassThroughFlow is an operator that passes through the values from the
// input channel to the output channel.
type PassThroughFlow[IN any, OUT any] struct {
	ctx          context.Context
	activated    atomic.Bool
	bufferSize   uint
	convert      func(IN) (OUT, error)
	errorHandler func(error)

	in  chan IN
	out chan OUT
}

// PassThroughBuilder is a fluent builder for PassThroughFlow.
type PassThroughBuilder[IN any, OUT any] struct {
	ctx          context.Context
	bufferSize   uint
	errorHandler func(error)
	convert      func(IN) (OUT, error)
}

// PassThrough creates a new PassThroughBuilder for building a
// PassThroughFlow.
func PassThrough[IN any, OUT any]() *PassThroughBuilder[IN, OUT] {
	return &PassThroughBuilder[IN, OUT]{
		ctx:     context.Background(),
		convert: defaultConvert[IN, OUT],
	}
}

// Context sets the context for the PassThroughFlow.
func (b *PassThroughBuilder[IN, OUT]) Context(
	ctx context.Context,
) *PassThroughBuilder[IN, OUT] {
	b.ctx = ctx
	return b
}

// BufferSize sets the buffer size for the PassThroughFlow channels.
func (b *PassThroughBuilder[IN, OUT]) BufferSize(
	size uint,
) *PassThroughBuilder[IN, OUT] {
	b.bufferSize = size
	return b
}

// Convert sets the convert function for the PassThroughFlow.
func (b *PassThroughBuilder[IN, OUT]) Convert(
	convert func(IN) (OUT, error),
) *PassThroughBuilder[IN, OUT] {
	b.convert = convert
	return b
}

// ErrorHandler sets the error handler for the PassThroughFlow.
func (b *PassThroughBuilder[IN, OUT]) ErrorHandler(
	handler func(error),
) *PassThroughBuilder[IN, OUT] {
	b.errorHandler = handler
	return b
}

// Build creates and starts the PassThroughFlow.
func (b *PassThroughBuilder[IN, OUT]) Build() *PassThroughFlow[IN, OUT] {
	operator := &PassThroughFlow[IN, OUT]{
		ctx:          b.ctx,
		bufferSize:   b.bufferSize,
		convert:      b.convert,
		errorHandler: b.errorHandler,
	}

	operator.in = make(chan IN, operator.bufferSize)
	operator.out = make(chan OUT, operator.bufferSize)

	go operator.start()

	return operator
}

// Out returns the channel from which the values can be read.
func (p *PassThroughFlow[IN, OUT]) Out() <-chan OUT {
	return p.out
}

// In returns the channel to which the values can be written.
func (o *PassThroughFlow[IN, OUT]) In() chan<- IN {
	return o.in
}

// ToFlow passes the values from the PassThroughFlow to the given flow. This
// is exclusive with ToSink, either one must be called.
func (p *PassThroughFlow[IN, OUT]) ToFlow(
	in primitives.Flow[OUT, OUT],
) primitives.Flow[OUT, OUT] {
	p.assertNotActive()

	go func() {
		defer close(in.In())
		for {
			select {
			case <-p.ctx.Done():
				return
			case v, ok := <-p.out:
				if !ok {
					return
				}
				select {
				case <-p.ctx.Done():
					return
				case in.In() <- v:
				}
			}
		}
	}()

	return in
}

// ToSink passes the values from the PassThroughFlow to the given sink. This
// is exclusive with ToFlow, either one must be called.
func (p *PassThroughFlow[IN, OUT]) ToSink(
	in primitives.Sink[OUT],
) {
	p.assertNotActive()

	go func() {
		defer close(in.In())
		for {
			select {
			case <-p.ctx.Done():
				return
			case v, ok := <-p.out:
				if !ok {
					return
				}
				select {
				case <-p.ctx.Done():
					return
				case in.In() <- v:
				}
			}
		}
	}()
}

func (p *PassThroughFlow[IN, OUT]) start() {
	defer close(p.out)

	for {
		select {
		case <-p.ctx.Done():
			return
		case v, ok := <-p.in:
			if !ok {
				return
			}
			w, err := p.convert(v)
			if err != nil {
				if p.errorHandler != nil {
					p.errorHandler(err)
				}
				return
			}
			select {
			case <-p.ctx.Done():
				return
			case p.out <- w:
			}
		}
	}
}

func (p *PassThroughFlow[IN, OUT]) assertNotActive() {
	if !p.activated.CompareAndSwap(false, true) {
		// TODO: Use a logger to print this error, don't panic
		panic("PassThroughFlow is already streaming, cannot be used as a flow again")
	}
}

func defaultConvert[IN any, OUT any](v IN) (OUT, error) {
	return any(v).(OUT), nil
}
