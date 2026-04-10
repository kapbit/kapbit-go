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

var deadLetterCase = TestCase{
	Name: "Should emit a DeadLetter event when retries limit exhaused",
	Setup: Setup{
		NodeID:         "node_id",
		PartitionCount: 1,
		Codec:          codec,
		Defs: []wfl.Definition{
			wfl.MustDefinition(wfl.Type("workflow_type_exec"), resultBuilder,
				wfl.Step{Execute: successAction},
				wfl.Step{Execute: failAction},
			),

			wfl.MustDefinition(wfl.Type("workflow_type_dead_letter"), resultBuilder,
				wfl.Step{Execute: successAction},
				wfl.Step{Execute: &retryAction{retryCount: 4}},
			),
		},
		// Defs: []wfl.Definition{
		// 	{
		// 		Type: wfl.Type("workflow_type_exec"),
		// 		Steps: []wfl.Step{
		// 			{Execute: successAction},
		// 			{Execute: failAction},
		// 		},
		// 		ResultBuilder: resultBuilder,
		// 	},
		// 	{
		// 		Type: wfl.Type("workflow_type_dead_letter"),
		// 		Steps: []wfl.Step{
		// 			{Execute: successAction},
		// 			{Execute: &retryAction{retryCount: 4}},
		// 		},
		// 		ResultBuilder: resultBuilder,
		// 	},
		// },
		Opts: []kapbit.SetOption{
			kapbit.WithWorker(
				wrk.WithMaxAttempts(3),
				wrk.WithPolicyOptions(
					swtch.WithNormalInterval(100*time.Millisecond),
				),
			),
		},
	},
	Items: []Item{
		{
			// First Workflow is used to set not 0 DeadLetterEvent.WorkflowOffset.
			Params: wfl.Params{
				ID:    "wfl-1",
				Type:  "workflow_type_exec",
				Input: wfl.Input("input_data"),
			},
			WantResult: wfl.Result("workflow_result_success"),
			WantError:  nil,
		},
		{
			Params: wfl.Params{
				ID:    "wfl-2",
				Type:  "workflow_type_dead_letter",
				Input: wfl.Input("input_data"),
			},
			WantResult: nil,
			WantError: kapbit.NewKapbitError(
				executor.NewWillRetryLaterError(
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
						Outcome:     wfl.Success("step_ok"),
					},
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-1",
						OutcomeSeq:  1,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Fail("step_fail"),
					},
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-1",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeCompensation,
						Outcome:     wfl.NoOutcome,
					},
					evnt.WorkflowResultEvent{
						NodeID: "node_id",
						Key:    "wfl-1",
						Result: "workflow_result_success",
					},
					evnt.WorkflowCreatedEvent{
						NodeID: "node_id",
						Key:    "wfl-2",
						Type:   "workflow_type_dead_letter",
						Input:  "input_data",
					},
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-2",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
					},
					&evnt.DeadLetterEvent{
						NodeID:         "node_id",
						Key:            "wfl-2",
						WorkflowOffset: 6,
					},
				},
			},
		},
	},
}
