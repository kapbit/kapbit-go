package worker_test

import (
	"context"
	"time"

	rtm "github.com/kapbit/kapbit-go/runtime"
	"github.com/kapbit/kapbit-go/test/mock"
	"github.com/ymz-ncnk/mok"
)

func StopTestCase() TestCase {
	var (
		ctx, cancel = context.WithTimeout(context.Background(), 50*time.Millisecond)
		engine      = mock.NewEngineMock()
	)
	defer cancel()

	engine.RegisterRetryQueue(
		func() <-chan *rtm.WorkflowRef {
			return make(chan *rtm.WorkflowRef)
		},
	)

	return TestCase{
		Name: "Should stop on ctx cancel",
		Config: TestConfig{
			Engine: engine,
		},
		Ctx:  ctx,
		mock: []*mok.Mock{},
	}
}
