package support

import (
	"math/rand"
	"time"
)

func ApplyEqualJitter(interval time.Duration) time.Duration {
	// 1. Safety Check: Avoid panic in rand.Int63n if interval is <= 0
	if interval <= 0 {
		return 0
	}
	half := int64(interval / 2)
	// 2. Generate random jitter.
	// We add +1 because Int63n is exclusive [0, n).
	// Range becomes [0, half]
	jitter := rand.Int63n(half + 1)
	// 3. Result range: [half, half + half] -> [interval/2, interval]
	return time.Duration(half + jitter)
}
