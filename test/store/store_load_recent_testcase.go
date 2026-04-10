package store

import (
	"context"
	"testing"
	"time"

	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	asserterror "github.com/ymz-ncnk/assert/error"
)

type LoadRecentTestCase struct {
	Name   string
	Setup  LoadRecentSetup
	Params LoadRecentParams
	Want   LoadRecentWant
}

type LoadRecentSetup struct {
	NodeID     string
	Codec      evs.Codec
	Repository evs.Repository
}

type LoadRecentParams struct {
	Timeout   time.Duration
	Partition int
	Count     int
	ChptFn    evs.CheckpointCallback
	EventFn   evs.EventCallback
}

type LoadRecentWant struct {
	Error error
	Check func(t *testing.T, store evs.Store)
}

func RunLoadRecentTest(t *testing.T, tc LoadRecentTestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		ctx, cancel := makeLoadRecentCtx(t, tc.Params.Timeout)
		defer cancel()

		store := evs.New(tc.Setup.NodeID, tc.Setup.Codec, tc.Setup.Repository)
		err := store.LoadRecent(ctx, tc.Params.Partition, tc.Params.Count,
			tc.Params.ChptFn, tc.Params.EventFn)
		asserterror.EqualError(t, err, tc.Want.Error)
		if tc.Want.Check != nil {
			tc.Want.Check(t, store)
		}
	})
}

func makeLoadRecentCtx(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(t.Context(), timeout)
	}
	return t.Context(), func() {}
}

func nopEventFn(_ int64, _ bool, _ evnt.Event) error { return nil }
func nopChptFn(_ int64) error                        { return nil }
