package runtime_test

import (
	"testing"

	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/support"
	"github.com/kapbit/kapbit-go/test"
	"github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
	asserterror "github.com/ymz-ncnk/assert/error"
)

func TestRuntime(t *testing.T) {
	t.Run("Should provide instances received during initialization",
		func(t *testing.T) {
			var (
				counter = &support.Counter{}
				window  = support.NewFIFOSet[wfl.ID](10)
				retryCh = make(chan *rtm.WorkflowRef, 5)
				r       = rtm.New(test.NopLogger, counter, window, retryCh)
			)
			asserterror.Equal(t, counter, r.Running())
			asserterror.Equal(t, window, r.IdempotencyWindow())
			asserterror.Equal(t, (<-chan *rtm.WorkflowRef)(retryCh), r.RetryQueue())
		})

	t.Run("Should enqueue a workflow reference when capacity is available",
		func(t *testing.T) {
			var (
				retryCh = make(chan *rtm.WorkflowRef, 1)
				ref     = newRef("wfl-1")
				r       = rtm.New(test.NopLogger, nil, nil, retryCh)
			)
			r.ScheduleRetry(ref)
			select {
			case received := <-retryCh:
				asserterror.Equal(t, received, ref)
			default:
				t.Fatal("Expected record to be in the channel, but it was empty")
			}
		})

	t.Run("Should panic when the retry queue is full",
		func(t *testing.T) {
			var (
				retryCh = make(chan *rtm.WorkflowRef, 1)
				ref1    = newRef("wfl-1")
				ref2    = newRef("wfl-2")
				r       = rtm.New(test.NopLogger, nil, nil, retryCh)
			)
			r.ScheduleRetry(ref1)
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Expected ScheduleRetry to panic, but it did not")
				}
			}()
			r.ScheduleRetry(ref2)
		})
}

func newRef(wid wfl.ID) *rtm.WorkflowRef {
	progress := wfl.NewProgress(wid, "test-type")
	mock := mock.NewWorkflowMock().RegisterProgress(func() *wfl.Progress {
		return progress
	}).RegisterNProgress(10, func() *wfl.Progress {
		return progress
	})
	return rtm.NewWorkflowRef(mock, 0)
}
