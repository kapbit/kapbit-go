package executor

import (
	"context"
	"errors"
	"log/slog"

	rtm "github.com/kapbit/kapbit-go/runtime"
)

type DoNothing struct {
	cause error
}

func (a DoNothing) Do(_ context.Context, service *Engine) error {
	return a.cause
}

// -----------------------------------------------------------------------------

type EmitWorkflowRejection struct {
	service *Engine
	ref     *rtm.WorkflowRef
	cause   error
	logger  *slog.Logger
}

func (a EmitWorkflowRejection) Do(ctx context.Context, service *Engine) error {
	a.logger.Warn("reject workflow",
		"wid", a.ref.Workflow().Spec().ID,
		"wtype", a.ref.Workflow().Spec().Def.Type,
		"err", a.cause,
	)
	err := a.service.EmitRejection(ctx, a.ref, a.cause.Error())
	if err != nil {
		return errors.Join(a.cause, err)
	}
	return a.cause
}

// -----------------------------------------------------------------------------

type FenceKapbit struct {
	cause error
}

func (a FenceKapbit) Do(_ context.Context, service *Engine) error {
	service.onFenced()
	return a.cause
}

// -----------------------------------------------------------------------------

type RetryWorkflowExecution struct {
	ref    *rtm.WorkflowRef
	cause  error
	logger *slog.Logger
}

func (a RetryWorkflowExecution) Do(_ context.Context, service *Engine) error {
	service.runtime.ScheduleRetry(a.ref)
	return NewWillRetryLaterError(a.cause)
}

// -----------------------------------------------------------------------------

type OnErrorAction interface {
	Do(ctx context.Context, service *Engine) error
}
