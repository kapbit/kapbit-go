package executor_test

import (
	"context"
	"testing"

	evnt "github.com/kapbit/kapbit-go/event"
	executor "github.com/kapbit/kapbit-go/executor"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/test"
	wfl "github.com/kapbit/kapbit-go/workflow"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

type ExecuteTestCase struct {
	Name   string
	Setup  ExecuteSetup
	Params ExecuteParams
	Want   ExecuteWant
}

type ExecuteSetup struct {
	Runtime *rtm.Runtime
	Tools   *evnt.Kit
}

type ExecuteParams struct {
	Ref *rtm.WorkflowRef
}

type ExecuteWant struct {
	Result      wfl.Result
	Error       error
	mock        []*mok.Mock
	Check       func(t *testing.T, ctx context.Context)
	ShouldPanic bool
}

func RunExecuteWorkflowTest(t *testing.T, tc ExecuteTestCase) {
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

		result, err := service.ExecWorkflow(ctx, tc.Params.Ref)
		asserterror.EqualError(t, err, tc.Want.Error)
		asserterror.Equal(t, result, tc.Want.Result)

		if tc.Want.Check != nil {
			tc.Want.Check(t, ctx)
		}
	})
}
