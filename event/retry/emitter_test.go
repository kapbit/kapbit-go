package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kapbit/kapbit-go/codes"
	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/kapbit/kapbit-go/event/retry"
	"github.com/kapbit/kapbit-go/test"
	"github.com/kapbit/kapbit-go/test/mock"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

func TestRetryEmitter_Emit(t *testing.T) {
	testCases := []struct {
		name        string
		failCount   int   // Number of transient errors before success
		terminalErr error // If provided, Handle returns this and stops immediately
		wantDelays  int
	}{
		{
			name:       "Success on first attempt",
			failCount:  0,
			wantDelays: 0,
		},
		{
			name:       "Recover after transient failures",
			failCount:  3, // Fails 3 times, succeeds on 4th
			wantDelays: 3,
		},
		{
			name:       "Recover after crossing warn threshold",
			failCount:  6, // Fails 6 times, succeeds on 7th
			wantDelays: 6,
		},
		{
			name:        "Stop immediately on Encode Error",
			terminalErr: codes.NewEncodeError(nil, "json", errors.New("malformed")),
			wantDelays:  0,
		},
		{
			name: "Stop immediately on Persistence Rejection",
			terminalErr: codes.NewPersistenceError(errors.New("bad request"),
				codes.PersistenceKindRejection),
			wantDelays: 0,
		},
		{
			name: "Stop immediately on NoActiveWriter",
			terminalErr: codes.NewPersistenceError(errors.New("node fenced"),
				codes.PersistenceKindFenced),
			wantDelays: 0,
		},
		{
			name:       "Invoke OnTransientError hook on failure",
			failCount:  1,
			wantDelays: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				handler          = mock.NewEventHandlerMock()
				policy           = mock.NewPolicyMock()
				hookCalledWith   error
				onTransientError = func(err error) {
					hookCalledWith = err
				}
			)
			emitter, err := retry.NewRetryEmitter(handler,
				retry.WithPolicy(policy),
				retry.WithLogger(test.NopLogger),
				retry.WithOnTransientError(onTransientError),
			)
			if err != nil {
				t.Fatalf("failed to create emitter: %v", err)
			}

			var event evnt.Event = "test-event"

			// 1. Setup Policy Mock expectations
			if tc.wantDelays > 0 {
				policy.RegisterNNextDelay(tc.wantDelays, func(state int) time.Duration {
					// State represents the current value of 'i'
					return 0 // Return 0 to prevent test delay
				})
			}

			// 2. Setup Handler Mock behavior
			if tc.terminalErr != nil {
				handler.RegisterHandle(
					func(ctx context.Context, e evnt.Event) error {
						return tc.terminalErr
					})
			} else {
				// Transient failures
				if tc.failCount > 0 {
					handler.RegisterNHandle(tc.failCount,
						func(ctx context.Context, e evnt.Event) error {
							return errors.New("transient infrastructure error")
						})
				}
				// The eventual success
				handler.RegisterHandle(func(ctx context.Context, e evnt.Event) error {
					return nil
				})
			}

			// Execute
			err = emitter.Emit(context.Background(), event)

			// Assertions
			if tc.terminalErr != nil {
				asserterror.Equal(t, tc.terminalErr, err)
			} else {
				asserterror.EqualError(t, err, nil)
			}

			if tc.failCount > 0 {
				if hookCalledWith == nil {
					t.Errorf("expected OnTransientError hook to be called")
				}
			} else if tc.terminalErr != nil {
				// For terminal errors, hook should NOT be called
				if hookCalledWith != nil {
					t.Errorf("expected OnTransientError hook NOT to be called for terminal error")
				}
			}

			infomap := mok.CheckCalls([]*mok.Mock{handler.Mock, policy.Mock})
			asserterror.EqualDeep(t, infomap, mok.EmptyInfomap)
		})
	}
}
