package kapbit_test

import (
	"context"
	"errors"
	"reflect"

	codecjson "github.com/kapbit/codec-json-go"
	"github.com/kapbit/kapbit-go/codes"
	wfl "github.com/kapbit/kapbit-go/workflow"
)

var (
	codec = codecjson.NewCodec(
		[]reflect.Type{reflect.TypeFor[string]()},
		[]reflect.Type{reflect.TypeFor[wfl.ValueOutcome[string]]()},
		[]reflect.Type{reflect.TypeFor[string]()},
	)
	resultBuilder wfl.ActionFn[wfl.Result] = func(ctx context.Context,
		input wfl.Input, progress *wfl.Progress) (
		wfl.Result, error,
	) {
		return "workflow_result_success", nil
	}
	errConn = errors.New("connection error")
)

var (
	successActionFactory = func(msg string) wfl.ActionFn[wfl.Outcome] {
		return func(ctx context.Context, input wfl.Input,
			progress *wfl.Progress,
		) (wfl.Outcome, error) {
			return wfl.Success(msg), nil
		}
	}

	successAction wfl.ActionFn[wfl.Outcome] = func(ctx context.Context,
		input wfl.Input, progress *wfl.Progress,
	) (wfl.Outcome, error) {
		return wfl.Success("step_ok"), nil
	}
	failAction wfl.ActionFn[wfl.Outcome] = func(ctx context.Context,
		input wfl.Input, progress *wfl.Progress,
	) (wfl.Outcome, error) {
		return wfl.Fail("step_fail"), nil
	}
)

type retryAction struct {
	retryCount int
	n          int
}

func (a *retryAction) Run(ctx context.Context, input wfl.Input,
	progress *wfl.Progress,
) (wfl.Outcome, error) {
	if a.n < a.retryCount {
		a.n++
		return nil, errConn
	}
	return wfl.Success("retry_step_ok"), nil
}

type circuitBreakerOpenAction struct {
	retryCount int
	n          int
}

func (a *circuitBreakerOpenAction) Run(ctx context.Context, input wfl.Input,
	progress *wfl.Progress,
) (wfl.Outcome, error) {
	if a.n < a.retryCount {
		a.n++
		return nil, codes.NewCircuitBreakerOpenError("service_name")
	}
	return wfl.Success("cb_open_step_ok"), nil
}
