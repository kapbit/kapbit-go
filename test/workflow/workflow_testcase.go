package workflow

import (
	"log/slog"
	"testing"

	evnt "github.com/kapbit/kapbit-go/event"
	wfl "github.com/kapbit/kapbit-go/workflow"
	"github.com/kapbit/kapbit-go/workflow/impl"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

type TestCase struct {
	Name  string
	Setup Setup
	Want  Want
	mock  []*mok.Mock
}

type Setup struct {
	NodeID   string
	Spec     wfl.Spec
	Progress *wfl.Progress
	Tools    *evnt.Kit
	Logger   *slog.Logger
}

type Want struct {
	State  wfl.State
	Result wfl.Result
	Error  error
}

func RunTestCase(t *testing.T, tc TestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		workflow := impl.NewWorkflow(tc.Setup.NodeID, tc.Setup.Spec,
			tc.Setup.Progress, tc.Setup.Tools, tc.Setup.Logger)
		result, err := workflow.Run(t.Context())
		asserterror.EqualError(t, err, tc.Want.Error)
		asserterror.EqualDeep(t, result, tc.Want.Result)
		asserterror.Equal(t, workflow.State(), tc.Want.State)

		infomap := mok.CheckCalls(tc.mock)
		asserterror.EqualDeep(t, infomap, mok.EmptyInfomap)
	})
}
