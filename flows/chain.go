package flows

import (
	"context"

	"github.com/arielf-camacho/data-stream/helpers"
	"github.com/arielf-camacho/data-stream/primitives"
)

// ToFlow passes the values from the From flow to the To flow. This is a utility
// function that can be used to chain which types differ. In you have a flow
// compatible with Flow[int, string], and you want to chain it with a flow
// compatible with Flow[string, float64], you can use this function to chain
// them together, as the own ToFlow() method of the Flow[int, string] will not
// work because of compiler type constraints.
//
//		Example of usage:
//
//		from := flows.Map(func(x int) (string, error) { return strconv.Itoa(x), nil }).Build()
//		to := flows.Map(func(x string) (float64, error) { return strconv.ParseFloat(x, 64), nil }).Build()
//		flows.ToFlow(context.Background(), from, to)
//
//	  from.ToFlow(to) // will not work because of compiler type constraints
func ToFlow[IN, OUT, NEXT any](
	ctx context.Context,
	from primitives.Flow[IN, OUT],
	to primitives.Flow[OUT, NEXT],
) primitives.Flow[OUT, NEXT] {
	helpers.StreamTo(ctx, from.Out(), to.In())
	return to
}

// SourceToFlow passes the values from the source to the flow. This is a utility
// function that can be used to chain which types differ. In you have a source
// compatible with Source[int], and you want to chain it with a flow
// compatible with Flow[int, string], you can use this function to chain
// them together, as the own ToFlow() method of the Source[int] will not
// work because of compiler type constraints.
//
//		Example of usage:
//
//		source := sources.Single(func() (int, error) { return 1, nil }).Build()
//		flow := flows.Map(func(x int) (string, error) { return strconv.Itoa(x), nil }).Build()
//		flows.SourceToFlow(context.Background(), source, flow)
//
//	  source.ToFlow(flow) // will not work because of compiler type constraints
func SourceToFlow[IN, OUT any](
	ctx context.Context,
	source primitives.Source[IN],
	flow primitives.Flow[IN, OUT],
) primitives.Flow[IN, OUT] {
	helpers.StreamTo(ctx, source.Out(), flow.In())
	return flow
}
