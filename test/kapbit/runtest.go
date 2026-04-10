package kapbit

import (
	"context"
	"testing"

	"github.com/kapbit/kapbit-go"
	asserterror "github.com/ymz-ncnk/assert/error"
	assertfatal "github.com/ymz-ncnk/assert/fatal"
	"github.com/ymz-ncnk/mok"
)

func RunTest(t *testing.T, tc TestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		ctx, cancelCause := context.WithCancelCause(context.Background())
		kbit, err := kapbit.NewWithTools(ctx, cancelCause, tc.Setup.Tools, tc.Setup.Options)
		assertfatal.Equal(t, err, nil)
		defer kbit.Close()

		if tc.Setup.Prepare != nil {
			tc.Setup.Prepare(kbit)
		}

		result, err := kbit.ExecWorkflow(tc.Params.ID, tc.Params.Type, tc.Params.Input)
		asserterror.EqualError(t, err, tc.Want.Error)
		asserterror.EqualDeep(t, result, tc.Want.Result)

		if tc.Want.Check != nil {
			tc.Want.Check(t, kbit)
		}

		infomap := mok.CheckCalls(tc.Setup.mock)
		asserterror.EqualDeep(t, infomap, mok.EmptyInfomap)
	})
}
