package workflow_test

import (
	"testing"

	"github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

// TestProgress_RecordCompensationOutcome_NoOutcome tests that compensation
// cannot start when the last execution outcome is NoOutcome
func TestProgress_RecordCompensationOutcome_NoOutcome(t *testing.T) {
	progress := wfl.NewProgress("wfl-1", "my-wfl")

	// Add a successful execution outcome
	outcome1 := mock.NewOutcomeMock().RegisterFailure(func() bool { return false })
	err := progress.RecordExecutionOutcome(0, outcome1)
	if err != nil {
		t.Fatalf("unexpected error recording execution outcome: %v", err)
	}

	// Add NoOutcome as the last execution outcome
	err = progress.RecordExecutionOutcome(1, wfl.NoOutcome)
	if err != nil {
		t.Fatalf("unexpected error recording NoOutcome: %v", err)
	}

	// Try to start compensation - this should fail because last execution is NoOutcome
	compensationOutcome := mock.NewOutcomeMock().RegisterFailure(func() bool { return false })
	err = progress.RecordCompensationOutcome(0, compensationOutcome)

	// We expect a ProtocolViolationError with ErrInvalidRollback
	if err == nil {
		t.Fatal("expected error when trying to compensate after NoOutcome, got nil")
	}

	protocolErr, ok := err.(*wfl.ProtocolViolationError)
	if !ok {
		t.Fatalf("expected ProtocolViolationError, got %T: %v", err, err)
	}

	if protocolErr.Unwrap() != wfl.ErrInvalidRollback {
		t.Fatalf("expected ErrInvalidRollback, got: %v", protocolErr.Unwrap())
	}
}
