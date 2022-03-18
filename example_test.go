package ratelimit_test

import (
	"fmt"
	"github.com/chanceyx/ratelimit"
	"time"
)

func Example() {
	rl := ratelimit.New(100, ratelimit.DEFAULTLIMITER) // per second

	prev := time.Now()
	for i := 0; i < 10; i++ {
		now := rl.Take()
		if i > 0 {
			fmt.Println(i, now.Sub(prev))
		}
		prev = now
	}

	// Output:
	// 1 10ms
	// 2 10ms
	// 3 10ms
	// 4 10ms
	// 5 10ms
	// 6 10ms
	// 7 10ms
	// 8 10ms
	// 9 10ms
}
