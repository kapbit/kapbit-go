package kapbit_test

import (
	"github.com/kapbit/kapbit-go"
	"github.com/kapbit/kapbit-go/codes"
	evnt "github.com/kapbit/kapbit-go/event"
	mem "github.com/kapbit/kapbit-go/repository/mem"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

var rejectionCase = TestCase{
	Name: "Should emit a Rejected event upon storage rejection",
	Setup: Setup{
		NodeID:           "node_id",
		PartitionCount:   1,
		PayloadSizeLimit: 35,
		Codec:            codec,
		Defs: []wfl.Definition{
			wfl.MustDefinition(wfl.Type("workflow_type_exec"), resultBuilder,
				wfl.Step{Execute: successAction},
			),

			wfl.MustDefinition(wfl.Type("workflow_type_rejected"), resultBuilder,
				wfl.Step{Execute: successAction},
				wfl.Step{Execute: successActionFactory("----------too-large-payload----------")},
			),
		},
		// Defs: []wfl.Definition{
		// 	{
		// 		Type: wfl.Type("workflow_type_exec"),
		// 		Steps: []wfl.Step{
		// 			{Execute: successAction},
		// 		},
		// 		ResultBuilder: resultBuilder,
		// 	},
		// 	{
		// 		Type: wfl.Type("workflow_type_rejected"),
		// 		Steps: []wfl.Step{
		// 			{Execute: successAction},
		// 			{Execute: successActionFactory("----------too-large-payload----------")},
		// 		},
		// 		ResultBuilder: resultBuilder,
		// 	},
		// },
	},
	Items: []Item{
		{
			// First Workflow is used to set not 0 WorkflowRejectedEvent.WorkflowOffset.
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
				Type:  "workflow_type_rejected",
				Input: wfl.Input("input_data"),
			},
			WantResult: nil,
			WantError: kapbit.NewKapbitError(
				codes.NewPersistenceError(
					mem.NewTooLargePayloadError(65),
					codes.PersistenceKindRejection),
			),
		},
	},
	Want: Want{
		{
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
					evnt.WorkflowResultEvent{
						NodeID: "node_id",
						Key:    "wfl-1",
						Result: "workflow_result_success",
					},
					evnt.WorkflowCreatedEvent{
						NodeID: "node_id",
						Key:    "wfl-2",
						Type:   "workflow_type_rejected",
						Input:  "input_data",
					},
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-2",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
					},
					&evnt.WorkflowRejectedEvent{
						NodeID:         "node_id",
						Key:            "wfl-2",
						WorkflowOffset: 4,
						Reason: codes.NewPersistenceError(
							mem.NewTooLargePayloadError(65),
							codes.PersistenceKindRejection,
						).Error(),
					},
				},
			},
		},
	},
}
