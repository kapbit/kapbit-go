package swtch

import (
	"log"
	"time"
)

const (
	UndefinedLevel int = iota
	NormalLevel
	SlowLevel
)

// type SwitchOptions struct {
// 	Fast time.Duration
// 	Slow time.Duration
// }

type Switch struct {
	options Options
}

func NewSwitch(opts ...SetOption) (policy Switch, err error) {
	o := Options{}
	Apply(&o, opts...)
	o.SetDefaults()
	if err = o.Validate(); err != nil {
		return
	}
	policy = Switch{options: o}
	return
}

func (s Switch) NextDelay(level int) time.Duration {
	switch level {

	case UndefinedLevel:
		log.Printf("[Switch] Undefined level, use SlowInterval")
		return s.options.SlowInterval

	case NormalLevel:
		return s.options.NormalInterval

	case SlowLevel:
		return s.options.SlowInterval

	default:
		log.Printf("[Switch] Unknown level %d, use SlowInterval", level)
		return s.options.SlowInterval
	}
}
