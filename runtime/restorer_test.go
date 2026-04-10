package runtime_test

import (
	"testing"

	utils "github.com/kapbit/kapbit-go/test/restorer"
)

func TestRestorer(t *testing.T) {
	for _, tc := range []utils.TestCase{
		utils.InvalidCheckpointTestCase(),
		utils.UnknownEventTypeTestCase(),
		utils.FactoryErrorTestCase(),
		utils.UnknownOutcomeType(),
		utils.ResultEventTestCase(),
		utils.DeadLetterEventTestCase(),
		utils.RejectedEventTestCase(),
		utils.SkipEventTestCase(),
		utils.DuplicateEventTestCase(),
		utils.OutcomeEventTestCase(),
		utils.BeforeCheckpointTestCase(),
	} {
		utils.RunTest(t, tc)
	}
}
