// Package supervisor provides supervisor trees, for Go
package supervisor

import (
	"context"
	"sync"
	"time"

	"github.com/dc0d/goroutines"
)

var (
	_dummy = goroutines.New()
)

type child struct{ intensity int }

type Supervisor struct {
	// ptr
	ctx context.Context
	wg  *sync.WaitGroup

	// val
	period time.Duration
	child
}

func New() *Supervisor {
	res := &Supervisor{
		period: time.Second * 5,
	}
	res.intensity = -1
	return res
}

func (s *Supervisor) Intensity(intensity int) *Supervisor {
	s.intensity = intensity
	return s
}

func (s *Supervisor) WaitGroup(wg *sync.WaitGroup) *Supervisor {
	s.wg = wg
	return s
}

func (s *Supervisor) Period(period time.Duration) *Supervisor {
	s.period = period
	return s
}

// Supervise acts as one for one and simple one for one (not much practical
// usage in having both)
func (s Supervisor) Supervise(ctx context.Context, children ...Tree) {
	if s.intensity < -1 {
		s.intensity = -1
	}
	if s.period <= 0 {
		s.period = time.Second * 5
	}
	s.ctx = ctx
	for _, v := range children {
		v := v
		news := s
		news.supervise(v)
	}
}

func (s Supervisor) supervise(child Tree) {
	if s.intensity == 0 {
		return
	}
	if s.intensity > 0 {
		s.intensity--
	}
	goroutines.New().
		WaitGroup(s.wg).
		WaitStart().
		Recover(func(e interface{}) {
			time.Sleep(s.period)
			go s.supervise(child)
		}).
		Go(func() {
			child(s.ctx)
		})
}

// Tree is the recursive type helps with creating trees of supervisors
type Tree func(context.Context, ...Tree)

// this package was a study, not much usage in it's current form. full blown
// solutions such as proto actor serve difference scenarious far better, and
// using an interface for actors (calling methods on each message) is more
// logical, because we can not force the user to respect <-ctx.Done() and that
// will make a mess of things
