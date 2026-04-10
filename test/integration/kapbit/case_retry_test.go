package kapbit_test

import (
	"time"

	"github.com/kapbit/kapbit-go"
	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/kapbit/kapbit-go/executor"
	swtch "github.com/kapbit/kapbit-go/support/retry/switch"
	wrk "github.com/kapbit/kapbit-go/worker"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

var retryCase = TestCase{
	Name: "Should be able to retry Wofklow",
	Setup: Setup{
		NodeID:         "node_id",
		PartitionCount: 1,
		Codec:          codec,
		Defs: []wfl.Definition{
			wfl.MustDefinition(wfl.Type("workflow_type_exec"), resultBuilder,
				wfl.Step{Execute: &retryAction{retryCount: 2}},
			),
		},

		// 	{
		// 		Type: wfl.Type("workflow_type_exec"),
		// 		Steps: []wfl.Step{
		// 			{Execute: &retryAction{retryCount: 2}},
		// 		},
		// 		ResultBuilder: resultBuilder,
		// 	},
		// },
		Opts: []kapbit.SetOption{
			kapbit.WithWorker(
				wrk.WithMaxAttempts(5),
				wrk.WithPolicyOptions(
					swtch.WithNormalInterval(50*time.Millisecond),
					swtch.WithSlowInterval(200*time.Millisecond),
				),
			),
		},
	},
	Items: []Item{
		{
			Params: wfl.Params{
				ID:    "wfl-1",
				Type:  "workflow_type_exec",
				Input: wfl.Input("input_data"),
			},
			WantResult: nil,
			WantError: kapbit.NewKapbitError(executor.NewWillRetryLaterError(
				wfl.NewExecutionError(errConn)),
			),
		},
	},
	Want: Want{
		{
			Wait:  500 * time.Millisecond,
			Chpts: []int64{-1},
			Events: [][]evnt.Event{
				{
					evnt.ActiveWriterEvent{
						NodeID:         "node_id",
						PartitionIndex: 0,
					},
					evnt.WorkflowCreatedEvent{
						NodeID: "node_id",
						Key:    "wfl-1",
						Type:   "workflow_type_exec",
						Input:  "input_data",
					},
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-1",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("retry_step_ok"),
					},
					evnt.WorkflowResultEvent{
						NodeID: "node_id",
						Key:    "wfl-1",
						Result: "workflow_result_success",
					},
				},
			},
		},
	},
}
