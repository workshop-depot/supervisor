package supervisor

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestIntensityError(t *testing.T) {
	var sum int64
	Supervise(func() error {
		atomic.AddInt64(&sum, 1)
		return errors.Errorf("DUMMY")
	},
		Intensity(3),
		Period(time.Millisecond*50))
	assert.Equal(t, int64(3), sum)
}

func TestIntensityPanic(t *testing.T) {
	var sum int64
	Supervise(func() error {
		atomic.AddInt64(&sum, 1)
		panic(errors.Errorf("DUMMY"))
	},
		Intensity(3),
		Period(time.Millisecond*50))
	assert.Equal(t, int64(3), sum)
}

func TestIntensityNoError(t *testing.T) {
	var sum int64
	Supervise(func() error {
		atomic.AddInt64(&sum, 1)
		return nil
	},
		Intensity(3),
		Period(time.Millisecond*50))
	assert.Equal(t, int64(1), sum)
}

func TestOnError(t *testing.T) {
	var sum int64
	Supervise(func() error {
		atomic.AddInt64(&sum, 1)
		panic("X")
	},
		Intensity(3),
		Period(time.Millisecond*50),
		OnError(func(error) { atomic.AddInt64(&sum, 1) }))
	assert.Equal(t, int64(6), sum)
}

func TestOnError2(t *testing.T) {
	var sum int64
	Supervise(func() error {
		atomic.AddInt64(&sum, 1)
		panic("X")
	},
		Intensity(3000),
		Period(time.Microsecond),
		OnError(func(error) { atomic.AddInt64(&sum, 1) }))
	assert.Equal(t, int64(6000), sum)
}

func TestOnError3(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var sum int64
	go Supervise(func() error {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		atomic.AddInt64(&sum, 1)
		panic("X")
	},
		Intensity(3),
		Period(time.Millisecond*50),
		OnError(func(error) { cancel() }))
	<-ctx.Done()
	assert.Equal(t, int64(1), sum)
}

func ExampleSupervise() {
	// if should run concurrently then use "go Supervise(...)"
	Supervise(func() error {
		fmt.Println("done")
		return nil
	})

	// Output:
	// done
}

func ExampleIntensity() {
	Supervise(func() error {
		fmt.Println(1)
		return errors.Errorf("FAILED")
	},
		Intensity(3),
		Period(time.Millisecond*50))

	// Output:
	// 1
	// 1
	// 1
}

func ExampleOnError_error() {
	Supervise(func() error {
		return errors.Errorf("FAILED")
	},
		Intensity(3),
		Period(time.Millisecond*50),
		OnError(func(err error) { fmt.Println(err) }))

	// Output:
	// FAILED
	// FAILED
	// FAILED
}

func ExampleOnError_panic() {
	Supervise(func() error {
		panic(errors.Errorf("FAILED"))
	},
		Intensity(3),
		Period(time.Millisecond*50),
		OnError(func(err error) { fmt.Println(err) }))

	// Output:
	// FAILED
	// FAILED
	// FAILED
}

func ExamplePeriod() {
	startedAt := time.Now()
	Supervise(func() error {
		return errors.Errorf("FAILED")
	},
		Intensity(3),
		Period(time.Millisecond*50),
		OnError(func(err error) { fmt.Println(time.Since(startedAt).Round(time.Millisecond)) }))

	// Output:
	// 0s
	// 50ms
	// 100ms
}
