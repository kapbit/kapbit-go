package persist_test

import (
	"testing"

	utils "github.com/kapbit/kapbit-go/test/handler"
)

func TestPersistHandler(t *testing.T) {
	for _, tc := range []utils.TestCase{
		utils.ActiveWriterTestCase(),
		utils.ResultTestCase(),
		utils.TwoCreateTestCase(),
		utils.DeadLetterTestCase(),
		utils.RejectedTestCase(),
		utils.CBOpenSystemEventTestCase(),
		utils.CBOpenWorkflowEventTestCase(),
		utils.CBOpenWorkflowCreatedTestCase(),
	} {
		utils.RunTest(t, tc)
	}
	utils.RunScenario(t, utils.ScenarioTestCase())
}
