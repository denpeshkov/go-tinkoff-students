package storage

import (
	"context"
	"runtime"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

// Result represents the Size function result
type Result struct {
	// Total Size of File objects
	Size int64
	// Count is a count of File objects processed
	Count int64
}

type DirSizer interface {
	// Size calculate a size of given Dir, receive a ctx and the root Dir instance
	// will return Result or error if happened
	Size(ctx context.Context, d Dir) (Result, error)
}

// sizer implement the DirSizer interface
type sizer struct {
	// maxWorkersCount number of workers for asynchronous run
	maxWorkersCount int
}

// NewSizer returns new DirSizer instance
func NewSizer() DirSizer {
	return &sizer{
		maxWorkersCount: runtime.NumCPU(),
	}
}

func (s *sizer) Size(ctx context.Context, d Dir) (Result, error) {
	dirs, files, err := d.Ls(ctx)
	if err != nil {
		return Result{}, err
	}

	var sz, cnt atomic.Int64

	r, err := filesSize(ctx, files)
	if err != nil {
		return Result{}, err
	}
	sz.Add(r.Size)
	cnt.Add(r.Count)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(s.maxWorkersCount)
	for _, d := range dirs {
		g.Go(func() error {
			r, err := s.Size(ctx, d)
			if err != nil {
				return err
			}
			sz.Add(r.Size)
			cnt.Add(r.Count)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return Result{}, err
	}

	return Result{Size: sz.Load(), Count: cnt.Load()}, nil
}

func filesSize(ctx context.Context, files []File) (*Result, error) {
	res := new(Result)
	for _, f := range files {
		// preempt
		select {
		case <-ctx.Done():
			return res, ctx.Err()
		default:
		}

		sz, err := f.Stat(ctx)
		if err != nil {
			return &Result{}, err
		}
		res.Count++
		res.Size += sz
	}
	return res, nil
}
