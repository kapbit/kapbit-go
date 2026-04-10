package mem

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kapbit/kapbit-go/codes"
	evs "github.com/kapbit/kapbit-go/event/store"
)

type Repository struct {
	partitions       map[int][]StoredRecord
	payloadSizeLimit int
	chpts            map[int]int64
	mu               sync.Mutex
}

func New(partitionCount, payloadSizeLimit int) *Repository {
	repo := &Repository{
		partitions:       make(map[int][]StoredRecord, partitionCount),
		payloadSizeLimit: payloadSizeLimit,
		chpts:            make(map[int]int64, partitionCount),
	}
	for i := range partitionCount {
		repo.partitions[i] = []StoredRecord{}
		repo.chpts[i] = -1
	}
	return repo
}

func (s *Repository) SaveEvent(ctx context.Context, partition int,
	record evs.EventRecord,
) (offset int64, err error) {
	payloadSize := len(record.Payload)
	if s.payloadSizeLimit > 0 && payloadSize > s.payloadSizeLimit {
		err = codes.NewPersistenceError(NewTooLargePayloadError(payloadSize),
			codes.PersistenceKindRejection)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if partition < 0 || partition >= len(s.partitions) {
		err = NewInvalidPartitionError(partition, len(s.partitions))
		return
	}
	offset = int64(len(s.partitions[partition]))
	s.partitions[partition] = append(s.partitions[partition],
		s.makeStoredRecord(offset, record))
	return
}

func (s *Repository) SaveCheckpoint(ctx context.Context, partition int, chpt int64) (
	err error,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if partition < 0 || partition >= len(s.partitions) {
		err = NewInvalidPartitionError(partition, len(s.partitions))
		return
	}
	s.chpts[partition] = chpt
	return
}

// LoadRecent loads recent events from storage
//
// To load all "full" events from the start: startCondition == nil && chpt == -1.
func (s *Repository) LoadRecent(ctx context.Context, partition int,
	startCondition evs.StartCondition,
	chptFn evs.CheckpointCallback,
	eventFn evs.RecentEventRecordCallback,
) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if partition < 0 || partition >= len(s.partitions) {
		err = NewInvalidPartitionError(partition, len(s.partitions))
		return
	}

	var (
		chpt         = s.feedChpt(partition, chptFn)
		offset int64 = 0
		log          = s.partitions[partition]
		ok     bool
	)
	if startCondition != nil {
		for i := len(log) - 1; i >= 0; i-- {
			ok, err = startCondition(log[i].Record)
			if err != nil {
				return
			}
			if ok {
				offset = int64(i)
				break
			}
		}
	}
	if offset > chpt+1 {
		offset = chpt + 1
	}
	return s.feedEvents(ctx, partition, offset, chpt, eventFn, true)
}

func (s *Repository) ScanPartition(ctx context.Context, partition int,
	chptFn evs.CheckpointCallback,
	eventFn evs.RecentEventRecordCallback,
) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if partition < 0 || partition >= len(s.partitions) {
		err = NewInvalidPartitionError(partition, len(s.partitions))
		return
	}
	var (
		chpt         = s.feedChpt(partition, chptFn)
		offset int64 = 0
	)
	return s.feedEvents(ctx, partition, offset, chpt, eventFn, false)
}

func (s *Repository) PartitionCount() int {
	return len(s.partitions)
}

func (s *Repository) WatchAfter(ctx context.Context, partition int, action func() error,
	callback evs.EventRecordCallback,
) (int64, error) {
	err := action()
	if err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	log := s.partitions[partition]
	for i := len(log) - 1; i >= 0; i-- {
		ok, err := callback(ctx, log[i].Record)
		if err != nil {
			return 0, err
		}
		if ok {
			return int64(i), nil
		}
	}
	return 0, codes.NewPersistenceError(errors.New("event not found"), codes.PersistenceKindOther)
}

func (s *Repository) Close(ctx context.Context) error {
	return nil
}

func (s *Repository) feedChpt(partition int, chptFn evs.CheckpointCallback) (
	chpt int64,
) {
	chpt, pst := s.chpts[partition]
	if !pst {
		chpt = -1
	}
	chptFn(chpt)
	return
}

func (s *Repository) feedEvents(ctx context.Context, partition int,
	offset, chpt int64,
	eventFn evs.RecentEventRecordCallback,
	withLightRecords bool,
) error {
	log := s.partitions[partition]
	for i := int(offset); i < len(log); i++ {
		var (
			beforeChpt = int64(i) <= chpt
			record     evs.EventRecord
		)
		// log[i].Record.Key != "" - for non workflow related events.
		if beforeChpt && withLightRecords && log[i].Record.Key != "" {
			record.Key = log[i].Record.Key
			record.Meta = log[i].Record.Meta
		} else {
			record = log[i].Record
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := eventFn(int64(i), beforeChpt, record); err != nil {
			return err
		}
	}
	return nil
}

func (s *Repository) makeStoredRecord(offset int64, record evs.EventRecord) StoredRecord {
	payloadCopy := make([]byte, len(record.Payload))
	copy(payloadCopy, record.Payload)
	record.Payload = payloadCopy

	return StoredRecord{
		Offset:    offset,
		Record:    record,
		CreatedAt: time.Now(),
	}
}
