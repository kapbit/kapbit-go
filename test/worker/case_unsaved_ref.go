package worker_test

import (
	"context"
	"testing"
	"time"

	rtm "github.com/kapbit/kapbit-go/runtime"
	swtch "github.com/kapbit/kapbit-go/support/retry/switch"
	wrk "github.com/kapbit/kapbit-go/worker"
	wfl "github.com/kapbit/kapbit-go/workflow"
	asserterror "github.com/ymz-ncnk/assert/error"
)

func UnsavedRefTestCase() TestCase {
	service, gate, retryCh, ref, mock := testComponents()

	service.RegisterRetryQueue(
		func() <-chan *rtm.WorkflowRef {
			return retryCh
		},
	).RegisterEmitWorkflowCreated(
		func(ctx context.Context, ref *rtm.WorkflowRef) (err error) {
			return nil
		},
	).RegisterExecWorkflow(
		func(ctx context.Context, ref *rtm.WorkflowRef) (result wfl.Result, err error) {
			return wfl.Result("success"), nil
		},
	)

	return TestCase{
		Name: "Should emit WorkflowCreatedEvent if ref is unsaved",
		Config: TestConfig{
			Engine: service,
			Gate:    gate,
		},
		Want: Want{
			Wait: 50 * time.Millisecond,
			Check: func(t *testing.T, ctx context.Context, worker *wrk.Worker) {
				asserterror.Equal(t, worker.Level(), swtch.NormalLevel)
				asserterror.Equal(t, ref.Saved(), true)
			},
		},
		mock: mock,
	}
}
