package kapbit_test

import (
	"time"

	"github.com/kapbit/kapbit-go"
	"github.com/kapbit/kapbit-go/codes"
	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/kapbit/kapbit-go/executor"
	swtch "github.com/kapbit/kapbit-go/support/retry/switch"
	wrk "github.com/kapbit/kapbit-go/worker"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

var circuitBrakerOpenCase = TestCase{
	Name: "Retries should be slow down when circuit breaker is open",
	Setup: Setup{
		NodeID:         "node_id",
		PartitionCount: 1,
		Codec:          codec,
		Defs: []wfl.Definition{
			wfl.MustDefinition(wfl.Type("workflow_type_cb_open"), resultBuilder,
				wfl.Step{Execute: &circuitBreakerOpenAction{retryCount: 4}},
			),
		},
		// Defs: []wfl.Definition{
		// 	{
		// 		Type: wfl.Type("workflow_type_cb_open"),
		// 		Steps: []wfl.Step{
		// 			{Execute: &circuitBreakerOpenAction{retryCount: 4}},
		// 		},
		// 		ResultBuilder: resultBuilder,
		// 	},
		// },
		Opts: []kapbit.SetOption{
			kapbit.WithWorker(
				wrk.WithMaxAttempts(5),
				wrk.WithPolicyOptions(
					swtch.WithNormalInterval(50*time.Millisecond),
					swtch.WithSlowInterval(100*time.Millisecond),
				),
			),
		},
	},
	Items: []Item{
		{
			Params: wfl.Params{
				ID:    "wfl-1",
				Type:  "workflow_type_cb_open",
				Input: wfl.Input("input_data"),
			},
			WantResult: nil,
			WantError: kapbit.NewKapbitError(
				executor.NewWillRetryLaterError(
					wfl.NewExecutionError(codes.NewCircuitBreakerOpenError("service_name")),
				),
			),
		},
	},
	Want: Want{
		{
			Wait:  200 * time.Millisecond,
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
						Type:   "workflow_type_cb_open",
						Input:  "input_data",
					},
				},
			},
		},
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
						Type:   "workflow_type_cb_open",
						Input:  "input_data",
					},
					evnt.StepOutcomeEvent{
						NodeID:      "node_id",
						Key:         "wfl-1",
						OutcomeSeq:  0,
						OutcomeType: wfl.OutcomeTypeExecution,
						Outcome:     wfl.Success("cb_open_step_ok"),
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
