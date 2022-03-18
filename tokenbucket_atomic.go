package ratelimit

import (
	"time"
	"unsafe"

	"sync/atomic"
)

type bucketState struct {
	last            time.Time
	availableTokens int64
}

type atomicTokenBucket struct {
	state   unsafe.Pointer
	padding [56]byte

	capacity   int64
	perRequest time.Duration
	clock      Clock
}

func newAtomicTokenBucketBased(rate int, opts ...Option) *atomicTokenBucket {
	config := getTokenBucketConfig(opts)
	perRequest := config.per / time.Duration(rate)
	l := &atomicTokenBucket{
		perRequest: perRequest,
		capacity:   config.capacity,
		clock:      config.clock,
	}

	initialState := bucketState{
		last:            time.Time{},
		availableTokens: 0,
	}

	atomic.StorePointer(&l.state, unsafe.Pointer(&initialState))
	return l
}

func (t *atomicTokenBucket) Take(count ...int64) time.Time {
	var (
		newState bucketState
		taken    bool
		interval time.Duration
	)

	tokens := count[0]
	for !taken {
		now := t.clock.Now()

		previousStatePointer := atomic.LoadPointer(&t.state)
		oldState := (*bucketState)(previousStatePointer)

		newState = bucketState{
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

func (t *atomicTokenBucket) Wait(d time.Duration) {
	if d > time.Duration(0) {
		t.clock.Sleep(d)
	}
}

func (t *atomicTokenBucket) TryTake(count ...int64) (time.Duration, time.Time) {
	var (
		newState bucketState
		taken    bool
		interval time.Duration
	)
	tokens := count[0]
	for !taken {
		now := t.clock.Now()

		previousStatePointer := atomic.LoadPointer(&t.state)
		oldState := (*bucketState)(previousStatePointer)

		newState = bucketState{
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

	return interval, newState.last
}
