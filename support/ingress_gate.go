package support

import "sync/atomic"

const (
	gateOpen   uint32 = 0
	gateClosed uint32 = 1
	gateSealed uint32 = 2
)

type EntryGate struct {
	state atomic.Uint32
}

func NewClosedEntryGate() *EntryGate {
	g := &EntryGate{}
	g.state.Store(gateClosed)
	return g
}

// Open returns true if the gate was closed and is now open.
// It returns false if the gate was already open or is sealed.
func (g *EntryGate) Open() bool {
	return g.state.CompareAndSwap(gateClosed, gateOpen)
}

// Close returns true if the gate was open and is now closed.
func (g *EntryGate) Close() bool {
	return g.state.CompareAndSwap(gateOpen, gateClosed)
}

func (g *EntryGate) Seal() {
	g.state.Store(gateSealed)
}

func (g *EntryGate) Closed() bool {
	return g.state.Load() != gateOpen
}
