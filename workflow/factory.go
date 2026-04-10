package workflow

// Params contains the parameters required to instantiate a workflow.
type Params struct {
	ID    ID
	Type  Type
	Input Input
}

// Factory is responsible for creating new Workflow instances.
// It uses Params to identify the workflow type and input, and Progress
// to restore the workflow's state if it was previously partially executed.
type Factory interface {
	// New instantiates a workflow for the given node, parameters, and progress.
	New(nodeID string, params Params, progress *Progress) (Workflow, error)
}
