package kapbit_test

import (
	"testing"
)

func TestIntegration(t *testing.T) {
	cases := []TestCase{
		execCase,
		compCase,
		compFailedCase,
		retryCase,
		circuitBrakerOpenCase,
		rejectionCase,
		deadLetterCase,
		recoverCase,
		partitionsCase,
	}

	for _, tc := range cases {
		runTestCase(t, tc)
	}
}
