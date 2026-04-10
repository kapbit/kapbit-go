package impl_test

import (
	"testing"

	test "github.com/kapbit/kapbit-go/test/workflow"
)

func TestWorkflow_Run(t *testing.T) {
	for _, tc := range []test.TestCase{
		test.ExecTestCase(t),
		test.ExecContinueTestCase(t),
		test.ExecPartialTestCase(t),
		test.ExecErrorCase(),
		test.ExecEmitErrorTestCase(),
		test.ExecNoOutcomeTestCase(t),
		test.ExecNoOutcomeStateTestCase(t),

		test.CompTestCase(t),
		test.CompContinueTestCase(t),
		test.CompPartialTestCase(t),
		test.CompErrorTestCase(t),
		test.CompEmitErrorTestCase(t),
		test.CompStartTestCase(t),
		test.CompNoOutcomeStateTestCase(t),

		test.FinErrorTestCase(),
		test.FinNoResultTestCase(t),

		test.ZeroStepsTestCase(t),
	} {
		test.RunTestCase(t, tc)
	}
}
