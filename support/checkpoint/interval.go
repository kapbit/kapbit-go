package checkpoint

import "fmt"

func NewSingleInterval(tuple Tuple) *Interval {
	return &Interval{start: tuple, end: tuple}
}

type Interval struct {
	start Tuple
	end   Tuple
	prev  *Interval
	next  *Interval
}

func (i *Interval) Start() Tuple {
	return i.start
}

func (i *Interval) End() Tuple {
	return i.end
}

// func (i *Interval) Value() Tuple {
// 	return i.endValue
// }

func (i *Interval) Len() int64 {
	return i.end.N - i.start.N + 1
}

func (i *Interval) String() string {
	return fmt.Sprintf("[%v %v %p %p %p]", i.start, i.end, i, i.prev, i.next)
}

func (i *Interval) prevNum(n int64) bool {
	return n == i.start.N-1
}

func (i *Interval) nextNum(n int64) bool {
	return n == i.end.N+1
}

func (i *Interval) leftNum(n int64) bool {
	return n < i.start.N-1
}

func (i *Interval) rightNum(n int64) bool {
	return n > i.end.N+1
}

func (i *Interval) mergeRight() {
	next := i.next
	if i.end.V > next.start.V {
		panic(
			fmt.Sprintf(
				"monotonicity violation: right interval start value (%d) must be bigger than current end value (%d)",
				next.start.V, i.end.V,
			),
		)
	}

	i.end = next.end
	// i.endValue = next.endValue
	i.next = next.next
	if next.next != nil {
		next.next.prev = i
	}
}

func (i *Interval) increaseLeft(v int64) {
	if i.start.V < v {
		panic(
			fmt.Sprintf(
				"monotonicity violation: new start value (%d) must be smaller than existing one (%v)",
				v, i.start,
			),
		)
	}
	i.start.N--
	i.start.V = v
}

func (i *Interval) increaseRight(v int64) {
	// if i.start.V > v {
	if i.end.V > v {
		panic(
			fmt.Sprintf(
				"monotonicity violation: new start value (%d) must be bigger than existing one (%v)",
				v, i.start,
			),
		)
	}
	i.end.N++
	i.end.V = v
}

func (i *Interval) in(n int64) bool {
	return n >= i.start.N && n <= i.end.N
}
