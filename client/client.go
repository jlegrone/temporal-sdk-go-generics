package client

import (
	"context"

	"github.com/jlegrone/temporal-sdk-go-generics/internal/reflect"
	"github.com/jlegrone/temporal-sdk-go-generics/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

type WorkflowRun[T temporal.Value] struct {
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

func WrapWorkflowRun[T temporal.Value](wfr client.WorkflowRun) *WorkflowRun[T] {
	return &WorkflowRun[T]{
		wrapped: wfr,
	}
}

// WorkflowClient provides type safe methods for interacting with a given workflow type.
type WorkflowClient[Req, Resp temporal.Value] struct {
	client       client.Client
	defaultOpts  client.StartWorkflowOptions
	workflowType string
}

// WithOptions sets default start options for workflow executions.
func (wfc *WorkflowClient[Req, Resp]) WithOptions(opts client.StartWorkflowOptions) *WorkflowClient[Req, Resp] {
	// TODO: merge these instead of overriding?
	wfc.defaultOpts = opts
	return wfc
}

// Start begins a workflow execution but does not block on its completion.
func (wfc *WorkflowClient[Req, Resp]) Start(ctx context.Context, workflowID string, req Req) (*WorkflowRun[Resp], error) {
	opts := wfc.defaultOpts
	opts.ID = workflowID
	wfr, err := wfc.client.ExecuteWorkflow(ctx, opts, wfc.workflowType, req)
	if err != nil {
		return nil, err
	}
	return WrapWorkflowRun[Resp](wfr), nil
}

// Run starts a workflow and waits for it to complete, returning the result.
func (wfc *WorkflowClient[Req, Resp]) Run(ctx context.Context, workflowID string, req Req) (Resp, error) {
	wfr, err := wfc.Start(ctx, workflowID, req)
	if err != nil {
		var emptyResult Resp
		return emptyResult, err
	}
	return wfr.Get(ctx)
}

// NewWorkflowClient instantiates a client for a given workflow func.
func NewWorkflowClient[Req, Resp temporal.Value](
	client client.Client,
	workflow func(workflow.Context, Req) (Resp, error),
) *WorkflowClient[Req, Resp] {
	workflowType, _ := reflect.GetFunctionName(workflow)
	return NewNamedWorkflowClient[Req, Resp](client, workflowType)
}

// NewNamedWorkflowClient instantiates a client for a given workflow type.
func NewNamedWorkflowClient[Req, Resp temporal.Value](
	client client.Client,
	workflowType string,
) *WorkflowClient[Req, Resp] {
	return &WorkflowClient[Req, Resp]{
		client:       client,
		workflowType: workflowType,
	}
}

// Alternative to WorkflowClient. We should probably choose one or the other to expose.
func ExecuteWorkflowFunc[Req, Resp temporal.Value](
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
func QueryWorkflowFunc[Req, Resp temporal.Value](
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
