package runtime

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	chptsup "github.com/kapbit/kapbit-go/support/checkpoint"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

// PositionTracker tracks the next logical sequence number for each partition
// and the physical offset for each workflow.
//
// Once the workflow is saved, it gain a position with tracker.ReserveNext.
// Also PositionTracker holds data needed for the CheckpointManager.
type PositionTracker struct {
	nextLogicalSeq []int64
	pos            map[wfl.ID]Position
	mu             sync.RWMutex
}

func NewPositionTracker(chpts []chptsup.Tuple) *PositionTracker {
	tracker := PositionTracker{
		nextLogicalSeq: make([]int64, len(chpts)),
		pos:            make(map[wfl.ID]Position),
	}
	for i, chpt := range chpts {
		tracker.nextLogicalSeq[i] = chpt.N + 1
	}
	return &tracker
}

func (p *PositionTracker) ReserveNext(wid wfl.ID, partition int, offset int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, pst := p.pos[wid]; pst {
		return
	}
	p.pos[wid] = Position{
		LogicalSeq:     p.reserveNextLogicalSeq(partition),
		Partition:      partition,
		PhysicalOffset: offset,
	}
}

func (p *PositionTracker) Position(wid wfl.ID) (position Position, pst bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	position, pst = p.pos[wid]
	return
}

func (p *PositionTracker) Untrack(wid wfl.ID) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.pos, wid)
}

func (p *PositionTracker) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var sb strings.Builder

	// Prefix matches the actual location (runtime package)
	fmt.Fprintf(&sb, "runtime.PositionTracker partitions=%d [", len(p.nextLogicalSeq))

	for i, next := range p.nextLogicalSeq {
		if i > 0 {
			sb.WriteString(" ")
		}
		fmt.Fprintf(&sb, "p%d:NextN=%d", i, next)
	}
	sb.WriteString("]")

	if len(p.pos) > 0 {
		sb.WriteString(" tracked=[")

		// Collect and sort keys for deterministic "a != b" comparisons
		keys := make([]string, 0, len(p.pos))
		for k := range p.pos {
			keys = append(keys, string(k))
		}
		sort.Strings(keys)

		for i, wid := range keys {
			if i > 0 {
				sb.WriteString(" ")
			}
			pos := p.pos[wfl.ID(wid)]
			// Formatting the position inline to look like {N, V}
			fmt.Fprintf(&sb, "%s:%v", wid, pos)
		}
		sb.WriteString("]")
	} else {
		sb.WriteString(" tracked=none")
	}

	return sb.String()
}

// func (p *PositionTracker) String() string {
// 	p.mu.RLock()
// 	defer p.mu.RUnlock()
//
// 	// 1. Summarize Partition Sequences
// 	partitionInfo := ""
// 	for i, next := range p.nextLogicalSeq {
// 		partitionInfo += fmt.Sprintf(" P%d:NextN(%d)", i, next)
// 	}
//
// 	// 2. Format individual workflow positions
// 	// We'll use a string builder or simple concatenation for the list
// 	posInfo := ""
// 	if len(p.pos) > 0 {
// 		posInfo = "\n  Tracked Workflows:"
// 		for wid, pos := range p.pos {
// 			// Using the Position.String() we defined: {P:%d, N:%d, V:%d}
// 			posInfo += fmt.Sprintf("\n    - %s: %s", wid, pos)
// 		}
// 	} else {
// 		posInfo = " (No active workflows)"
// 	}
//
// 	return fmt.Sprintf("PositionTracker(Partitions:%d | Sequences:%s)%s",
// 		len(p.nextLogicalSeq),
// 		partitionInfo,
// 		posInfo,
// 	)
// }

// func (p *PositionTracker) shift(vals []int64) {
// 	if len(vals) != len(p.nextLogicalSeq) {
// 		panic(fmt.Sprintf("PositionTracker was initialized with %d partitions, but Shift was called with %v",
// 			len(p.nextLogicalSeq), len(vals)))
// 	}
// 	p.mu.Lock()
// 	defer p.mu.Unlock()
// 	for i, v := range vals {
// 		if i < len(p.nextLogicalSeq) {
// 			p.nextLogicalSeq[i] += v
// 		}
// 	}
// 	for wid, position := range p.pos {
// 		position.LogicalSeq += vals[position.Partition]
// 		p.pos[wid] = position
// 	}
// }

func (p *PositionTracker) reserveNextLogicalSeq(partition int) int64 {
	next := p.nextLogicalSeq[partition]
	p.nextLogicalSeq[partition]++
	return next
}
