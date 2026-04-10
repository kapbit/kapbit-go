package checkpoint

import (
	"fmt"
	"log"
	"strings"
)

// Provider is a checkpoint provider.
//
// After init, Next() always returns chpt + 1. During the work next increases
// continuously despite the chpt values. For example:
// - Add 1 1
// - Add 2 3
// - Add 3 8
// - Add 4 9
type Provider struct {
	chptSize int64
	chpt     Tuple
	first    *Interval
	last     *Interval
}

func NewProvider(chptSize int) *Provider {
	return NewProviderWith(chptSize, Tuple{N: -1, V: -1})
}

func NewProviderWith(chptSize int, chpt Tuple) *Provider {
	if chptSize < 1 {
		panic("chptSize must be >= 1")
	}
	if chpt.N < -1 {
		panic("chpt must be >= -1")
	}
	return &Provider{chptSize: int64(chptSize), chpt: chpt}
}

// Add records a checkpoint (n, v)
//
// TODO Add check for v value, it should always be >= than chpt, and previous
// v.
func (b *Provider) Add(tuple Tuple) (chpt Tuple, ok bool) {
	if tuple.N <= b.chpt.N {
		panic(
			fmt.Sprintf(
				"monotonicity violation: tuple.N %v is behind chpt %v", tuple, b.chpt,
			),
		)
	}

	if tuple.V <= b.chpt.V {
		panic(
			fmt.Sprintf(
				"monotonicity violation: tuple.V %v is behind chpt %v", tuple, b.chpt,
			),
		)
	}

	if b.first == nil {
		b.addFirst(tuple)
		return b.shift()
	}

	curr := b.first
	for curr != nil {
		switch {

		case b.inLeftGap(curr, tuple.N):
			b.addBetween(curr.prev, curr, tuple)
			// if chpt == 1, we should try to shift
			return b.shift()

		case curr.prevNum(tuple.N):
			return b.increaseLeft(curr, tuple.V)

		case curr.nextNum(tuple.N):
			if b.canMerge(curr) {
				return b.mergeRight(curr)
			}
			return b.increaseRight(curr, tuple.V)

		case b.inRightGap(curr, tuple.N):
			b.addBetween(curr, curr.next, tuple)
			return

		case curr.in(tuple.N):
			log.Printf("n %d in the interval %v", tuple.N, curr)
			return
		}

		curr = curr.next
	}
	return
}

func (b *Provider) Checkpoint() Tuple {
	return b.chpt
}

func (b *Provider) First() *Interval {
	return b.first
}

func (b *Provider) Last() *Interval {
	return b.last
}

// StringCompact returns a string representation of the provider.
//
// Format: chpt:{N,V} [{N,V}, {N,V}][...]
func (b *Provider) StringCompact() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "chpt:%v", b.Checkpoint())
	for curr := b.first; curr != nil; curr = curr.next {
		fmt.Fprintf(&sb, " [%v,%v]", curr.Start(), curr.End())
	}
	return sb.String()
}

func (b *Provider) String() string {
	var sb strings.Builder
	for curr := b.first; curr != nil; curr = curr.next {
		sb.WriteString(curr.String())
	}
	return sb.String()
}

func (b *Provider) addFirst(tuple Tuple) {
	intv := NewSingleInterval(tuple)
	b.first = intv
	b.last = intv
}

func (b *Provider) inLeftGap(inl *Interval, n int64) bool {
	// It's to the left of the current interval,
	// AND there's nothing before it OR it's to the right of the previous one.
	return inl.leftNum(n) && (inl.prev == nil || inl.prev.rightNum(n))
}

func (b *Provider) inRightGap(inl *Interval, n int64) bool {
	// It's to the right of the current interval,
	// AND there's nothing after it OR it's to the left of the next one.
	return inl.rightNum(n) && (inl.next == nil || inl.next.leftNum(n))
}

func (b *Provider) increaseLeft(inl *Interval, v int64) (chpt Tuple,
	ok bool,
) {
	inl.increaseLeft(v)
	if inl.prev == nil {
		return b.shift()
	}
	return
}

func (b *Provider) increaseRight(inl *Interval, v int64) (chpt Tuple,
	ok bool,
) {
	inl.increaseRight(v)
	if inl.prev == nil {
		return b.shift()
	}
	return
}

func (b *Provider) canMerge(inl *Interval) bool {
	next := inl.next
	return next != nil && next.prevNum(inl.end.N+1)
}

func (b *Provider) mergeRight(inl *Interval) (chpt Tuple, ok bool) {
	inl.mergeRight()
	if inl.next == nil {
		b.last = inl
	}
	if inl.prev == nil {
		return b.shift()
	}
	return
}

func (b *Provider) addBetween(left, right *Interval, tuple Tuple) {
	if left == nil && right == nil {
		panic("can't add between nil intervals")
	}
	if left == nil {
		b.prepend(right, tuple)
		return
	}
	if right == nil {
		b.append(left, tuple)
		return
	}
	if tuple.V <= left.End().V || tuple.V >= right.Start().V {
		panic(fmt.Sprintf(
			"monotonicity violation: tuple %v must be between %v and %v",
			tuple, left.End(), right.Start(),
		))
	}
	inl := NewSingleInterval(tuple)
	inl.prev = left
	inl.next = right
	left.next = inl
	right.prev = inl
}

func (b *Provider) prepend(right *Interval, tuple Tuple) {
	inl := NewSingleInterval(tuple)
	if tuple.V >= right.End().V {
		panic(fmt.Sprintf(
			"monotonicity violated: tuple %v must be smaller than following %v",
			tuple, right.End(),
		))
	}
	inl.next = right
	right.prev = inl
	b.first = inl
}

func (b *Provider) append(left *Interval, tuple Tuple) {
	inl := NewSingleInterval(tuple)
	if tuple.V <= left.Start().V {
		panic(fmt.Sprintf(
			"monotonicity violated: tuple %v must be bigger than previous %v",
			tuple, left.Start(),
		))
	}
	inl.prev = left
	left.next = inl
	b.last = inl
}

func (b *Provider) shift() (chpt Tuple, ok bool) {
	l := b.first.Len()

	if b.first.prevNum(b.chpt.N) && l >= b.chptSize {
		// b.chpt = Tuple{b.first.End(), b.first.Value()}
		b.chpt = b.first.End()
		if b.first.next == nil { // b.first == b.last
			b.first = nil
			b.last = nil
		} else {
			b.first = b.first.next
			b.first.prev = nil
		}
		return b.chpt, true
	}

	return
}
