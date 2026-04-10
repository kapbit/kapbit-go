package handler

import (
	"fmt"
	"sync"
	"testing"

	evnt "github.com/kapbit/kapbit-go/event"
	"github.com/kapbit/kapbit-go/event/persist"
	"github.com/kapbit/kapbit-go/test"
	wfl "github.com/kapbit/kapbit-go/workflow"
	asserterror "github.com/ymz-ncnk/assert/error"
	assertfatal "github.com/ymz-ncnk/assert/fatal"
)

// Callback is a function type used to schedule the next event for a workflow.
type Callback func(wid wfl.ID, next evnt.Event)

// Scenario is used for multi-partition, multi-workflow handler tests. It collects
// handled events, and produced checkpoints. Also it triggers next workflow event
// handling by registered callback, because events should be handled in order.
type Scenario struct {
	Workflows         map[wfl.ID][]evnt.Event
	events            map[wfl.ID][]evnt.Event
	eventsByPartition [][]evnt.Event
	chptsByPartition  [][]int64
	callback          Callback
	mu                sync.Mutex
}

// NewScenario creates a new Scenario with the given number of partitions and workflows.
func NewScenario(partitions int, workflows map[wfl.ID][]evnt.Event) *Scenario {
	chpts := make([][]int64, partitions)
	scenario := Scenario{
		Workflows:         workflows,
		events:            map[wfl.ID][]evnt.Event{},
		eventsByPartition: make([][]evnt.Event, partitions),
		chptsByPartition:  chpts,
	}
	return &scenario
}

// AppendEvent records an event for a specific partition.
// It tracks events by workflow and partition, and triggers the callback
// to schedule the next event for the workflow.
func (s *Scenario) AppendEvent(partition int, event evnt.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if wevent, ok := event.(evnt.WorkflowEvent); ok {
		wid := wevent.WorkflowID()
		sl, pst := s.events[wid]
		if !pst {
			sl = []evnt.Event{event}
		} else {
			sl = append(sl, event)
		}
		s.eventsByPartition[partition] = append(s.eventsByPartition[partition], event)
		s.events[wid] = sl

		if len(s.Workflows[wid]) <= len(sl) {
			s.callback(wid, nil)
			return
		}
		next := s.Workflows[wid][len(sl)]
		s.callback(wid, next)
	}
}

// AppendChpt records a checkpoint for a partition.
func (s *Scenario) AppendChpt(partition int, chpt int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chptsByPartition[partition] = append(s.chptsByPartition[partition], chpt)
}

// OnReceive sets the callback to be invoked when an event is appended.
func (s *Scenario) OnReceive(fn Callback) {
	s.callback = fn
}

// EventsCount returns the total number of events across all workflows.
func (s *Scenario) EventsCount() int {
	count := 0
	for _, events := range s.Workflows {
		count += len(events)
	}
	return count
}

// Verify checks that:
// 1. All workflow events were received back.
// 2. All workflows before the checkpoint are terminated, and the next workflow
// is not terminated.
func (s *Scenario) Verify(t *testing.T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.verifyWorkflowCompleteness(t)
	s.verifyCheckpointInvariants(t)
}

func (s *Scenario) verifyWorkflowCompleteness(t *testing.T) {
	// Do not remove -------------------------------------------------------------
	//
	// for partition, events := range s.eventsByPartition {
	// 	fmt.Printf("partition %d events:\n", partition)
	// 	for i, event := range events {
	// 		fmt.Printf("%d %s %T\n", i, event.(evnt.WorkflowEvent).WorkflowID(), event)
	// 	}
	// }
	// ---------------------------------------------------------------------------

	for wid, wantEvents := range s.Workflows {
		events, pst := s.events[wid]
		if !pst {
			t.Errorf("workflow %s was defined in scenario but never processed", wid)
			continue
		}
		asserterror.EqualDeep(t, events, wantEvents)
	}
}

func (s *Scenario) verifyCheckpointInvariants(t *testing.T) {
	for partition, items := range s.chptsByPartition {
		// Do not remove -----------------------------------------------------------
		//
		// fmt.Printf("partition %d chpts: %v\n", partition, items)
		// -------------------------------------------------------------------------

		if len(items) == 0 {
			s.nextOneNotCompleted(t, partition, -1)
			continue
		}
		lastChpt := items[len(items)-1]
		s.allPrevCompleted(t, partition, int(lastChpt))
		s.nextOneNotCompleted(t, partition, int(lastChpt))
	}
}

func (s *Scenario) allPrevCompleted(t *testing.T, partition, index int) {
	for i := 0; i <= index; i++ {
		createdEvent, ok := s.eventsByPartition[partition][i].(evnt.WorkflowCreatedEvent)
		if !ok {
			continue
		}
		wid := createdEvent.WorkflowID()
		asserterror.Equal(t, s.completedWorkflow(wid), true)
	}
}

func (s *Scenario) nextOneNotCompleted(t *testing.T, partition, index int) {
	nextIndex := index + 1
	events := s.eventsByPartition[partition]
	var wid wfl.ID

	for i := nextIndex; i < len(events); i++ {
		createdEvent, ok := events[i].(evnt.WorkflowCreatedEvent)
		if ok {
			wid = createdEvent.WorkflowID()
			break
		}
	}
	if wid == "" {
		return
	}
	asserterror.Equal(t, s.completedWorkflow(wid), false)
}

func (s *Scenario) completedWorkflow(wid wfl.ID) bool {
	var (
		wevents []evnt.Event
		pst     bool
	)
	if wevents, pst = s.Workflows[wid]; !pst {
		panic(fmt.Sprintf("could not find workflow %s in Scenario.Workflows", wid))
	}
	l := len(wevents)
	return wevents[l-1].(evnt.WorkflowEvent).IsTerminal()
}

func RunScenario(t *testing.T, tc TestCase) {
	t.Run(tc.Name, func(t *testing.T) {
		var (
			// One channel is used to hold all the events of one Workflow.
			chs          = makeChannels(tc)
			wg           = sync.WaitGroup{}
			handler, err = persist.NewPersistHandler(tc.Setup.Store, tc.Setup.Chptman,
				tc.Setup.Tracker,
				persist.WithLogger(test.NopLogger))
		)
		if err != nil {
			t.Fatal(err)
		}

		tc.Scenario.OnReceive(func(wid wfl.ID, next evnt.Event) {
			go func() {
				chs[wid] <- next
			}()
		})

		for wid, wevent := range tc.Scenario.Workflows {
			wg.Add(1)

			go func() {
				err := handler.Handle(t.Context(), wevent[0])
				assertfatal.EqualError(t, err, nil)

				for {
					event := <-chs[wid]
					if event == nil {
						wg.Done()
						return
					}
					err = handler.Handle(t.Context(), event)
					asserterror.EqualError(t, err, nil)
				}
			}()
		}
		wg.Wait()
		tc.Want.Check(t, tc.Setup)
	})
}

func makeChannels(tc TestCase) map[wfl.ID](chan evnt.Event) {
	chs := map[wfl.ID](chan evnt.Event){}
	for wid := range tc.Scenario.Workflows {
		chs[wid] = make(chan evnt.Event, 1000)
	}
	return chs
}
