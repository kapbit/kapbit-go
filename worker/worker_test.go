package worker_test

import (
	"testing"

	utils "github.com/kapbit/kapbit-go/test/worker"
)

func TestWorker(t *testing.T) {
	for _, tc := range []utils.TestCase{
		utils.SavedRefTestCase(),
		utils.UnsavedRefTestCase(),
		utils.NonCircuitBreakerErrorTestCase(),
		utils.CircuitBreakerErrorTestCase(),
		utils.DeadLetterTestCase(),
		utils.StopTestCase(),
	} {
		utils.RunTest(t, tc)
	}
}
