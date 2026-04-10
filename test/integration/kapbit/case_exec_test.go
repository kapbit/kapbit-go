package kapbit_test

import (
	evnt "github.com/kapbit/kapbit-go/event"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

var execCase = TestCase{
	Name: "Should be able to exec Wofklow with multiple steps",
	Setup: Setup{
		NodeID:         "node_id",
		PartitionCount: 1,
		Codec:          codec,
		Defs: []wfl.Definition{
			wfl.MustDefinition(wfl.Type("workflow_type_exec"), resultBuilder,
				wfl.Step{Execute: successAction},
				wfl.Step{Execute: successAction},
				wfl.Step{Execute: successAction},
			),
		},
		// Defs: []wfl.Definition{
		// 	{
		// 		Type: wfl.Type("workflow_type_exec"),
		// 		Steps: []wfl.Step{
		// 			{Execute: successAction},
		// 			{Execute: successAction},
		// 			{Execute: successAction},
		// 		},
		// 		ResultBuilder: resultBuilder,
		// 	},
		// },
	},
	Items: []Item{
		{
			Params: wfl.Params{
				ID:    "wfl-1",
				Type:  "workflow_type_exec",
				Input: wfl.Input("input_data"),
			},
			WantResult: wfl.Result("workflow_result_success"),
			WantError:  nil,
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
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-1",
						OutcomeSeq:  1,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
					},
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-1",
						OutcomeSeq:  2,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
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
