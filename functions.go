package supervisor

import (
	"fmt"
	"time"
)

//-----------------------------------------------------------------------------

// Supervise a helper method, runs in sync, use as "go Supervise(...)",
// takes care of restarts (in case of panic or error),
// can be used as a SimpleOneForOne supervisor, an intensity < 0 means restart forever
func Supervise(
	action func() error,
	intensity int,
	period time.Duration,
	onError func(error)) {
	if intensity > 1 && period <= 0 {
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

//-----------------------------------------------------------------------------

// run calls the function, does captures panics
func run(action func() error) (errrun error) {
	defer func() {
		if e := recover(); e != nil {
			if err, ok := e.(error); ok {
				errrun = err
				return
			}
			errrun = errorf("UNKNOWN: %v", e)
		}
	}()
	return action()
}

//-----------------------------------------------------------------------------

// errorf value type (string) error
func errorf(format string, a ...interface{}) error {
	return errString(fmt.Sprintf(format, a...))
}

//-----------------------------------------------------------------------------
