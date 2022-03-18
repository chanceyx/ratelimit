package ratelimit

import (
	"time"
)

type LimiterType int

const (
	DEFAULTLIMITER = iota
	TOKENBUCKET
	SLIDEWINDOW
)

type Limiter interface {
	Take(...int64) time.Time
	TryTake(...int64) (time.Duration, time.Time)
	Wait(time.Duration)
}

type Clock interface {
	Now() time.Time
	Sleep(time.Duration)
}

type Option interface {
	apply(c *config)
}

func New(rate int, limiterType LimiterType, opts ...Option) Limiter {
	if limiterType == TOKENBUCKET {
		return newAtomicTokenBucketBased(rate, opts...)
	}
	return newAtomicBased(rate, opts...)
}
