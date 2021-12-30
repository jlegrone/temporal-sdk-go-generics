// Package compilecheck contains assertions on types from other packages.
package compilecheck

import (
	"context"

	"github.com/jlegrone/temporal-sdk-go-generics/client"
	"github.com/jlegrone/temporal-sdk-go-generics/temporal"
	"github.com/jlegrone/temporal-sdk-go-generics/workflow"
)

func stringLengthWorkflow(ctx workflow.Context, str string) (int, error) {
	return len(str), nil
}

func stringLengthActivity(ctx context.Context, str string) (int, error) {
	return len(str), nil
}

var (
	_ externalWorkflowClient[string, int]    = client.NewWorkflowClient(nil, stringLengthWorkflow)
	_ workerChildWorkflowClient[string, int] = workflow.NewChildWorkflowClient(stringLengthWorkflow)
	_ workerActivityClient[string, int]      = workflow.NewActivityClient(stringLengthActivity)
)

type externalWorkflowClient[Req, Resp temporal.Value] interface {
	Start(ctx context.Context, workflowID string, req Req) (*client.WorkflowRun[Resp], error)
	Run(ctx context.Context, workflowID string, req Req) (Resp, error)
}

type workerChildWorkflowClient[Req, Resp temporal.Value] interface {
	Start(ctx workflow.Context, workflowID string, req Req) *workflow.ChildWorkflowFuture[Resp]
	Run(ctx workflow.Context, workflowID string, req Req) (Resp, error)
}

type workerActivityClient[Req, Resp temporal.Value] interface {
	Start(ctx workflow.Context, req Req) *workflow.Future[Resp]
	Run(ctx workflow.Context, req Req) (Resp, error)
	StartLocal(ctx workflow.Context, req Req) *workflow.Future[Resp]
	RunLocal(ctx workflow.Context, req Req) (Resp, error)
}
