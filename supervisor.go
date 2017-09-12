// Package supervisor provides supervision utilities
package supervisor

import (
	"time"

	"github.com/dc0d/club/errors"
)

//-----------------------------------------------------------------------------

type supervisorConf struct {
	intensity int
	period    time.Duration
	onError   func(error)
}

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

// Supervise a helper method, runs in sync, use as "go Supervise(...)",
// takes care of restarts (in case of panic or error),
// can be used as a SimpleOneForOne supervisor, an intensity < 0 means restart forever.
// It is sync because in most cases we do not need to escape the scope of current
// goroutine.
func Supervise(
	action func() error,
	options ...Option) {
	var sc supervisorConf
	for _, opt := range options {
		sc = opt(sc)
	}
	if sc.intensity == 0 {
		sc.intensity = 1
	}
	if sc.period == 0 {
		sc.period = time.Second * 5
	}

	intensity := sc.intensity
	period := sc.period
	onError := sc.onError

	if intensity != 1 && period <= 0 {
		period = time.Second * 5
	}

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
func run(action func() error) (errrun error) {
	defer func() {
		if e := recover(); e != nil {
			if err, ok := e.(error); ok {
				errrun = err
				return
			}
			errrun = errors.Errorf("UNKNOWN: %v", e)
		}
	}()
	return action()
}

//-----------------------------------------------------------------------------
