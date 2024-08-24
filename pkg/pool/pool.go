package pool

import "context"

type Pool interface {
	Do(ctx context.Context, w Worker) error
}

type Worker func() error
