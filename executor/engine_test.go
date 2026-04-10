package executor_test

import (
	"testing"

	test "github.com/kapbit/kapbit-go/test/executor"
)

func TestEmitWorfklowCreated(t *testing.T) {
	for _, tc := range []test.EmitCreatedTestCase{
		test.EmitCreatedTestCases.ShouldSuccess(),
		test.EmitCreatedTestCases.EncodeError(),
		test.EmitCreatedTestCases.NoActiveWriter(),
		test.EmitCreatedTestCases.PersistenceRejection(),
		test.EmitCreatedTestCases.UnexpectedErrorKind(),
		test.EmitCreatedTestCases.UnexpectedError(),
	} {
		test.RunEmitCreatedTest(t, tc)
	}
}

func TestEmitDeadletter(t *testing.T) {
	for _, tc := range []test.EmitDeadLetterTestCase{
		test.EmitDeadLetterTestCases.ShouldSuccess(),
		test.EmitDeadLetterTestCases.EncodeError(),
		test.EmitDeadLetterTestCases.NoActiveWriter(),
		test.EmitDeadLetterTestCases.Rejection(),
		test.EmitDeadLetterTestCases.UnexpectedKind(),
		test.EmitDeadLetterTestCases.UnexpectedError(),
	} {
		test.RunEmitDeadLetterTest(t, tc)
	}
}

func TestEmitRejected(t *testing.T) {
	for _, tc := range []test.EmitRejectedTestCase{
		test.EmitRejectedTestCases.ShouldSuccess(),
		test.EmitRejectedTestCases.EncodeError(),
		test.EmitRejectedTestCases.NoActiveWriter(),
		test.EmitRejectedTestCases.Rejection(),
		test.EmitRejectedTestCases.UnexpectedKind(),
		test.EmitRejectedTestCases.UnexpectedError(),
	} {
		test.RunEmitRejectedTest(t, tc)
	}
}

func TestExecuteWorkflow(t *testing.T) {
	for _, tc := range []test.ExecuteTestCase{
		test.ExecuteTestCases.ShouldSuccess(),
		test.ExecuteTestCases.ProtocolViolationError(),
		test.ExecuteTestCases.EncodeError(),
		test.ExecuteTestCases.NoActiveWriterError(),
		test.ExecuteTestCases.RejectionError(),
		test.ExecuteTestCases.UnexpectedKind(),
		test.ExecuteTestCases.UnexpectedError(),
	} {
		test.RunExecuteWorkflowTest(t, tc)
	}
}

func TestRegisterWorkflow(t *testing.T) {
	testCases := []test.RegisterTestCase{
		test.RegisterTestCases.ShouldSuccess(),
		test.RegisterTestCases.NoAvailaleSlots(),
		test.RegisterTestCases.IdempotencyViolation(),
		test.RegisterTestCases.EncodeError(),
		test.RegisterTestCases.NoActiveWriter(),
		test.RegisterTestCases.RejectionError(),
		test.RegisterTestCases.UnexpectedKind(),
		test.RegisterTestCases.UnexpectedError(),
	}
	for _, tc := range testCases {
		test.RunRegisterWorkflowTest(t, tc)
	}
}
