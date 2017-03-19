// Package supervisor provides supervisor trees, for Go
package supervisor

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Tree is the recursive type helps with creating trees of supervisors
type Tree func(context.Context, ...Tree)

// Supervise starts up a new supervisor with given children
func Supervise(ctx context.Context, children ...Tree) {
	op := getOptions(ctx)

	switch op.strategy {
	case OneForOne:
		superviseOneForOne(ctx, op, children...)
	case OneForAll:
		superviseOneForAll(ctx, op, children...)
	case SimpleOneForOne:
		superviseSimpleOneForOne(ctx, op, children...)
	}
}

func superviseSimpleOneForOne(ctxSrc context.Context, op options, children ...Tree) {
	var idSeq int64
	repo := make(map[int64]*child)
	stoppedID := make(chan int64)

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctxSrc)
	defer cancel()

	register := func(newChild Tree) (*child, int64) {
		id := atomic.AddInt64(&idSeq, 1)
		c := &child{
			child:     newChild,
			intensity: op.intensity,
		}
		repo[id] = c
		return c, id
	}

	for _, cv := range children {
		cv := cv
		v, id := register(cv)
		startSimpleOneForOne(ctx, op, wg, stoppedID, v, id)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case <-ctx.Done():
			return
		case id := <-stoppedID:
			cld, ok := repo[id]
			if !ok {
				continue
			}
			if cld == nil {
				delete(repo, id)
				continue
			}
			startSimpleOneForOne(ctx, op, wg, stoppedID, cld, id)
			cld.intensity--
			if cld.intensity <= 0 {
				delete(repo, id)
			}
		case newChild, ok := <-op.sofo:
			if !ok {
				return
			}
			if newChild == nil {
				continue
			}
			v, id := register(newChild)
			startSimpleOneForOne(ctx, op, wg, stoppedID, v, id)
		}
	}
}

func startSimpleOneForOne(
	ctx context.Context,
	op options,
	wg *sync.WaitGroup,
	stoppedIndex chan int64,
	cld *child,
	index int64) {
	if cld.intensity <= 0 {
		return
	}
	intensity := cld.intensity
	f := cld.child

	wg.Add(1)
	waitStart(func() {
		defer wg.Done()
		defer func() {
			if e := recover(); e != nil {
				// TODO:
			}
			stoppedIndex <- index
		}()
		if intensity < op.intensity {
			<-time.After(op.period)
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
		f(ctx)
	})
}

func superviseOneForAll(ctxSrc context.Context, op options, children ...Tree) {
	list := makeList(op, children...)

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctxSrc)
	defer cancel()
	for k, v := range list {
		k, v := k, v
		startOneForAll(ctx, cancel, op, wg, v, k)
	}

	ticks := time.NewTicker(time.Second * 5)
	defer ticks.Stop()

LOOP_OFA:
	for {
		if any := sort.Search(len(list), func(ix int) bool { return list[ix] != nil }); len(list) <= any {
			defer cancel()
			return
		}

		select {
		case <-ctxSrc.Done():
			return
		default:
		}

		select {
		case <-ticks.C:
		case <-ctxSrc.Done():
			return
		case <-ctx.Done():
			if any := sort.Search(len(list), func(ix int) bool { return list[ix] != nil }); len(list) <= any {
				defer cancel()
				return
			}

			wg.Wait()

			ctx, cancel = context.WithCancel(ctxSrc)
			// defer cancel()
			for k, v := range list {
				k, v := k, v
				if v == nil {
					continue
				}
				startOneForAll(ctx, cancel, op, wg, v, k)
				v.intensity--
				if v.intensity < 0 {
					list[k] = nil
				}

				select {
				case <-ctx.Done():
					continue LOOP_OFA
				default:
				}
			}
		}
	}
}

func startOneForAll(
	ctx context.Context,
	cancel context.CancelFunc,
	op options,
	wg *sync.WaitGroup,
	cld *child,
	index int) {
	if cld.intensity <= 0 {
		return
	}
	intensity := cld.intensity
	f := cld.child

	wg.Add(1)
	waitStart(func() {
		defer wg.Done()
		defer cancel()
		defer func() {
			if e := recover(); e != nil {
				// TODO:
			}
		}()
		if intensity < op.intensity {
			<-time.After(op.period)
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
		f(ctx)
	})
}

func superviseOneForOne(ctxSrc context.Context, op options, children ...Tree) {
	list := makeList(op, children...)
	stoppedIndex := make(chan int, len(children))

	wg := &sync.WaitGroup{}
	defer wg.Wait()

	ctx, cancel := context.WithCancel(ctxSrc)
	defer cancel()

	for k, v := range list {
		k, v := k, v
		startOneForOne(ctx, op, wg, stoppedIndex, v, k)
	}

	for {
		if any := sort.Search(len(list), func(ix int) bool { return list[ix] != nil }); len(list) <= any {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case <-ctx.Done():
			return
		case i := <-stoppedIndex:
			cld := list[i]
			if cld == nil {
				continue
			}
			startOneForOne(ctx, op, wg, stoppedIndex, cld, i)
			cld.intensity--
			if cld.intensity <= 0 {
				list[i] = nil
			}
		}
	}
}

func startOneForOne(
	ctx context.Context,
	op options,
	wg *sync.WaitGroup,
	stoppedIndex chan int,
	cld *child,
	index int) {
	if cld.intensity <= 0 {
		return
	}
	intensity := cld.intensity
	f := cld.child

	wg.Add(1)
	waitStart(func() {
		defer wg.Done()
		defer func() {
			if e := recover(); e != nil {
				// TODO:
			}
			stoppedIndex <- index
		}()
		if intensity < op.intensity {
			<-time.After(op.period)
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
		f(ctx)
	})
}

func waitStart(f func()) {
	started := make(chan struct{})
	go func() {
		close(started)
		f()
	}()
	<-started
}

func makeList(op options, children ...Tree) []*child {
	list := make([]*child, len(children))
	for k, v := range children {
		c := child{
			child:     v,
			intensity: op.intensity,
		}
		list[k] = &c
	}
	return list
}

type child struct {
	child     Tree
	intensity int
}

func getOptions(ctx context.Context) options {
	op := options{
		strategy:  OneForOne,
		intensity: 1,
		period:    time.Second * 5,
	}
	switch x := ctx.(type) {
	case *optionContext:
		op = x.options
	}

	return op
}

// WithOptions from a given context, creates a context with given options to
// be passed to Supervise function
func WithOptions(
	ctx context.Context,
	strategy Strategy,
	intensity int,
	period time.Duration) (context.Context, chan<- Tree) {
	op := options{
		strategy:  strategy,
		intensity: intensity,
		period:    period,
	}
	var sofoChildren chan Tree
	if strategy == SimpleOneForOne {
		sofoChildren = make(chan Tree)
		op.sofo = sofoChildren
	}
	nextCtx := &optionContext{
		Context: ctx,
		options: op,
	}
	return nextCtx, sofoChildren
}

type optionContext struct {
	context.Context
	options
}

type options struct {
	strategy  Strategy
	intensity int
	period    time.Duration
	sofo      <-chan Tree
}

// Strategy indicates the restart strategy
type Strategy int

// Valid Strategy values
const (
	OneForOne Strategy = iota + 1
	OneForAll
	SimpleOneForOne
)

// one_for_one
// one_for_all
// rest_for_one == go pipeline (in some cases)
// simple_one_for_one

// debug				(?)
// error handling chan	(?)

// Tree should return error?
