package restorer_test

import rtm "github.com/kapbit/kapbit-go/runtime"

func refs() []*rtm.WorkflowRef {
	ref3 := rtm.NewWorkflowRef(workflow3, 0)
	ref3.MarkSaved()

	ref1 := rtm.NewWorkflowRef(workflow1, 0)
	ref1.MarkSaved()

	// Completed workflows
	// ref5 := rtm.NewWorkflowRef(workflow5, 0)
	// ref5.MarkSaved()
	//
	// ref4 := rtm.NewWorkflowRef(workflow4, 0)
	// ref4.MarkSaved()
	//
	// ref2 := rtm.NewWorkflowRef(workflow2, 0)
	// ref2.MarkSaved()
	//
	return []*rtm.WorkflowRef{ref3, ref1 /* ref5, ref4, ref2 */}
}
