package executor_test

import (
	"context"
	"testing"

	evnt "github.com/kapbit/kapbit-go/event"
	executor "github.com/kapbit/kapbit-go/executor"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/test"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

type EmitRejectedTestCase struct {
	Name   string
	Setup  EmitRejectedSetup
	Params EmitRejectedParams
	Want   EmitRejectedWant
}

type EmitRejectedSetup struct {
	Runtime *rtm.Runtime
	Tools   *evnt.Kit
}

type EmitRejectedParams struct {
	Ref    *rtm.WorkflowRef
	Reason string
}

type EmitRejectedWant struct {
	Error       error
	ShouldPanic bool
	Check       func(t *testing.T, ctx context.Context)
	mock        []*mok.Mock
}

func RunEmitRejectedTest(t *testing.T, tc EmitRejectedTestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		service := executor.NewEngine(tc.Setup.Runtime, tc.Setup.Tools,
			executor.OnFenced(cancel), test.NopLogger)

		defer func() {
			r := recover()
			if tc.Want.ShouldPanic {
				if r == nil {
					t.Errorf("expected panic, but code did not panic")
				}
				// Optional: check if the panic message contains "KAPBIT INTERNAL BUG:"
			} else if r != nil {
				t.Errorf("unexpected panic: %v", r)
			}
		}()

		err := service.EmitRejection(ctx, tc.Params.Ref, tc.Params.Reason)
		asserterror.EqualError(t, err, tc.Want.Error)

		if tc.Want.Check != nil {
			tc.Want.Check(t, ctx)
		}

		infomap := mok.CheckCalls(tc.Want.mock)
		asserterror.EqualDeep(t, infomap, mok.EmptyInfomap)
	})
}
