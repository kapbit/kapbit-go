package executor_test

import (
	"context"
	"testing"

	evnt "github.com/kapbit/kapbit-go/event"
	executor "github.com/kapbit/kapbit-go/executor"
	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/test"
	"github.com/kapbit/kapbit-go/test/mock"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

type RegisterTestCase struct {
	Name   string
	Setup  Setup
	Params Params
	Want   RegisterWant
}

type Setup struct {
	Runtime *rtm.Runtime
	Tools   *evnt.Kit
}

type Params struct {
	Workflow mock.WorkflowMock
}

type RegisterWant struct {
	Ref         *rtm.WorkflowRef
	Error       error
	Check       func(t *testing.T, ctx context.Context)
	ShouldPanic bool
	mock        []*mok.Mock
}

func RunRegisterWorkflowTest(t *testing.T, tc RegisterTestCase) {
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

		ref, err := service.RegisterWorkflow(ctx, tc.Params.Workflow)
		asserterror.EqualError(t, err, tc.Want.Error)
		asserterror.EqualDeep(t, ref, tc.Want.Ref)

		tc.Want.Check(t, ctx)
	})
}
