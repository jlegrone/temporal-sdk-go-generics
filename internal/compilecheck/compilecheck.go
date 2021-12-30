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
	// Really gross, sorry. The "Self" type parameter allows us to verify
	// that interface impelementations return themselves for certain methods.
	_ externalWorkflowClient[*client.WorkflowClient[string, int], string, int]           = client.NewWorkflowClient(nil, stringLengthWorkflow)
	_ workerChildWorkflowClient[*workflow.ChildWorkflowClient[string, int], string, int] = workflow.NewChildWorkflowClient(stringLengthWorkflow)
	_ workerActivityClient[*workflow.ActivityClient[string, int], string, int]           = workflow.NewActivityClient(stringLengthActivity)
)

type optionsBuilder[Self any] interface {
	WithTaskQueue(taskQueue string) Self
}

type externalWorkflowClient[Self any, Req, Resp temporal.Value] interface {
	Start(ctx context.Context, workflowID string, req Req) (*client.WorkflowRun[Resp], error)
	Run(ctx context.Context, workflowID string, req Req) (Resp, error)
	WithOptions(opts client.StartWorkflowOptions) Self
	optionsBuilder[Self]
}

type workerChildWorkflowClient[Self any, Req, Resp temporal.Value] interface {
	Start(ctx workflow.Context, workflowID string, req Req) *workflow.ChildWorkflowFuture[Resp]
	Run(ctx workflow.Context, workflowID string, req Req) (Resp, error)
	WithOptions(opts workflow.ChildWorkflowOptions) Self
	optionsBuilder[Self]
}

type workerActivityClient[Self any, Req, Resp temporal.Value] interface {
	Start(ctx workflow.Context, req Req) *workflow.Future[Resp]
	StartLocal(ctx workflow.Context, req Req) *workflow.Future[Resp]
	Run(ctx workflow.Context, req Req) (Resp, error)
	RunLocal(ctx workflow.Context, req Req) (Resp, error)
	WithOptions(opts workflow.ActivityOptions) Self
	WithLocalOptions(opts workflow.LocalActivityOptions) Self
	optionsBuilder[Self]
}
