package checkpoint

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

type Result struct {
	Value Tuple
	Ok    bool
}

type Manager struct {
	logger     *slog.Logger
	partitions []*partitionState
	chptSize   int
	mu         sync.Mutex
}

func NewManager(logger *slog.Logger, chpts []Tuple, chptSize int) *Manager {
	partitions := make([]*partitionState, len(chpts))
	for i, chpt := range chpts {
		partitions[i] = &partitionState{
			chptProvider: NewProviderWith(chptSize, chpt),
		}
	}
	return &Manager{
		logger:     logger.With("comp", "chptman"),
		partitions: partitions,
		chptSize:   chptSize,
	}
}

// Add records a checkpoint (n, v) for the given partition.
// Grows the partitions slice if needed.
func (m *Manager) Add(partition int, tuple Tuple) (result Result,
	err error,
) {
	m.logger.Info("receive", "tuple", tuple)
	m.mu.Lock()
	if partition >= len(m.partitions) || partition < 0 {
		m.mu.Unlock()
		err = fmt.Errorf("checkpoint manager: partition index %d out of range", partition)
		return
	}
	state := m.partitions[partition]
	m.mu.Unlock()

	state.mu.Lock()
	defer state.mu.Unlock()
	result.Value, result.Ok = state.chptProvider.Add(tuple)
	return
}

func (m *Manager) Checkpoint(partition int) (chpt Tuple, err error) {
	m.mu.Lock()
	if partition >= len(m.partitions) || partition < 0 {
		m.mu.Unlock()
		err = fmt.Errorf("checkpoint manager: partition index %d out of range", partition)
		return
	}
	state := m.partitions[partition]
	m.mu.Unlock()

	state.mu.Lock()
	defer state.mu.Unlock()
	chpt = state.chptProvider.Checkpoint()
	return
}

func (m *Manager) PartitionCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.partitions)
}

func (m *Manager) StringPartition(partition int) string {
	var state *partitionState
	m.mu.Lock()
	state = m.partitions[partition]
	m.mu.Unlock()

	state.mu.Lock()
	defer state.mu.Unlock()
	return state.chptProvider.StringCompact()
}

func (m *Manager) String() string {
	m.mu.Lock()
	l := len(m.partitions)
	chptSize := m.chptSize
	m.mu.Unlock()

	var b strings.Builder
	// Format: CheckpointManager partitions=N chptSize=N [p0="..." p1="..."]
	fmt.Fprintf(&b, "checkpoint.Manager partitions=%d chptSize=%d [", l, chptSize)

	for i := range l {
		if i > 0 {
			b.WriteString(" ")
		}
		fmt.Fprintf(&b, "p%d=%q", i, m.StringPartition(i))
	}
	b.WriteString("]")
	return b.String()
}

type partitionState struct {
	mu           sync.Mutex
	chptProvider *Provider
}
