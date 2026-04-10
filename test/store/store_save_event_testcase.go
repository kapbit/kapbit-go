package store

import (
	"context"
	"testing"
	"time"

	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	asserterror "github.com/ymz-ncnk/assert/error"
)

type SaveEventTestCase struct {
	Name   string
	Setup  SaveEventSetup
	Params SaveEventParams
	Want   SaveEventWant
}

type SaveEventSetup struct {
	NodeID     string
	Codec      evs.Codec
	Repository evs.Repository
}

type SaveEventParams struct {
	Timeout time.Duration
	Event   evnt.Event
}

type SaveEventWant struct {
	Error error
	Check func(t *testing.T, store evs.Store, offset int64)
}

func RunSaveEventTest(t *testing.T, tc SaveEventTestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		ctx, cancel := makeSaveEventCtx(t, tc.Params.Timeout)
		defer cancel()

		store := evs.New(tc.Setup.NodeID, tc.Setup.Codec, tc.Setup.Repository)
		offset, err := store.SaveEvent(ctx, 0, tc.Params.Event)
		asserterror.EqualError(t, err, tc.Want.Error)
		if tc.Want.Check != nil {
			tc.Want.Check(t, store, offset)
		}
	})
}

func makeSaveEventCtx(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(t.Context(), timeout)
	}
	return t.Context(), func() {}
}
