package bcc

import (
	"context"
	"net/url"
	"time"
)

type Arguments map[string]string

func Defaults() Arguments {
	return make(Arguments)
}

func (args Arguments) ToURLValues() url.Values {
	v := url.Values{}
	for key, value := range args {
		v.Set(key, value)
	}
	return v
}

func (args Arguments) merge(extraArgs []Arguments) {
	for _, extraArg := range extraArgs {
		for key, val := range extraArg {
			args[key] = val
		}
	}
}

// From https://github.com/aws/aws-sdk-go/blob/main/aws/context_sleep.go

// SleepWithContext will wait for the timer duration to expire, or the context
// is canceled. Which ever happens first. If the context is canceled the Context's
// error will be returned.
//
// Expects Context to always return a non-nil error if the Done channel is closed.
func SleepWithContext(ctx context.Context, dur time.Duration) error {
	t := time.NewTimer(dur)
	defer t.Stop()

	select {
	case <-t.C:
		break
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func loopWaitLock(manager *Manager, path string) (err error) {
	var wait struct {
		Locked bool `json:"locked"`
	}
	for {
		err = manager.Get(path, Defaults(), &wait)
		if err != nil {
			return
		}
		if !wait.Locked {
			break
		}
		time.Sleep(time.Second)
	}
	return
}
