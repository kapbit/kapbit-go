package support

import "sync/atomic"

const (
	gateOpen   uint32 = 0
	gateClosed uint32 = 1
	gateSealed uint32 = 2
)

type IngressGate struct {
	state atomic.Uint32
}

func NewClosedIngressGate() *IngressGate {
	g := &IngressGate{}
	g.state.Store(gateClosed)
	return g
}

// Open returns true if the gate was closed and is now open.
// It returns false if the gate was already open or is sealed.
func (g *IngressGate) Open() bool {
	return g.state.CompareAndSwap(gateClosed, gateOpen)
}

// Close returns true if the gate was open and is now closed.
func (g *IngressGate) Close() bool {
	return g.state.CompareAndSwap(gateOpen, gateClosed)
}

func (g *IngressGate) Seal() {
	g.state.Store(gateSealed)
}

func (g *IngressGate) Closed() bool {
	return g.state.Load() != gateOpen
}
