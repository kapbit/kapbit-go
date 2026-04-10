package restorer_test

import (
	evnt "github.com/kapbit/kapbit-go/event"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

// events:
// 1. w + e + e                   type-a
// 2. w + e + r                   type-b  done
// 3. w + e + e + e + c + c       type-a
// 4. w + e + e + c + r           type-c  done
// 5. w + r                       type-d  done
// 6. w + d                       type-d  done
// 7. w + e + j                   type-b  done
// 8. w + d                       type-d  done
// 9. w + d                       type-d  done
// 10. w + e + r	                type-b  done
// 11. w + d                      type-d  done
var (
	DefaultNodeID = "node_id"
	a1            = evnt.ActiveWriterEvent{
		NodeID:         DefaultNodeID,
		PartitionIndex: 1,
		Timestamp:      1,
	}
	a2 = evnt.ActiveWriterEvent{
		NodeID:         DefaultNodeID,
		PartitionIndex: 1,
		Timestamp:      2,
	}
	a3 = evnt.ActiveWriterEvent{
		NodeID:         DefaultNodeID,
		PartitionIndex: 1,
		Timestamp:      3,
	}

	// wfl-1
	w1 = evnt.WorkflowCreatedEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-1"),
		Type:      wfl.Type("type-a"),
		Input:     wfl.Input("input-1"),
		Timestamp: 0,
	}
	e11 = evnt.StepOutcomeEvent{
		NodeID:      a1.NodeID,
		Key:         wfl.ID("wfl-1"),
		Type:        wfl.Type("type-a"),
		OutcomeSeq:  wfl.OutcomeSeq(0),
		OutcomeType: wfl.OutcomeTypeExecution,
		Outcome:     wfl.Success("success-e11"),
		Timestamp:   0,
	}
	e12 = evnt.StepOutcomeEvent{
		NodeID:      a1.NodeID,
		Key:         wfl.ID("wfl-1"),
		Type:        wfl.Type("type-a"),
		OutcomeSeq:  wfl.OutcomeSeq(1),
		OutcomeType: wfl.OutcomeTypeExecution,
		Outcome:     wfl.Success("success-e12"),
		Timestamp:   0,
	}

	// wfl-2
	w2 = evnt.WorkflowCreatedEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-2"),
		Type:      wfl.Type("type-b"),
		Input:     wfl.Input("input-2"),
		Timestamp: 0,
	}
	e21 = evnt.StepOutcomeEvent{
		NodeID:      a1.NodeID,
		Key:         wfl.ID("wfl-2"),
		Type:        wfl.Type("type-b"),
		OutcomeSeq:  wfl.OutcomeSeq(0),
		OutcomeType: wfl.OutcomeTypeExecution,
		Outcome:     wfl.Success("success-e21"),
		Timestamp:   0,
	}
	r2 = evnt.WorkflowResultEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-2"),
		Type:      wfl.Type("type-b"),
		Result:    wfl.Result("result-2"),
		Timestamp: 0,
	}

	// wfl-3
	w3 = evnt.WorkflowCreatedEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-3"),
		Type:      wfl.Type("type-a"),
		Input:     wfl.Input("input-3"),
		Timestamp: 0,
	}
	e31 = evnt.StepOutcomeEvent{
		NodeID:      a1.NodeID,
		Key:         wfl.ID("wfl-3"),
		Type:        wfl.Type("type-a"),
		OutcomeSeq:  wfl.OutcomeSeq(0),
		OutcomeType: wfl.OutcomeTypeExecution,
		Outcome:     wfl.Success("success-e31"),
		Timestamp:   0,
	}
	e32 = evnt.StepOutcomeEvent{
		NodeID:      a1.NodeID,
		Key:         wfl.ID("wfl-3"),
		Type:        wfl.Type("type-a"),
		OutcomeSeq:  wfl.OutcomeSeq(1),
		OutcomeType: wfl.OutcomeTypeExecution,
		Outcome:     wfl.Success("success-e32"),
		Timestamp:   0,
	}
	e33 = evnt.StepOutcomeEvent{
		NodeID:      a1.NodeID,
		Key:         wfl.ID("wfl-3"),
		Type:        wfl.Type("type-a"),
		OutcomeSeq:  wfl.OutcomeSeq(2),
		OutcomeType: wfl.OutcomeTypeExecution,
		Outcome:     wfl.Fail("fail-e33"),
		Timestamp:   0,
	}
	c31 = evnt.StepOutcomeEvent{
		NodeID:      a1.NodeID,
		Key:         wfl.ID("wfl-3"),
		Type:        wfl.Type("type-a"),
		OutcomeSeq:  wfl.OutcomeSeq(0),
		OutcomeType: wfl.OutcomeTypeCompensation,
		Outcome:     wfl.Success("success-c31"),
		Timestamp:   0,
	}
	c32 = evnt.StepOutcomeEvent{
		NodeID:      a1.NodeID,
		Key:         wfl.ID("wfl-3"),
		Type:        wfl.Type("type-a"),
		OutcomeSeq:  wfl.OutcomeSeq(1),
		OutcomeType: wfl.OutcomeTypeCompensation,
		Outcome:     wfl.Success("success-c32"),
		Timestamp:   0,
	}

	// wfl-4
	w4 = evnt.WorkflowCreatedEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-4"),
		Type:      wfl.Type("type-c"),
		Input:     wfl.Input("input-4"),
		Timestamp: 0,
	}
	e41 = evnt.StepOutcomeEvent{
		NodeID:      a1.NodeID,
		Key:         wfl.ID("wfl-4"),
		Type:        wfl.Type("type-c"),
		OutcomeSeq:  wfl.OutcomeSeq(0),
		OutcomeType: wfl.OutcomeTypeExecution,
		Outcome:     wfl.Success("success-e41"),
		Timestamp:   0,
	}
	e42 = evnt.StepOutcomeEvent{
		NodeID:      a1.NodeID,
		Key:         wfl.ID("wfl-4"),
		Type:        wfl.Type("type-c"),
		OutcomeSeq:  wfl.OutcomeSeq(1),
		OutcomeType: wfl.OutcomeTypeExecution,
		Outcome:     wfl.Fail("fail-e42"),
		Timestamp:   0,
	}
	c41 = evnt.StepOutcomeEvent{
		NodeID:      a1.NodeID,
		Key:         wfl.ID("wfl-4"),
		Type:        wfl.Type("type-c"),
		OutcomeSeq:  wfl.OutcomeSeq(0),
		OutcomeType: wfl.OutcomeTypeCompensation,
		Outcome:     wfl.Success("success-c41"),
		Timestamp:   0,
	}
	r4 = evnt.WorkflowResultEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-4"),
		Type:      wfl.Type("type-c"),
		Result:    wfl.Result("result-4"),
		Timestamp: 0,
	}

	// wfl-5
	w5 = evnt.WorkflowCreatedEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-5"),
		Type:      wfl.Type("type-d"),
		Input:     wfl.Input("input-5"),
		Timestamp: 0,
	}
	r5 = evnt.WorkflowResultEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-5"),
		Type:      wfl.Type("type-d"),
		Result:    wfl.Result("result-5"),
		Timestamp: 0,
	}

	// wfl-6
	w6 = evnt.WorkflowCreatedEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-6"),
		Type:      wfl.Type("type-d"),
		Input:     wfl.Input("input-6"),
		Timestamp: 0,
	}
	d6 = &evnt.DeadLetterEvent{
		NodeID:         a1.NodeID,
		Key:            wfl.ID("wfl-6"),
		Type:           wfl.Type("type-d"),
		WorkflowOffset: 5,
		Timestamp:      0,
	}

	// wfl-7
	w7 = evnt.WorkflowCreatedEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-7"),
		Type:      wfl.Type("type-b"),
		Input:     wfl.Input("input-7"),
		Timestamp: 0,
	}
	e71 = evnt.StepOutcomeEvent{
		NodeID:      a1.NodeID,
		Key:         wfl.ID("wfl-4"),
		Type:        wfl.Type("type-b"),
		OutcomeSeq:  wfl.OutcomeSeq(0),
		OutcomeType: wfl.OutcomeTypeExecution,
		Outcome:     wfl.Success("success-e71"),
		Timestamp:   0,
	}
	j7 = &evnt.WorkflowRejectedEvent{
		NodeID:         a1.NodeID,
		Key:            wfl.ID("wfl-7"),
		Type:           wfl.Type("type-b"),
		WorkflowOffset: 5,
		Timestamp:      0,
	}

	// wfl-8
	w8 = evnt.WorkflowCreatedEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-8"),
		Type:      wfl.Type("type-d"),
		Input:     wfl.Input("input-8"),
		Timestamp: 0,
	}
	d8 = &evnt.DeadLetterEvent{
		NodeID:         a1.NodeID,
		Key:            wfl.ID("wfl-8"),
		Type:           wfl.Type("type-d"),
		WorkflowOffset: 1,
		Timestamp:      0,
	}

	// wfl-9
	w9 = evnt.WorkflowCreatedEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-9"),
		Type:      wfl.Type("type-d"),
		Input:     wfl.Input("input-9"),
		Timestamp: 0,
	}
	d9 = &evnt.DeadLetterEvent{
		NodeID:         a1.NodeID,
		Key:            wfl.ID("wfl-9"),
		Type:           wfl.Type("type-d"),
		WorkflowOffset: 3,
		Timestamp:      0,
	}

	// wfl-10
	w10 = evnt.WorkflowCreatedEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-10"),
		Type:      wfl.Type("type-b"),
		Input:     wfl.Input("input-10"),
		Timestamp: 0,
	}
	e101 = evnt.StepOutcomeEvent{
		NodeID:      a1.NodeID,
		Key:         wfl.ID("wfl-10"),
		Type:        wfl.Type("type-b"),
		OutcomeSeq:  wfl.OutcomeSeq(0),
		OutcomeType: wfl.OutcomeTypeExecution,
		Outcome:     wfl.Success("success-e101"),
		Timestamp:   0,
	}
	r10 = evnt.WorkflowResultEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-10"),
		Type:      wfl.Type("type-b"),
		Result:    wfl.Result("result-10"),
		Timestamp: 0,
	}

	// wfl-11
	w11 = evnt.WorkflowCreatedEvent{
		NodeID:    a1.NodeID,
		Key:       wfl.ID("wfl-11"),
		Type:      wfl.Type("type-d"),
		Input:     wfl.Input("input-11"),
		Timestamp: 0,
	}
	d11 = &evnt.DeadLetterEvent{
		NodeID:         a1.NodeID,
		Key:            wfl.ID("wfl-11"),
		Type:           wfl.Type("type-d"),
		WorkflowOffset: 6,
		Timestamp:      0,
	}
)

type EventRecord struct {
	offset     int64
	beforeChpt bool
	event      evnt.Event
}

var (
	historyP1 = []EventRecord{
		{0, false, a1},
		{1, false, w3},
		{2, false, w1},
		{3, false, e31},
		{4, false, w5},
		{5, false, e11},
		{6, false, e32},
		{7, false, a3},
		{8, false, r5},
		{9, false, e12},
		{10, false, e33},
		{11, false, c31},
		{12, false, c32},
	}
	historyP2 = []EventRecord{
		{2, false, w2},
		{3, false, e41},
		{4, false, e21},
		{5, false, w6},
		{6, false, e42},
		{7, false, r2},
		{8, false, w7},
		{9, false, e71},
		{10, false, d6},
		{11, false, c41},
		{12, false, r4},
		{13, false, j7},
	}
	historyP3 = []EventRecord{
		{3, true, w9},
		{4, true, w10},
		{6, true, w11},
		{7, false, r10},
		{8, false, d11},
		{9, false, d9},
	}
)
