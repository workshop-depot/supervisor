package executioncontext

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecutionContext1(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wctx, _ := New(ctx)

	var sum int64

	for i := 0; i < 10; i++ {
		i := i + 1
		wctx.WaitGroup().Add(1)
		go func() {
			defer wctx.WaitGroup().Done()
			<-wctx.Done()
			atomic.AddInt64(&sum, int64(i))
		}()
	}

	cancel()
	wctx.WaitGroup().Wait()
	assert.Equal(t, int64(55), sum)
}

func TestExecutionContext2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wctx, _ := New(ctx)

	var sum int64

	for i := 0; i < 10; i++ {
		i := i + 1
		wctx.WaitGroup().Add(1)
		go func() {
			<-wctx.Done()
			atomic.AddInt64(&sum, int64(i))
		}()
	}

	cancel()
	waitDone := make(chan struct{})
	go func() {
		defer close(waitDone)
		wctx.WaitGroup().Wait()
	}()
	select {
	case <-waitDone:
	case <-time.After(time.Millisecond * 100):
		atomic.AddInt64(&sum, 11)
	}
	assert.Equal(t, int64(66), sum)
}

func TestExecutionContext3(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wctx, _ := New(ctx)

	var sum int64

	for i := 0; i < 10; i++ {
		i := i + 1
		wctx.WaitGroup().Add(1)
		go func() {
			defer wctx.WaitGroup().Done()
			if i == 3 {
				return
			}
			<-wctx.Done()
			atomic.AddInt64(&sum, int64(i))
		}()
	}

	cancel()
	wctx.WaitGroup().Wait()
	assert.Equal(t, int64(52), sum)
}
