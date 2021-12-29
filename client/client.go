package client

import (
	"context"

	"github.com/jlegrone/sdk-go-generics/common"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

type WorkflowRun[T common.Value] struct {
	wrapped client.WorkflowRun
}

func (wfr *WorkflowRun[T]) GetID() string {
	return wfr.wrapped.GetID()
}

func (wfr *WorkflowRun[T]) GetRunID() string {
	return wfr.wrapped.GetRunID()
}

func (wfr *WorkflowRun[T]) Get(ctx context.Context) (T, error) {
	var v T
	if err := wfr.wrapped.Get(ctx, &v); err != nil {
		return v, err
	}
	return v, nil
}

func WrapWorkflowRun[T common.Value](wfr client.WorkflowRun) *WorkflowRun[T] {
	return &WorkflowRun[T]{
		wrapped: wfr,
	}
}

func ExecuteWorkflowFunc[Req, Resp common.Value](
	ctx context.Context,
	c client.Client,
	opts client.StartWorkflowOptions,
	workflow func(workflow.Context, Req) (Resp, error),
	req Req,
) (*WorkflowRun[Resp], error) {
	wfr, err := c.ExecuteWorkflow(ctx, opts, workflow, req)
	if err != nil {
		return nil, err
	}
	return WrapWorkflowRun[Resp](wfr), nil
}

// TODO(jlegrone): This has some problems, mainly that the query func must accept
//                 an argument (which is a common case).
func QueryWorkflowFunc[Req, Resp common.Value](
	ctx context.Context,
	c client.Client,
	workflowID string, runID string,
	queryType string,
	query func(Req) (Resp, error),
	req Req,
) (Resp, error) {
	var result Resp
	encoded, err := c.QueryWorkflow(ctx, workflowID, runID, queryType, req)
	if err != nil {
		return result, err
	}
	if err := encoded.Get(&result); err != nil {
		return result, err
	}
	return result, nil
}
