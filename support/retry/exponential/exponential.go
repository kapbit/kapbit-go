package exponential

import (
	"math"
	"time"
)

const UndefinedLevel = 0

type Exponential struct {
	options Options
}

func NewExponential(opts ...SetOption) (policy Exponential, err error) {
	o := DefaultOptions()
	if err = Apply(&o, opts...); err != nil {
		return
	}
	policy = Exponential{o}
	return
}

func MustNewExponential(opts ...SetOption) Exponential {
	policy, err := NewExponential(opts...)
	if err != nil {
		panic(err)
	}
	return policy
}

func (e Exponential) NextDelay(level int) time.Duration {
	// Cap the level to avoid massive math operations.
	// 2^63 is the limit for int64, so anything above 64 is pointless
	// to calculate if we are just going to cap it at MaxInterval anyway.
	const safeLevelCap = 62

	if level > safeLevelCap {
		// If we are this deep, we are definitely at MaxInterval
		return e.options.MaxInterval
	}

	// Level 0 usually means "First attempt" or "Reset".
	// Depending on preference, this could be 0 or Initial.
	if level <= 0 {
		return e.options.InitialInterval
	}

	delay := float64(e.options.InitialInterval) * math.Pow(e.options.Multiplier, float64(level))
	if delay > float64(e.options.MaxInterval) {
		return e.options.MaxInterval
	}
	return time.Duration(delay)
}
