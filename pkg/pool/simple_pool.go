package pool

import "context"

type SimplePool struct {
	workers chan struct{}
}

func NewSimplePool(size int) *SimplePool {
	return &SimplePool{
		workers: make(chan struct{}, size),
	}
}

func (s *SimplePool) Do(ctx context.Context, w Worker) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case s.workers <- struct{}{}:
		defer func() {
			<-s.workers
		}()
		return w()
	}
}
