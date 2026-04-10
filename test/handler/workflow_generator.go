package handler

import (
	"fmt"
	"math/rand"

	evnt "github.com/kapbit/kapbit-go/event"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

func GenerateRandomEvents(n int) map[wfl.ID][]evnt.Event {
	workflows := make(map[wfl.ID][]evnt.Event)
	for i := range n {
		id := wfl.ID(fmt.Sprintf("wfl-%d", i))
		events := []evnt.Event{evnt.WorkflowCreatedEvent{Key: id}}

		// Optional StepOutcome events (0 to 3)
		numSteps := rand.Intn(2)
		for range numSteps {
			events = append(events, evnt.StepOutcomeEvent{
				Key:         id,
				OutcomeType: wfl.OutcomeTypeExecution,
			})
		}

		// Optional Terminal event
		switch rand.Intn(4) {
		case 1:
			events = append(events, evnt.WorkflowResultEvent{Key: id})
		case 2:
			events = append(events, &evnt.DeadLetterEvent{Key: id})
		case 3:
			events = append(events, &evnt.WorkflowRejectedEvent{Key: id})
		}
		// case 0: do nothing (incomplete workflow)

		workflows[id] = events
	}
	return workflows
}
