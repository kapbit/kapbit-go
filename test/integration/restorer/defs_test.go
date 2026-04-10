package restorer_test

import (
	"github.com/kapbit/kapbit-go/test/mock"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

var defs = wfl.Definitions{
	wfl.MustDefinition(wfl.Type("type-a"), nil,
		wfl.Step{
			Execute:    mock.NewActionMock[wfl.Outcome](),
			Compensate: mock.NewActionMock[wfl.Outcome](),
		},
		wfl.Step{
			Execute:    mock.NewActionMock[wfl.Outcome](),
			Compensate: mock.NewActionMock[wfl.Outcome](),
		},
		wfl.Step{
			Execute: mock.NewActionMock[wfl.Outcome](),
		},
	),

	wfl.MustDefinition(wfl.Type("type-b"), nil,
		wfl.Step{
			Execute: mock.NewActionMock[wfl.Outcome](),
		},
	),

	wfl.MustDefinition(wfl.Type("type-c"), nil,
		wfl.Step{
			Execute:    mock.NewActionMock[wfl.Outcome](),
			Compensate: mock.NewActionMock[wfl.Outcome](),
		},
		wfl.Step{
			Execute: mock.NewActionMock[wfl.Outcome](),
		},
	),

	wfl.MustDefinition(wfl.Type("type-c"), nil, wfl.Step{}),

	wfl.MustDefinition(wfl.Type("type-d"), nil),
}
