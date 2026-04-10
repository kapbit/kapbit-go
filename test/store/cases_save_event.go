package store

import (
	"context"
	"errors"
	"testing"

	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	"github.com/kapbit/kapbit-go/test/mock"
	asserterror "github.com/ymz-ncnk/assert/error"
)

func SaveEventEncodeErrorTestCase() SaveEventTestCase {
	name := "Should return an error when Codec.Encode fails"

	var (
		encodeErr = errors.New("encode error")
		event     = evnt.ActiveWriterEvent{NodeID: "node-1", PartitionIndex: 0}
	)

	return SaveEventTestCase{
		Name: name,
		Setup: SaveEventSetup{
			NodeID: "node-1",
			Codec: mock.NewCodecMock().RegisterEncode(
				func(evnt.Event) (evs.EventRecord, error) {
					return evs.EventRecord{}, encodeErr
				},
			),
			Repository: mock.NewRepositoryMock(),
		},
		Params: SaveEventParams{
			Event: event,
		},
		Want: SaveEventWant{
			Error: encodeErr,
		},
	}
}

func SaveEventRepositoryErrorTestCase() SaveEventTestCase {
	name := "Should return an error when Repository.SaveEvent fails"

	var (
		repoErr = errors.New("repository error")
		event   = evnt.ActiveWriterEvent{NodeID: "node-1", PartitionIndex: 0}
		record  = evs.EventRecord{Key: "some-key"}
	)

	return SaveEventTestCase{
		Name: name,
		Setup: SaveEventSetup{
			NodeID: "node-1",
			Codec: mock.NewCodecMock().RegisterEncode(
				func(evnt.Event) (evs.EventRecord, error) {
					return record, nil
				},
			),
			Repository: mock.NewRepositoryMock().RegisterSaveEvent(
				func(_ context.Context, _ int, _ evs.EventRecord) (int64, error) {
					return 0, repoErr
				},
			),
		},
		Params: SaveEventParams{
			Event: event,
		},
		Want: SaveEventWant{
			Error: repoErr,
		},
	}
}

func SaveEventSuccessTestCase() SaveEventTestCase {
	name := "Should encode the event and save it to the repository"

	var (
		wantOffset int64 = 42
		wantRecord       = evs.EventRecord{Key: "some-key"}
		event            = evnt.ActiveWriterEvent{NodeID: "node-1", PartitionIndex: 0}
		gotRecord  evs.EventRecord
	)

	return SaveEventTestCase{
		Name: name,
		Setup: SaveEventSetup{
			NodeID: "node-1",
			Codec: mock.NewCodecMock().RegisterEncode(
				func(evnt.Event) (evs.EventRecord, error) {
					return wantRecord, nil
				},
			),
			Repository: mock.NewRepositoryMock().RegisterSaveEvent(
				func(_ context.Context, _ int, r evs.EventRecord) (int64, error) {
					gotRecord = r
					return wantOffset, nil
				},
			),
		},
		Params: SaveEventParams{
			Event: event,
		},
		Want: SaveEventWant{
			Error: nil,
			Check: func(t *testing.T, _ evs.Store, offset int64) {
				asserterror.EqualDeep(t, gotRecord, wantRecord)
				asserterror.Equal(t, offset, wantOffset)
			},
		},
	}
}
