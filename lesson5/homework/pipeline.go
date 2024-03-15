package executor

import (
	"context"
)

type (
	In  <-chan any
	Out = In
)

type Stage func(in In) (out Out)

func ExecutePipeline(ctx context.Context, in In, stages ...Stage) Out {
	in = func(in In) Out {
		for _, st := range stages {
			in = st(in)
		}
		return in
	}(in)

	out := make(chan any)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				// if closed
				if !ok {
					return
				}
				out <- v
			}
		}
	}()
	return out
}
