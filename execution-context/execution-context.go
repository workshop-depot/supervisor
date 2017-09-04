package executioncontext

import (
	"context"
	"sync"
)

// WaitGroup interface for built-in WaitGroup
type WaitGroup interface {
	Add(delta int)
	Done()
	Wait()
}

// ExecutionContext combination of context.Context & WaitGroup
type ExecutionContext interface {
	context.Context
	WaitGroup() WaitGroup
}

type concreteExecutionContext struct {
	context.Context
	wg *sync.WaitGroup
}

func (x *concreteExecutionContext) WaitGroup() WaitGroup { return x.wg }

// ErrNilContext means we got a nil context while we shouldn't
var ErrNilContext = errorf("ERR_NIL_CONTEXT")

// New .
func New(ctx context.Context) (ExecutionContext, error) {
	if ctx == nil {
		return nil, ErrNilContext
	}
	return &concreteExecutionContext{
		Context: ctx,
		wg:      &sync.WaitGroup{},
	}, nil
}
