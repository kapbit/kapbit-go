package kapbit_test

import (
	"time"

	kapbit "github.com/kapbit/kapbit-go"
	evnt "github.com/kapbit/kapbit-go/event"
	rtm "github.com/kapbit/kapbit-go/runtime"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

var partitionsCase = TestCase{
	Name: "Should be able to save events on different partitions",
	Setup: Setup{
		NodeID:         "node_id",
		PartitionCount: 2,
		Codec:          codec,
		Defs: []wfl.Definition{
			wfl.MustDefinition(wfl.Type("workflow_type_exec"), resultBuilder,
				wfl.Step{Execute: successAction},
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
		// },
		Opts: []kapbit.SetOption{
			kapbit.WithRuntime(
				rtm.WithCheckpointSize(1),
			),
		},
		Parallel: true,
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
		{
			Params: wfl.Params{
				ID:    "wfl-2",
				Type:  "workflow_type_exec",
				Input: wfl.Input("input_data"),
			},
			WantResult: wfl.Result("workflow_result_success"),
			WantError:  nil,
		},
		{
			Params: wfl.Params{
				ID:    "wfl-3",
				Type:  "workflow_type_exec",
				Input: wfl.Input("input_data"),
			},
			WantResult: wfl.Result("workflow_result_success"),
			WantError:  nil,
		},
		{
			Params: wfl.Params{
				ID:    "wfl-4",
				Type:  "workflow_type_exec",
				Input: wfl.Input("input_data"),
			},
			WantResult: wfl.Result("workflow_result_success"),
			WantError:  nil,
		},
		{
			Params: wfl.Params{
				ID:    "wfl-5",
				Type:  "workflow_type_exec",
				Input: wfl.Input("input_data"),
			},
			WantResult: wfl.Result("workflow_result_success"),
			WantError:  nil,
		},
	},
	Want: Want{
		{
			Wait: 500 * time.Millisecond,
			// Chpts: []int64{7, 4},
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
						Key:    "wfl-3",
						Type:   "workflow_type_exec",
						Input:  "input_data",
					},
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-3",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
					},
					evnt.WorkflowResultEvent{
						NodeID: "node_id",
						Key:    "wfl-3",
						Result: "workflow_result_success",
					},
					evnt.WorkflowCreatedEvent{
						NodeID: "node_id",
						Key:    "wfl-5",
						Type:   "workflow_type_exec",
						Input:  "input_data",
					},
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-5",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
					},
					evnt.WorkflowResultEvent{
						NodeID: "node_id",
						Key:    "wfl-5",
						Result: "workflow_result_success",
					},
				},
				{
					evnt.ActiveWriterEvent{
						NodeID:         "node_id",
						PartitionIndex: 1,
					},
					evnt.WorkflowCreatedEvent{
						NodeID: "node_id",
						Key:    "wfl-2",
						Type:   "workflow_type_exec",
						Input:  "input_data",
					},
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-2",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
					},
					evnt.WorkflowResultEvent{
						NodeID: "node_id",
						Key:    "wfl-2",
						Result: "workflow_result_success",
					},
					evnt.WorkflowCreatedEvent{
						NodeID: "node_id",
						Key:    "wfl-4",
						Type:   "workflow_type_exec",
						Input:  "input_data",
					},
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-4",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
					},
					evnt.WorkflowResultEvent{
						NodeID: "node_id",
						Key:    "wfl-4",
						Result: "workflow_result_success",
					},
				},
			},
		},
	},
	IdempotencyCheckParams: []wfl.Params{
		{ID: "wfl-1", Type: "workflow_type_exec"},
		{ID: "wfl-2", Type: "workflow_type_exec"},
		{ID: "wfl-3", Type: "workflow_type_exec"},
		{ID: "wfl-4", Type: "workflow_type_exec"},
		{ID: "wfl-5", Type: "workflow_type_exec"},
	},
}
