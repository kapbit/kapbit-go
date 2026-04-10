package handler

import (
	"testing"

	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/kapbit/kapbit-go/event/persist"
	evs "github.com/kapbit/kapbit-go/event/store"
	"github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/support"
	chptsup "github.com/kapbit/kapbit-go/support/checkpoint"
	"github.com/kapbit/kapbit-go/test"
	asserterror "github.com/ymz-ncnk/assert/error"
)

type TestCase struct {
	Name           string
	Setup          TestSetup
	Events         []evnt.Event
	Scenario       *Scenario
	HandlerOptions []persist.SetOption
	Want           Want
}

type TestSetup struct {
	Store    evs.EventStore
	Chptman  *chptsup.Manager
	Tracker  *runtime.PositionTracker
	Resolver support.PartitionResolver
}

type Want struct {
	Errors []error
	Check  func(t *testing.T, config TestSetup)
}

func RunTest(t *testing.T, tc TestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		if len(tc.Want.Errors) != len(tc.Events) {
			t.Fatalf("Want.Errors length (%d) must match Events length (%d)",
				len(tc.Want.Errors), len(tc.Events))
		}

		opts := append([]persist.SetOption{persist.WithLogger(test.NopLogger)}, tc.HandlerOptions...)
		handler, err := persist.NewPersistHandler(tc.Setup.Store, tc.Setup.Chptman,
			tc.Setup.Tracker, opts...)
		if err != nil {
			t.Fatal(err)
		}

		for i, event := range tc.Events {
			err := handler.Handle(t.Context(), event)
			asserterror.EqualError(t, err, tc.Want.Errors[i])
		}

		if tc.Want.Check != nil {
			tc.Want.Check(t, tc.Setup)
		}
	})
}
