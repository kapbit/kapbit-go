package kapbit_test

import (
	"testing"

	test "github.com/kapbit/kapbit-go/test/kapbit"
)

func TestKapbit(t *testing.T) {
	for _, tc := range []test.TestCase{
		test.SuccessTestCase(),
		test.GateClosedTestCase(),
		test.CircuitBreakerErrorTestCase(),
		test.FactoryErrorTestCase(),
		test.ClosedTestCase(),
		test.FencedTestCase(),
		test.ExecutionErrorTestCase(),
		test.RegistrationErrorTestCase(),
		test.RegistrationCircuitBreakerErrorTestCase(),
	} {
		test.RunTest(t, tc)
	}
}
