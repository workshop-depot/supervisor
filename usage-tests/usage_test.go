package usagetests

import (
	"context"
	"errors"
	"fmt"
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
		atomic.AddInt64(&cnt, 1)
		cc := atomic.LoadInt64(&cnt)
		if 4 == cc || 8 == cc {
			return errors.New("SHOULD NOT 4")
		}
		return nil
	}
	go sup.One4All(6, time.Millisecond*10, f, f, f, f, f, f)

	<-time.After(time.Millisecond * 200)

	wg.Wait()
	assert.Equal(t, int64(14), atomic.LoadInt64(&cnt))
}
