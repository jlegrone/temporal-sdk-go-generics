package client_test

import (
	"context"
	"testing"

	gclient "github.com/jlegrone/temporal-sdk-go-generics/client"
	"github.com/jlegrone/temporal-sdk-go-generics/temporal"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

type testWorkflowRun[T temporal.Value] struct {
	*testEncodedValue[T]
	client.WorkflowRun
}

type testEncodedValue[T any] struct {
	val T
	err error
}

func (ev *testEncodedValue[T]) HasValue() bool {
	return ev == nil
}

func (ev *testEncodedValue[T]) Get(valuePtr any) error {
	if ev == nil {
		panic("no workflow run data")
	}
	if ev.err != nil {
		return ev.err
	}
	*valuePtr.(*T) = ev.val
	return nil
}

func newWorkflowRun[T temporal.Value](response T, err error) *testWorkflowRun[T] {
	return &testWorkflowRun[T]{
		testEncodedValue: &testEncodedValue[T]{response, err},
	}
}

func (wfr *testWorkflowRun[T]) Get(_ context.Context, valuePtr any) error {
	return wfr.testEncodedValue.Get(valuePtr)
}

type testClient[WorkflowResp, QueryResp temporal.Value] struct {
	client.Client
	workflowRun   *testWorkflowRun[WorkflowResp]
	queryResponse *testEncodedValue[QueryResp]
}

func (tc *testClient[_, _]) ExecuteWorkflow(_ context.Context, _ client.StartWorkflowOptions, _ any, _ ...any) (client.WorkflowRun, error) {
	return tc.workflowRun, nil
}

func (tc *testClient[_, _]) QueryWorkflow(_ context.Context, workflowID, runID, queryType string, args ...any) (converter.EncodedValue, error) {
	return tc.queryResponse, nil
}

func newTestClient() *testClient[string, string] {
	return &testClient[string, string]{
		workflowRun:   newWorkflowRun("Test workflow response", nil),
		queryResponse: &testEncodedValue[string]{"Test query response", nil},
	}
}

func HelloWorkflow(ctx workflow.Context, subject string) (string, error) {
	panic("unimplemented")
}

func TestExecuteWorkflowFunc(t *testing.T) {
	var (
		ctx = context.Background()
		c   = newTestClient()
	)

	wfr, err := gclient.ExecuteWorkflowFunc(ctx, c, client.StartWorkflowOptions{}, HelloWorkflow, "Temporal")
	assert.NoError(t, err)

	result, err := wfr.Get(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "Test workflow response", result)
}

func TestWorkflowClient_Run(t *testing.T) {
	c := newTestClient()
	result, err := gclient.NewWorkflowClient(c, HelloWorkflow).
		WithOptions(client.StartWorkflowOptions{TaskQueue: "test-tq"}).
		Run(context.Background(), "test-wf-id", "Temporal")
	assert.NoError(t, err)
	assert.Equal(t, "Test workflow response", result)
}

func HelloQuery(index int) (string, error) {
	panic("unimplemented")
}

func TestQueryWorkflowFunc(t *testing.T) {
	var (
		ctx = context.Background()
		c   = newTestClient()
	)

	result, err := gclient.QueryWorkflowFunc(ctx, c, "", "", "my-query", HelloQuery, 42)
	assert.NoError(t, err)
	assert.Equal(t, "Test query response", result)
}
