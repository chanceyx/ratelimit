package ratelimit

import (
	"sync/atomic"
	"time"
	"unsafe"
)

type windowState struct {
	last            time.Time
	availableTokens int64
}

type atomicSlideWindow struct {
	state unsafe.Pointer
	padding     [56]byte

	perRequest time.Duration
	windowSize int64
	capacity   int64
	clock      Clock
}

func newAtomicSlideWindowBase(rate int, opts ...Option) *atomicSlideWindow {
	config := getDefaultConfig(opts)
	perRequest := config.per / time.Duration(rate)
	l := &atomicSlideWindow{
		perRequest: perRequest,
		windowSize: config.windowSize,
		capacity:   config.capacity,
		clock:      config.clock,
	}

	initialState := windowState{
		last:            time.Time{},
		availableTokens: l.capacity,
	}

	atomic.StorePointer(&l.state, unsafe.Pointer(&initialState))
	return l
}

func (t *atomicSlideWindow) Take(count ...int64) time.Time {
	var (
		newState windowState
		taken    bool
		interval time.Duration
	)

	tokens := count[0]
	for !taken {
		now := t.clock.Now()

		previousStatePointer := atomic.LoadPointer(&t.state)
		oldState := (*bucketState)(previousStatePointer)

		newState = windowState{
			last:            now,
			availableTokens: oldState.availableTokens,
		}

		if oldState.last.IsZero() {
			taken = atomic.CompareAndSwapPointer(&t.state, previousStatePointer, unsafe.Pointer(&newState))
			continue
		}

		newState.availableTokens += int64(now.Sub(oldState.last)/t.perRequest) - tokens

		if newState.availableTokens > t.capacity {
			newState.availableTokens = t.capacity
		}

		if newState.availableTokens < 0 {
			newState.last = newState.last.Add(time.Duration(-1*newState.availableTokens) * t.perRequest)
			interval, newState.availableTokens = time.Duration(-1*newState.availableTokens)*t.perRequest, 0
		}

		taken = atomic.CompareAndSwapPointer(&t.state, previousStatePointer, unsafe.Pointer(&newState))
	}
	t.clock.Sleep(interval)
	return newState.last
}