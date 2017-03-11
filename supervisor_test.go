package supervisor

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	defer time.Sleep(time.Millisecond * 300)
	var cancelRoot context.CancelFunc
	rootCtx, cancelRoot = context.WithCancel(context.Background())
	defer cancelRoot()

	m.Run()
}

var (
	rootCtx context.Context
)

func TestOneForOne(t *testing.T) {
	names := make(chan string, 1000)
	iters := make(map[string]int)

	ctx := WithOptions(
		rootCtx,
		OneForOne,
		2,
		time.Millisecond*300)

	c1 := makeTest(0, 1, true, `C1`, names)
	c2 := makeTest(0, 1, true, `C2`, names)
	go Supervise(ctx, c1, c2)

L1:
	for {
		select {
		case n := <-names:
			prev, _ := iters[n]
			iters[n] = prev + 1
		case <-time.After(time.Second * 2):
			break L1
		}
	}
	if iters[`C1`] != 2 || iters[`C2`] != 2 {
		t.Fail()
	}
}

func TestOneForAll(t *testing.T) {
	names := make(chan string, 1000)
	iters := make(map[string]int)

	ctx := WithOptions(
		rootCtx,
		OneForAll,
		2,
		time.Millisecond*300)

	c1 := makeTest(0, 2, true, `C1`, names)
	c2 := makeTest(0, 1, true, `C2`, names)
	go Supervise(ctx, c1, c2)

L1:
	for {
		select {
		case n := <-names:
			prev, _ := iters[n]
			iters[n] = prev + 1
		case <-time.After(time.Second * 2):
			break L1
		}
	}
	if iters[`C1`] != 4 || iters[`C2`] != 3 {
		t.Fail()
	}
}

func makeTest(
	timeout time.Duration,
	countout int,
	doPanic bool,
	name string,
	iter chan<- string) Tree {
	return func(ctx context.Context, _ ...Tree) {
		var chtimeout <-chan time.Time
		var chcountout <-chan time.Time

	LOOP_TEST:
		for {
			iter <- name
			// select {
			// case iter <- name:
			// default:
			// }

			if timeout > 0 {
				chtimeout = time.After(timeout)
			}
			if countout > 0 {
				chcountout = time.After(time.Millisecond * 100)
			}

			select {
			case <-ctx.Done():
				return
			case <-chtimeout:
				break LOOP_TEST
			case <-chcountout:
				countout--
				if countout <= 0 {
					break LOOP_TEST
				}
			}
		}
		if doPanic {
			panic(`TEST`)
		}
	}
}

func timerScope(name string, opCount ...int) func() {
	log.Println(name, `started`)
	start := time.Now()
	return func() {
		elapsed := time.Now().Sub(start)
		log.Printf("%s took %v", name, elapsed)
		if len(opCount) == 0 {
			return
		}

		N := opCount[0]
		if N <= 0 {
			return
		}

		E := float64(elapsed)
		FRC := E / float64(N)

		log.Printf("op/sec %.2f", float64(N)/(E/float64(time.Second)))

		switch {
		case FRC > float64(time.Second):
			log.Printf("sec/op %.2f", (E/float64(time.Second))/float64(N))
		case FRC > float64(time.Millisecond):
			log.Printf("milli-sec/op %.2f", (E/float64(time.Millisecond))/float64(N))
		case FRC > float64(time.Microsecond):
			log.Printf("micro-sec/op %.2f", (E/float64(time.Microsecond))/float64(N))
		default:
			log.Printf("nano-sec/op %.2f", (E/float64(time.Nanosecond))/float64(N))
		}
	}
}

func waitFunc(f func(), exitDelay time.Duration) error {
	funcDone := make(chan struct{})
	go func() {
		defer close(funcDone)
		f()
	}()

	if exitDelay <= 0 {
		<-funcDone

		return nil
	}

	select {
	case <-time.After(exitDelay):
		return errors.New(`TIMEOUT`)
	case <-funcDone:
	}

	return nil
}
