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
	ctx        context.Context
	activated  atomic.Bool
	bufferSize uint

	convert func(IN) OUT
	in      chan IN
	out     chan OUT
}

// PassThroughBuilder is a fluent builder for PassThroughFlow.
type PassThroughBuilder[IN any, OUT any] struct {
	ctx        context.Context
	bufferSize uint
	convert    func(IN) OUT
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

func (b *PassThroughBuilder[IN, OUT]) Convert(
	convert func(IN) OUT,
) *PassThroughBuilder[IN, OUT] {
	b.convert = convert
	return b
}

// Build creates and starts the PassThroughFlow.
func (b *PassThroughBuilder[IN, OUT]) Build() *PassThroughFlow[IN, OUT] {
	operator := &PassThroughFlow[IN, OUT]{
		ctx:        b.ctx,
		bufferSize: b.bufferSize,
		convert:    b.convert,
	}

	operator.in = make(chan IN, operator.bufferSize)
	operator.out = make(chan OUT, operator.bufferSize)

	go operator.start()

	return operator
}

func (p *PassThroughFlow[IN, OUT]) Out() <-chan OUT {
	return p.out
}

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
		for v := range p.out {
			select {
			case <-p.ctx.Done():
				return
			case in.In() <- v:
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
		for v := range p.out {
			select {
			case <-p.ctx.Done():
				return
			case in.In() <- v:
			}
		}
	}()
}

func (p *PassThroughFlow[IN, OUT]) start() {
	defer close(p.out)

	for v := range p.in {
		w := p.convert(v)
		select {
		case <-p.ctx.Done():
			return
		case p.out <- w:
		}
	}
}

func (p *PassThroughFlow[IN, OUT]) assertNotActive() {
	if !p.activated.CompareAndSwap(false, true) {
		// TODO: Use a logger to print this error, don't panic
		panic("PassThroughFlow is already streaming, cannot be used as a flow again")
	}
}

func defaultConvert[IN any, OUT any](v IN) OUT {
	return any(v).(OUT)
}
