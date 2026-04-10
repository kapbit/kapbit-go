package restorer

import (
	"testing"

	evs "github.com/kapbit/kapbit-go/event/store"
	rtm "github.com/kapbit/kapbit-go/runtime"
	wfl "github.com/kapbit/kapbit-go/workflow"
	asserterror "github.com/ymz-ncnk/assert/error"
)

type TestCase struct {
	Name   string
	Config TestConfig
	Want   Want
}

type TestConfig struct {
	NodeID  string
	Factory wfl.Factory
	Store   evs.EventStore
	Opts    []rtm.SetOption
}

type Want struct {
	CounterValue         int
	IdempotencyWindowStr string
	Refs                 []*rtm.WorkflowRef
	ChptmanStr           string
	TrackerStr           string
	Error                error
}

func RunTest(t *testing.T, tc TestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		restorer, err := rtm.NewRestorer(tc.Config.NodeID, tc.Config.Factory, tc.Config.Store,
			tc.Config.Opts...)
		if err != nil {
			asserterror.EqualError(t, err, tc.Want.Error)
			return
		}
		numPartitions := tc.Config.Store.PartitionCount()
		aweOffsets := make([]int64, numPartitions)
		runtime, chptman, tracker, err := restorer.Restore(t.Context(), aweOffsets)
		asserterror.EqualError(t, err, tc.Want.Error)

		if tc.Want.Error != nil {
			return
		}

		if chptman != nil {
			asserterror.Equal(t, chptman.String(), tc.Want.ChptmanStr)
		} else if tc.Want.ChptmanStr != "" {
			t.Errorf("chptman is nil, but expected %q", tc.Want.ChptmanStr)
		}
		if tracker != nil {
			asserterror.Equal(t, tracker.String(), tc.Want.TrackerStr)
		} else if tc.Want.TrackerStr != "" {
			t.Errorf("tracker is nil, but expected %q", tc.Want.TrackerStr)
		}

		verifyRuntime(t, runtime, tc.Want)
	})
}

func verifyRuntime(t *testing.T, rt *rtm.Runtime, want Want) {
	asserterror.Equal(t, rt.Running().Count(), want.CounterValue)
	asserterror.Equal(t, rt.IdempotencyWindow().String(), want.IdempotencyWindowStr)
	for _, wantRef := range want.Refs {
		select {
		case ref := <-rt.RetryQueue():
			asserterror.Equal(t, ref.Equal(wantRef), true)
		default:
			t.Errorf("expected %v to be in retry queue", wantRef)
		}
	}
}
