package ratelimit

import (
	"time"
	"unsafe"

	"sync/atomic"
)

type state struct {
	last     time.Time
	sleepFor time.Duration
}

type atomicLimiter struct {
	state   unsafe.Pointer
	padding [56]byte

	perRequest time.Duration
	maxSlack   time.Duration
	clock      Clock
}

func newAtomicBased(rate int, opts ...Option) *atomicLimiter {
	config := getDefaultConfig(opts)
	perRequest := config.per / time.Duration(rate)
	l := &atomicLimiter{
		perRequest: perRequest,
		maxSlack:   -1 * time.Duration(config.slack) * perRequest,
		clock:      config.clock,
	}

	initialState := state{
		last:     time.Time{},
		sleepFor: 0,
	}

	atomic.StorePointer(&l.state, unsafe.Pointer(&initialState))
	return l
}

func (t *atomicLimiter) Take(count ...int64) time.Time {
	var (
		newState state
		taken    bool
		interval time.Duration
	)

	for !taken {
		now := t.clock.Now()

		previousStatePointer := atomic.LoadPointer(&t.state)
		oldState := (*state)(previousStatePointer)

		newState = state{
			last:     now,
			sleepFor: oldState.sleepFor,
		}

		if oldState.last.IsZero() {
			taken = atomic.CompareAndSwapPointer(&t.state, previousStatePointer, unsafe.Pointer(&newState))
			continue
		}

		newState.sleepFor += t.perRequest - now.Sub(oldState.last)

		if newState.sleepFor < t.maxSlack {
			newState.sleepFor = t.maxSlack
		}
		if newState.sleepFor > 0 {
			newState.last = newState.last.Add(newState.sleepFor)
			interval, newState.sleepFor = newState.sleepFor, 0
		}
		taken = atomic.CompareAndSwapPointer(&t.state, previousStatePointer, unsafe.Pointer(&newState))
	}
	t.clock.Sleep(interval)
	return newState.last
}

func (t *atomicLimiter) Wait(d time.Duration) {
	if d > time.Duration(0) {
		t.clock.Sleep(d)
	}
}

func (t *atomicLimiter) TryTake(count ...int64) (time.Duration, time.Time) {
	var (
		newState state
		taken    bool
		interval time.Duration
	)

	for !taken {
		now := t.clock.Now()

		previousStatePointer := atomic.LoadPointer(&t.state)
		oldState := (*state)(previousStatePointer)

		newState = state{
			last:     now,
			sleepFor: oldState.sleepFor,
		}

		if oldState.last.IsZero() {
			taken = atomic.CompareAndSwapPointer(&t.state, previousStatePointer, unsafe.Pointer(&newState))
			continue
		}

		newState.sleepFor += t.perRequest - now.Sub(oldState.last)

		if newState.sleepFor < t.maxSlack {
			newState.sleepFor = t.maxSlack
		}
		if newState.sleepFor > 0 {
			newState.last = newState.last.Add(newState.sleepFor)
			interval, newState.sleepFor = newState.sleepFor, 0
		}
		taken = atomic.CompareAndSwapPointer(&t.state, previousStatePointer, unsafe.Pointer(&newState))
	}

	return interval, newState.last
}
