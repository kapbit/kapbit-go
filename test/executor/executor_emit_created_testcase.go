package executor_test

import (
	"context"
	"testing"
	"time"

	evnt "github.com/kapbit/kapbit-go/event"
	executor "github.com/kapbit/kapbit-go/executor"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/test"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

type EmitCreatedTestCase struct {
	Name   string
	Setup  EmitCreatedSetup
	Params EmitCreatedParams
	Want   EmitCreatedWant
}

type EmitCreatedSetup struct {
	Runtime *rtm.Runtime
	Tools   *evnt.Kit
}

type EmitCreatedParams struct {
	Timeout time.Duration
	Ref     *rtm.WorkflowRef
}

type EmitCreatedWant struct {
	Error       error
	ShouldPanic bool
	Check       func(t *testing.T, ctx context.Context)
	mock        []*mok.Mock
}

func RunEmitCreatedTest(t *testing.T, tc EmitCreatedTestCase) {
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

		err := service.EmitWorkflowCreated(ctx, tc.Params.Ref)
		asserterror.EqualError(t, err, tc.Want.Error)

		if tc.Want.Check != nil {
			tc.Want.Check(t, ctx)
		}

		infomap := mok.CheckCalls(tc.Want.mock)
		asserterror.EqualDeep(t, infomap, mok.EmptyInfomap)
	})
}
