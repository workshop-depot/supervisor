// Package supervisor provides supervision utilities for running a function,
// and retry it if there is an error or panic, for a certain number of times (or infinitely),
// periodically, at the specified periods.
package supervisor

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

//-----------------------------------------------------------------------------

type supervisorConf struct {
	intensity int
	period    time.Duration
	onError   func(error)
}

//-----------------------------------------------------------------------------

// Option represents a supervision option
type Option func(supervisorConf) supervisorConf

// Intensity of -1 means run worker forever.
func Intensity(intensity int) Option {
	return func(sc supervisorConf) supervisorConf {
		sc.intensity = intensity
		return sc
	}
}

// Period of time to sleep between restarts
func Period(period time.Duration) Option {
	return func(sc supervisorConf) supervisorConf {
		sc.period = period
		return sc
	}
}

// OnError will get called in case of an error
func OnError(onError func(error)) Option {
	return func(sc supervisorConf) supervisorConf {
		sc.onError = onError
		return sc
	}
}

//-----------------------------------------------------------------------------

// Supervise a helper method, runs in sync, use as "go Supervise(...)" if should run concurrently,
// takes care of restarts (in case of panic or error),
// can be used as a SimpleOneForOne supervisor, an intensity < 0 means restart forever.
// It is sync because in most cases we do not need to escape the scope of current
// goroutine.
func Supervise(
	action func() error,
	options ...Option) {
	intensity, period, onError := validateAndRefine(options...)

	for intensity != 0 {
		if intensity > 0 {
			intensity--
		}
		if err := run(action); err != nil {
			if onError != nil {
				onError(err)
			}
			if intensity != 0 {
				time.Sleep(period)
			}
		} else {
			break
		}
	}
}

// run calls the function, does captures panics
func run(action func() error) (errRun error) {
	defer func() {
		if e := recover(); e != nil {
			if err, ok := e.(error); ok {
				errRun = err
				return
			}
			errRun = errors.Errorf("UNKNOWN: %+v", e)
		}
	}()
	return action()
}

func validateAndRefine(options ...Option) (int, time.Duration, func(error)) {
	var sc supervisorConf
	for _, opt := range options {
		sc = opt(sc)
	}
	if sc.intensity == 0 {
		sc.intensity = 1
	}
	if sc.intensity != 1 && sc.period <= 0 {
		sc.period = time.Second * 5
	}
	return sc.intensity,
		sc.period,
		sc.onError
}

//-----------------------------------------------------------------------------

func init() {
	fmt.Println("warning: use package github.com/dc0d/retry instead.")
}

//-----------------------------------------------------------------------------
