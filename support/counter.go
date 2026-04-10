package support

type Counter struct {
	threshold int
	count     int
}

func NewCounter(threshold int) *Counter {
	return &Counter{threshold: threshold}
}

func NewCounterWith(threshold, n int) *Counter {
	return &Counter{threshold: threshold, count: n}
}

func (c *Counter) Increment() (ok bool) {
	if c.count < c.threshold {
		c.count++
		ok = true
	}
	return
}

func (c *Counter) Decrement() (ok bool) {
	if c.count > 0 {
		c.count--
		ok = true
	}
	return
}

func (c *Counter) Count() int {
	return c.count
}
