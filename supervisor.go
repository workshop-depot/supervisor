// Package supervisor provides supervision utilities, for Go
package supervisor

import (
	"context"
	"sync"
	"time"

	"github.com/dc0d/goroutines"
)

// ExeCtx provides an execution context, including an context.Context and a *sync.WaitGroup.
type ExeCtx struct {
	Ctx context.Context
	WG  *sync.WaitGroup
}

// NewExeCtx creates an instance of ExeCtx
func NewExeCtx(ctx context.Context, wg *sync.WaitGroup) *ExeCtx {
	res := &ExeCtx{
		Ctx: ctx,
		WG:  wg,
	}
	return res
}

// Supervisor supervises other goroutines
type Supervisor struct {
	exectx        *ExeCtx
	ensureStarted bool
}

// NewSupervisor creates an instance of Supervisor
func NewSupervisor(exectx *ExeCtx, ensureStarted bool) *Supervisor {
	res := new(Supervisor)
	res.exectx = exectx
	res.ensureStarted = ensureStarted
	return res
}

func (sp *Supervisor) defaultRetry(e interface{}, intensity int, fn func() error, dt time.Duration) {
	// log.Println("error:", e)
	time.Sleep(dt)
	go sp.sup(intensity, dt, fn, sp.defaultRetry)
}

func (sp *Supervisor) sup(
	intensity int,
	period time.Duration,
	fn func() error,
	retry func(interface{}, int, func() error, time.Duration)) {
	select {
	case <-sp.exectx.Ctx.Done():
		return
	default:
	}
	if intensity == 0 {
		return
	}
	if intensity > 0 {
		intensity--
	}
	dt := time.Second
	if period > 0 {
		dt = period
	}
	if retry == nil {
		retry = sp.defaultRetry
	}
	internalRetry := func(e1 interface{}) {
		retry(e1, intensity, fn, dt)
	}
	utl := goroutines.New()
	if sp.ensureStarted {
		utl = utl.EnsureStarted()
	}
	utl.AddToGroup(sp.exectx.WG).
		Recover(func(e interface{}) {
			internalRetry(e)
		}).
		Go(func() {
			if err := fn(); err != nil {
				internalRetry(err)
			}
		})
}

// Simple141 provides simple one for one supervision. period must be
// greater than zero to take effect. An intensity of -1 means run forever.
func (sp *Supervisor) Simple141(
	intensity int,
	period time.Duration,
	fn func() error) {
	retry := func(e1 interface{}, intensity1 int, fn1 func() error, dt1 time.Duration) {
		// log.Println("error:", e)
		time.Sleep(dt1)
		go sp.Simple141(intensity1, dt1, fn1)
	}
	sp.sup(intensity, period, fn, retry)
}

func (sp *Supervisor) one4All(fn ...func(context.Context) error) (ferr error) {
	all := len(fn)
	if all == 0 {
		return nil
	}

	subCtx, subCancel := context.WithCancel(sp.exectx.Ctx)
	defer subCancel()

	seterr := make(chan error, all)
	for _, v := range fn {
		v := v
		sp.Simple141(1, time.Millisecond, func() (cerr error) {
			defer func() {
				seterr <- cerr
				if cerr != nil {
					subCancel()
				}
			}()
			cerr = v(subCtx)
			return
		})
	}

	cnt := 0
	for {
		_err := <-seterr
		if _err != nil {
			ferr = _err
		}
		cnt++
		if cnt == len(fn) {
			break
		}
	}

	return
}

// One4All if one of goroutines stops, all other will get the signal to stop too.
func (sp *Supervisor) One4All(
	intensity int,
	period time.Duration,
	fn ...func(context.Context) error) {
	sp.Simple141(intensity, period, func() error {
		return sp.one4All(fn...)
	})
}

func (sp *Supervisor) rest4One(fn ...func(context.Context) error) (ferr error) {
	all := len(fn)
	if all == 0 {
		return nil
	}

	rootCtx, rootCancel := context.WithCancel(sp.exectx.Ctx)
	defer rootCancel()

	seterr := make(chan error, all)
	var _subCtx = rootCtx
	for _, v := range fn {
		subCtx, subCancel := context.WithCancel(_subCtx)
		v := v
		sp.Simple141(1, time.Millisecond, func() (cerr error) {
			defer func() {
				seterr <- cerr
				if cerr != nil {
					subCancel()
				}
			}()
			cerr = v(subCtx)
			return
		})
		_subCtx = subCtx
	}

	cnt := 0
	for {
		_err := <-seterr
		if _err != nil {
			ferr = _err
		}
		cnt++
		if cnt == len(fn) {
			break
		}
	}

	return
}

// Rest4One rest for one
func (sp *Supervisor) Rest4One(
	intensity int,
	period time.Duration,
	fn ...func(context.Context) error) {
	sp.Simple141(intensity, period, func() error {
		return sp.rest4One(fn...)
	})
}
