package impl

import (
	"log/slog"

	evnt "github.com/kapbit/kapbit-go/event"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

type Factory struct {
	defs   map[wfl.Type]wfl.Definition
	tools  *evnt.Kit
	logger *slog.Logger
}

func NewFactory(defs wfl.Definitions, tools *evnt.Kit,
	logger *slog.Logger,
) *Factory {
	return &Factory{defs: defs.ToMap(), tools: tools, logger: logger}
}

// New creates a new workflow.
//
// Returns wfl.WorkflowTypeNotRegisteredError if the workflow type is not registered.
func (f *Factory) New(nodeID string, params wfl.Params, progress *wfl.Progress) (
	workflow wfl.Workflow, err error,
) {
	var (
		spec = wfl.Spec{ID: params.ID, Input: params.Input}
		pst  bool
	)
	spec.Def, pst = f.defs[params.Type]
	if !pst {
		err = wfl.NewWorkflowTypeNotRegisteredError(params.ID, params.Type)
		return
	}
	workflow = NewWorkflow(nodeID, spec, progress, f.tools, f.logger)
	return
}
