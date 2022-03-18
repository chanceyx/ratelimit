package ratelimit

import (
	"time"

	"github.com/andres-erbsen/clock"
)

const (
	DISABLE = iota
)

type config struct {
	clock      Clock
	slack      int
	per        time.Duration
	capacity   int64
	windowSize int64
}

func getDefaultConfig(opts []Option) config {
	c := config{
		clock:      clock.New(),
		slack:      10,
		per:        time.Second,
		capacity:   DISABLE,
		windowSize: DISABLE,
	}

	for _, opt := range opts {
		opt.apply(&c)
	}
	return c
}

func getTokenBucketConfig(opts []Option) config {
	c := config{
		clock:      clock.New(),
		slack:      DISABLE,
		per:        time.Second,
		capacity:   10,
		windowSize: DISABLE,
	}

	for _, opt := range opts {
		opt.apply(&c)
	}
	return c
}

func getSlideWindowConfig(opts []Option) config {
	c := config{
		clock:      clock.New(),
		slack:      DISABLE,
		per:        time.Second,
		capacity:   10,
		windowSize: 10,
	}

	for _, opt := range opts {
		opt.apply(&c)
	}
	return c
}

type clockOption struct {
	clock Clock
}

func (o clockOption) apply(c *config) {
	c.clock = o.clock
}

func WithClock(clock Clock) Option {
	return clockOption{clock: clock}
}

type slackOption int

func (o slackOption) apply(c *config) {
	c.slack = int(o)
}

func Per(per time.Duration) Option {
	return perOption(per)
}

type perOption time.Duration

func (o perOption) apply(c *config) {
	c.per = time.Duration(o)
}

type windowSizeOption int64

func (o windowSizeOption) apply(c *config) {
	c.windowSize = int64(o)
}

func WithWindowSize(w int64) Option {
	return windowSizeOption(w)
}

type capacityOption int64

func (o capacityOption) apply(c *config) {
	c.capacity = int64(o)
}

func WithCapacity(c int64) Option {
	return capacityOption(c)
}

var WithoutSlack Option = slackOption(0)

func WithSlack(slack int) Option {
	return slackOption(slack)
}

type unlimited struct{}

func NewUnlimited() Limiter {
	return unlimited{}
}

func (unlimited) Take(count ...int64) time.Time {
	return time.Now()
}

func (unlimited) TryTake(count ...int64) (time.Duration, time.Time) {
	return time.Duration(0), time.Now()
}

func (unlimited) Wait(duration time.Duration) {
	time.Sleep(duration)
}
