package worker_test

import (
	"context"
	"testing"
	"time"

	rtm "github.com/kapbit/kapbit-go/runtime"
	swtch "github.com/kapbit/kapbit-go/support/retry/switch"
	wrk "github.com/kapbit/kapbit-go/worker"
	asserterror "github.com/ymz-ncnk/assert/error"
)

func DeadLetterTestCase() TestCase {
	service, gate, retryCh, ref, mock := testComponents()
	ref.MarkSaved()
	ref.IncrementRetriesCount(2)

	service.RegisterRetryQueue(
		func() <-chan *rtm.WorkflowRef {
			return retryCh
		},
	).RegisterEmitDeadLetter(
		func(ctx context.Context, ref *rtm.WorkflowRef) error {
			return nil
		},
	)

	return TestCase{
		Name: "Should emit DeadLetterEvent if too many retries",
		Config: TestConfig{
			Engine: service,
			Gate:    gate,
			Opts: []wrk.SetOption{
				wrk.WithMaxAttempts(2),
			},
		},
		Want: Want{
			Wait: 50 * time.Millisecond,
			Check: func(t *testing.T, ctx context.Context, worker *wrk.Worker) {
				asserterror.Equal(t, gate.Closed(), false)
				asserterror.Equal(t, worker.Level(), swtch.NormalLevel)
			},
		},
		mock: mock,
	}
}
