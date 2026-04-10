package kapbit_test

import (
	"testing"
	"time"

	"github.com/kapbit/kapbit-go"
	evnt "github.com/kapbit/kapbit-go/event"
	evs "github.com/kapbit/kapbit-go/event/store"
	rtm "github.com/kapbit/kapbit-go/runtime"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

var recoverCase = TestCase{
	Name: "Should recover and complete unfinished Workflows",
	Setup: Setup{
		NodeID:         "node_id",
		PartitionCount: 2,
		Codec:          codec,
		Defs: []wfl.Definition{
			wfl.MustDefinition(wfl.Type("workflow_type_exec"), resultBuilder,
				wfl.Step{Execute: successAction},
			),

			wfl.MustDefinition(wfl.Type("workflow_type_comp"), resultBuilder,
				wfl.Step{
					Execute:    successAction,
					Compensate: successAction,
				},
				wfl.Step{Execute: failAction},
			),
		},

		// 	{
		// 		Type: wfl.Type("workflow_type_exec"),
		// 		Steps: []wfl.Step{
		// 			{Execute: successAction},
		// 		},
		// 		ResultBuilder: resultBuilder,
		// 	},
		// 	{
		// 		Type: wfl.Type("workflow_type_comp"),
		// 		Steps: []wfl.Step{
		// 			{
		// 				Execute:    successAction,
		// 				Compensate: successAction,
		// 			},
		// 			{Execute: failAction},
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
	Prepare: func(t *testing.T, s evs.EventStore) (err error) {
		events := []evnt.Event{
			// 0
			evnt.ActiveWriterEvent{
				NodeID:         "node_id_prev",
				PartitionIndex: 0,
			},
			// 1
			// 1 init
			evnt.WorkflowCreatedEvent{
				NodeID: "node_id_prev",
				Key:    "wfl-1",
				Type:   "workflow_type_exec",
				Input:  "input_data",
			},
			// 2
			// 4 init
			evnt.WorkflowCreatedEvent{
				NodeID: "node_id_prev",
				Key:    "wfl-4",
				Type:   "workflow_type_exec",
				Input:  "input_data",
			},
			// 3
			// 1 duplicate
			evnt.WorkflowCreatedEvent{
				NodeID: "node_id_prev",
				Key:    "wfl-1",
				Type:   "workflow_type_exec",
				Input:  "input_data",
			},
			// 4
			// 1
			evnt.StepOutcomeEvent{
				NodeID:      "node_id_prev",
				Key:         "wfl-1",
				OutcomeSeq:  0,
				OutcomeType: wfl.OutcomeTypeExecution,
				Outcome:     wfl.Success("step_ok"),
			},
			// 5
			// 4
			evnt.StepOutcomeEvent{
				NodeID:      "node_id_prev",
				Key:         "wfl-4",
				OutcomeSeq:  0,
				OutcomeType: wfl.OutcomeTypeExecution,
				Outcome:     wfl.Success("step_ok"),
			},
			// 6
			// 3 init
			evnt.WorkflowCreatedEvent{
				NodeID: "node_id_prev",
				Key:    "wfl-3",
				Type:   "workflow_type_exec",
				Input:  "input_data",
			},
			// 7
			// 4 terminate
			&evnt.WorkflowRejectedEvent{
				NodeID:         "node_id_prev",
				Key:            "wfl-4",
				WorkflowOffset: 4,
				Reason:         "result too large",
			},
			// 8
			// 1
			evnt.StepOutcomeEvent{
				NodeID:      "node_id_prev",
				Key:         "wfl-1",
				OutcomeSeq:  0,
				OutcomeType: wfl.OutcomeTypeExecution,
				Outcome:     wfl.Success("step_ok"),
			},
			// 9
			// 3 terminate
			&evnt.DeadLetterEvent{
				NodeID:         "node_id_prev",
				Key:            "wfl-3",
				WorkflowOffset: 4,
			},
		}
		for _, event := range events {
			_, err = s.SaveEvent(t.Context(), 0, event)
			if err != nil {
				return
			}
		}

		// Second parttiion.

		// In the second partition we have one completed workflow wfl-5 and one
		// uncompleted wfl-6.
		events = []evnt.Event{
			// 0
			evnt.ActiveWriterEvent{
				NodeID:         "node_id_prev",
				PartitionIndex: 1,
			},
			// 1
			// wfl-5 init
			evnt.WorkflowCreatedEvent{
				NodeID: "node_id_prev",
				Key:    "wfl-5",
				Type:   "workflow_type_exec",
				Input:  "input_data",
			},
			// 2
			evnt.StepOutcomeEvent{
				NodeID:      "node_id_prev",
				Key:         "wfl-5",
				OutcomeSeq:  0,
				OutcomeType: wfl.OutcomeTypeExecution,
				Outcome:     wfl.Success("step_ok"),
			},
			// 3
			evnt.WorkflowResultEvent{
				NodeID: "node_id_prev",
				Key:    "wfl-5",
				Result: wfl.Result("workflow_result_success"),
			},
			// 4
			// wfl-6 init
			evnt.WorkflowCreatedEvent{
				NodeID: "node_id_prev",
				Key:    "wfl-6",
				Type:   "workflow_type_comp",
				Input:  "input_data",
			},
			// 5
			evnt.StepOutcomeEvent{
				NodeID:      "node_id_prev",
				Key:         "wfl-6",
				OutcomeSeq:  0,
				OutcomeType: wfl.OutcomeTypeExecution,
				Outcome:     wfl.Success("step_ok"),
			},
			// 6
			evnt.StepOutcomeEvent{
				NodeID:      "node_id_prev",
				Key:         "wfl-6",
				OutcomeSeq:  1,
				OutcomeType: wfl.OutcomeTypeExecution,
				Outcome:     wfl.Fail("fail_ok"),
			},
		}
		for _, event := range events {
			_, err = s.SaveEvent(t.Context(), 1, event)
			if err != nil {
				return
			}
		}
		s.SaveCheckpoint(t.Context(), 1, 1)
		return
	},
	Items: []Item{
		{
			Params: wfl.Params{
				ID: "wfl-2_0", // _0 suffix is needed for the workflow to be in the
				// second partition
				Type:  "workflow_type_exec",
				Input: wfl.Input("input_data"),
			},
			WantResult: wfl.Result("workflow_result_success"),
			WantError:  nil,
		},
	},
	Want: Want{
		{
			Wait:  500 * time.Millisecond,
			Chpts: []int64{11, 4},
			Events: [][]evnt.Event{
				// From the previous node:
				// 1, 4 (terminated), 3 (terminated)
				//
				// From the current node:
				// 1 (terminated), 2 (terminated)
				{
					// 0
					evnt.ActiveWriterEvent{
						NodeID:         "node_id_prev",
						PartitionIndex: 0,
					},
					// 1
					// 1 init
					evnt.WorkflowCreatedEvent{
						NodeID: "node_id_prev",
						Key:    "wfl-1",
						Type:   "workflow_type_exec",
						Input:  "input_data",
					},
					// 2
					// 4 init
					evnt.WorkflowCreatedEvent{
						NodeID: "node_id_prev",
						Key:    "wfl-4",
						Type:   "workflow_type_exec",
						Input:  "input_data",
					},
					// 3
					// 1 duplicate
					evnt.WorkflowCreatedEvent{
						NodeID: "node_id_prev",
						Key:    "wfl-1",
						Type:   "workflow_type_exec",
						Input:  "input_data",
					},
					// 4
					// 1
					evnt.StepOutcomeEvent{
						NodeID:      "node_id_prev",
						Key:         "wfl-1",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
					},
					// 5
					// 4
					evnt.StepOutcomeEvent{
						NodeID:      "node_id_prev",
						Key:         "wfl-4",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
					},
					// 6
					// 3 init
					evnt.WorkflowCreatedEvent{
						NodeID: "node_id_prev",
						Key:    "wfl-3",
						Type:   "workflow_type_exec",
						Input:  "input_data",
					},
					// 7
					// 4 terminate
					&evnt.WorkflowRejectedEvent{
						NodeID:         "node_id_prev",
						Key:            "wfl-4",
						WorkflowOffset: 4,
						Reason:         "result too large",
					},
					// 8
					// 1
					evnt.StepOutcomeEvent{
						NodeID:      "node_id_prev",
						Key:         "wfl-1",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
					},
					// 9
					// 3 terminate
					&evnt.DeadLetterEvent{
						NodeID:         "node_id_prev",
						Key:            "wfl-3",
						WorkflowOffset: 4,
					},

					// Next Writer completes wfl-1, and executes wfl-2_0.
					// 10
					evnt.ActiveWriterEvent{
						NodeID:         "node_id",
						PartitionIndex: 0,
					},
					// 11
					// WorkflowCreatedEvent happens before the retry worker will finish
					// the wfl-1.
					evnt.WorkflowCreatedEvent{
						NodeID: "node_id",
						Key:    "wfl-2_0",
						Type:   "workflow_type_exec",
						Input:  "input_data",
					},
					// 12
					evnt.WorkflowResultEvent{
						NodeID: "node_id",
						Key:    "wfl-1",
						Result: "workflow_result_success",
					},
					// 13
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-2_0",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
					},
					// 14
					evnt.WorkflowResultEvent{
						NodeID: "node_id",
						Key:    "wfl-2_0",
						Result: "workflow_result_success",
					},
				},

				// Second partition

				{
					// 0
					evnt.ActiveWriterEvent{
						NodeID:         "node_id_prev",
						PartitionIndex: 1,
					},
					// 1
					// wfl-5 init
					evnt.WorkflowCreatedEvent{
						NodeID: "node_id_prev",
						Key:    "wfl-5",
						Type:   "workflow_type_exec",
						Input:  "input_data",
					},
					// 2
					evnt.StepOutcomeEvent{
						NodeID:      "node_id_prev",
						Key:         "wfl-5",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
					},
					// 3
					evnt.WorkflowResultEvent{
						NodeID: "node_id_prev",
						Key:    "wfl-5",
						Result: wfl.Result("workflow_result_success"),
					},
					// 4
					// wfl-6 init
					evnt.WorkflowCreatedEvent{
						NodeID: "node_id_prev",
						Key:    "wfl-6",
						Type:   "workflow_type_comp",
						Input:  "input_data",
					},
					// 5
					evnt.StepOutcomeEvent{
						NodeID:      "node_id_prev",
						Key:         "wfl-6",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("step_ok"),
					},
					// 6
					evnt.StepOutcomeEvent{
						NodeID:      "node_id_prev",
						Key:         "wfl-6",
						OutcomeSeq:  1,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Fail("fail_ok"),
					},
					// 7
					evnt.ActiveWriterEvent{
						NodeID:         "node_id",
						PartitionIndex: 1,
					},
					// 8
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-6",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeCompensation,
						Outcome:     wfl.Success("step_ok"),
					},
					// 9
					evnt.WorkflowResultEvent{
						NodeID: "node_id",
						Key:    "wfl-6",
						Result: wfl.Result("workflow_result_success"),
					},
				},
			},
		},
	},
	IdempotencyCheckParams: []wfl.Params{
		{ID: "wfl-1", Type: "workflow_type_exec"},
		{ID: "wfl-2_0", Type: "workflow_type_exec"},
		{ID: "wfl-3", Type: "workflow_type_exec"},
		{ID: "wfl-4", Type: "workflow_type_exec"},
		{ID: "wfl-5", Type: "workflow_type_exec"},
		{ID: "wfl-6", Type: "workflow_type_comp"},
	},
}
