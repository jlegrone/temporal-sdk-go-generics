package workflow_test

import (
	"context"
	"fmt"

	"github.com/jlegrone/temporal-sdk-go-generics/workflow"
)

// Interface implemented by worker struct
type FooWorkerI interface {
	// Workflows
	Foo(workflow.Context, string) (bool, error)
	Bar(workflow.Context, bool) (int, error)
	// Activities
	Baz(context.Context, int) (string, error)
}

// Interface implemented by FooWorkerClient. Optional for workers to embed
// this rather than FooWorkerClient for testing purposes.
type FooWorkerClientI interface {
	ChildWorkflowFoo() *workflow.ChildWorkflowClient[string, bool]
	ChildWorkflowBar() *workflow.ChildWorkflowClient[bool, int]
	ActivityBaz() *workflow.ActivityClient[int, string]
}

type FooWorkerClient struct {
	worker FooWorkerI
	taskQueue string
}

// Ensure client implements interface
var _ FooWorkerClientI = FooWorkerClient{}

func (f FooWorkerClient) ChildWorkflowFoo() *workflow.ChildWorkflowClient[string, bool] {
	return workflow.NewChildWorkflowClient(f.worker.Foo).WithTaskQueue(f.taskQueue)
}

func (f FooWorkerClient) ChildWorkflowBar() *workflow.ChildWorkflowClient[bool, int] {
	return workflow.NewChildWorkflowClient(f.worker.Bar).WithTaskQueue(f.taskQueue)
}

func (f FooWorkerClient) ActivityBaz() *workflow.ActivityClient[int, string] {
	return workflow.NewActivityClient(f.worker.Baz).WithTaskQueue(f.taskQueue)
}

func NewFooWorkerClient(taskQueue string) FooWorkerClient {
	return FooWorkerClient{
		taskQueue: taskQueue,
	}
}

// Worker implementation can remain private since we export a worker client struct.
type fooWorker struct {
	// Embed worker client interface for convenience
	FooWorkerClientI
}

// Ensure worker implements interface
var _ FooWorkerI = fooWorker{}

func (w fooWorker) Foo(ctx workflow.Context, req string) (bool, error) {
	resp, err := w.ActivityBaz().Run(ctx, 1)
	if err != nil {
		return false, err
	}
	return resp == "1", nil
}

func (w fooWorker) Bar(ctx workflow.Context, req bool) (int, error) {
	resp, err := w.ChildWorkflowFoo().Run(ctx, "test-foo", fmt.Sprintf("%t", req))
	if err != nil {
		return -1, err
	}
	if resp {
		return 1, nil
	}
	return 0, nil
}

func (w fooWorker) Baz(ctx context.Context, req int) (string, error) {
	return fmt.Sprintf("%d", req), nil
}
