package usagetests

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dc0d/supervisor"
	"github.com/stretchr/testify/assert"
)

var (
	rootCtx   context.Context
	rootGroup = &sync.WaitGroup{}
)

func TestMain(m *testing.M) {
	rootGroupStopped := make(chan struct{})
	go func() {
		rootGroup.Wait()
		close(rootGroupStopped)
	}()
	defer func() {
		select {
		case <-rootGroupStopped:
		case <-time.After(time.Millisecond * 300):
		}
	}()
	var cancelRoot context.CancelFunc
	rootCtx, cancelRoot = context.WithCancel(context.Background())
	defer cancelRoot()

	m.Run()
}

func TestSimple141WithPanic(t *testing.T) {
	sup := supervisor.NewSupervisor(supervisor.NewExeCtx(rootCtx, rootGroup), true)

	var cnt int64
	sup.Simple141(3, time.Millisecond, func() error {
		atomic.AddInt64(&cnt, 1)
		panic("N/A")
	})

	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, int64(3), cnt)
}

func TestSimple141WithError(t *testing.T) {
	sup := supervisor.NewSupervisor(supervisor.NewExeCtx(rootCtx, rootGroup), true)

	var cnt int64
	sup.Simple141(3, time.Millisecond, func() error {
		atomic.AddInt64(&cnt, 1)
		return fmt.Errorf("ERR:N/A")
	})

	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, int64(3), cnt)
}

func TestSimple141EnsureStarted(t *testing.T) {
	sup := supervisor.NewSupervisor(supervisor.NewExeCtx(rootCtx, rootGroup), true)

	var cnt int64
	sup.Simple141(3, time.Millisecond, func() error {
		atomic.AddInt64(&cnt, 1)
		return nil
	})

	assert.Equal(t, int64(1), atomic.LoadInt64(&cnt))
}

func TestSimple141FireAndForget(t *testing.T) {
	sup := supervisor.NewSupervisor(supervisor.NewExeCtx(rootCtx, rootGroup), false)

	var cnt int64
	sup.Simple141(3, time.Millisecond, func() error {
		atomic.AddInt64(&cnt, 1)
		return fmt.Errorf("ERR:N/A")
	})

	assert.Equal(t, int64(0), atomic.LoadInt64(&cnt))
}

func TestOne4All(t *testing.T) {
	wg := new(sync.WaitGroup)
	sup := supervisor.NewSupervisor(supervisor.NewExeCtx(rootCtx, wg), true)

	var cnt int64
	f := func(ctx context.Context) error {
		atomic.AddInt64(&cnt, 1)
		return nil
	}
	sup.One4All(3, time.Millisecond, f, f, f)

	wg.Wait()
	assert.Equal(t, int64(3), atomic.LoadInt64(&cnt))
}

func TestOne4AllWithError(t *testing.T) {
	wg := new(sync.WaitGroup)
	sup := supervisor.NewSupervisor(supervisor.NewExeCtx(rootCtx, wg), true)

	var cnt int64
	f := func(ctx context.Context) error {
		cc := atomic.AddInt64(&cnt, 1)
		runtime.Gosched()
		if 4 == cc || 8 == cc {
			return errors.New("SHOULD NOT 4")
		}
		return nil
	}
	go sup.One4All(6, time.Millisecond*10, f, f, f, f, f, f)

	<-time.After(time.Millisecond * 200)

	wg.Wait()
	assert.Equal(t, int64(18), atomic.LoadInt64(&cnt))
}

func TestOne4Rest(t *testing.T) {
	wg := new(sync.WaitGroup)
	sup := supervisor.NewSupervisor(supervisor.NewExeCtx(rootCtx, wg), true)

	delay := make(chan int, 100)
	lastDelay := 10
	go func() {
		for {
			lastDelay += 10
			delay <- lastDelay
		}
	}()

	var cnt int64
	f := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(<-delay) * time.Millisecond):
		}
		cc := atomic.LoadInt64(&cnt)
		if cc == 3 || cc == 6 || cc == 18 {
			atomic.CompareAndSwapInt64(&cnt, cc, cc+1)
			return errors.New("SHOULD NOT 4")
		}
		atomic.AddInt64(&cnt, 1)
		return nil
	}
	go sup.Rest4One(6, time.Millisecond*10, f, f, f, f, f, f, f, f, f, f, f, f)

	<-time.After(time.Millisecond * 1200)

	wg.Wait()
	assert.Equal(t, int64(31), atomic.LoadInt64(&cnt))
}
