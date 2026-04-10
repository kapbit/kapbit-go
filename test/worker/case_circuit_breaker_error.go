package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/kapbit/kapbit-go/codes"
	rtm "github.com/kapbit/kapbit-go/runtime"
	swtch "github.com/kapbit/kapbit-go/support/retry/switch"
	wrk "github.com/kapbit/kapbit-go/worker"
	wfl "github.com/kapbit/kapbit-go/workflow"
	asserterror "github.com/ymz-ncnk/assert/error"
)

func CircuitBreakerErrorTestCase() TestCase {
	service, gate, retryCh, ref, mock := testComponents()
	ref.MarkSaved()

	service.RegisterRetryQueue(
		func() <-chan *rtm.WorkflowRef {
			return retryCh
		},
	).RegisterExecWorkflow(
		func(ctx context.Context, ref *rtm.WorkflowRef) (result wfl.Result, err error) {
			return nil, codes.NewCircuitBreakerOpenError("my-service")
		},
	)
	gate.Close()

	return TestCase{
		Name: "Should mantain slow execution level on cirbuit-breaker error",
		Config: TestConfig{
			Engine: service,
			Gate:    gate,
		},
		Want: Want{
			Wait: 50 * time.Millisecond,
			Check: func(t *testing.T, ctx context.Context, worker *wrk.Worker) {
				asserterror.Equal(t, gate.Closed(), true)
				asserterror.Equal(t, worker.Level(), swtch.SlowLevel)
			},
		},
		mock: mock,
	}
}
