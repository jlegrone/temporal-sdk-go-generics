package workflow_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	"github.com/jlegrone/temporal-sdk-go-generics/temporal"
	"github.com/jlegrone/temporal-sdk-go-generics/workflow"
)

// Workflow is a Hello World workflow definition, significantly over-engineered to exercise
// more of the workflow package than would usually be necessary.
func Workflow(ctx workflow.Context, name string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	})

	// Start a timer to handle workflow timeout. Normally we'd just register the
	// future returned by workflow.NewTimer with a selector directly, but this
	// allows us to test generic channel sending/receiving.
	timoutChannel := workflow.NewNamedChannel[temporal.None](ctx, "timeout")
	workflow.Go(ctx, func(ctx workflow.Context) {
		if err := workflow.NewTimer(ctx, time.Minute).Get(ctx, nil); err != nil {
			panic(err)
		}
		timoutChannel.Send(ctx, temporal.None{})
	})

	punctuation, err := workflow.SideEffect(ctx, func(ctx workflow.Context) string {
		return "!"
	})
	if err != nil {
		return "", err
	}

	var (
		greeting          string
		selectErr         error
		selector          = workflow.NewSelector(ctx)
		activityClient    = workflow.NewActivityClient(Activity).WithTaskQueue("helloworld")
		activityExecution = activityClient.Start(ctx, name)
	)
	selector = activityExecution.AddFuture(ctx, selector, func(val string, err error) {
		if err != nil {
			selectErr = err
			return
		}
		greeting = val
	})
	selector = timoutChannel.AddReceive(selector, func(_ temporal.None, more bool) {
		selectErr = fmt.Errorf("my custom timer fired: %w", workflow.ErrDeadlineExceeded)
	})
	selector.Select(ctx)

	if selectErr != nil {
		return "", selectErr
	}

	return greeting + punctuation, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	return fmt.Sprintf("Hello %s", name), nil
}

func Test_Workflow_success(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activity implementation
	env.OnActivity(Activity, mock.Anything, "Temporal").Return("Hello Temporal", nil)

	env.ExecuteWorkflow(Workflow, "Temporal")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "Hello Temporal!", result)
}

func Test_Workflow_timeout(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activity implementation
	env.OnActivity(Activity, mock.Anything, "Temporal").Return("Hello Temporal", nil).After(time.Minute)

	env.ExecuteWorkflow(Workflow, "Temporal")

	require.True(t, env.IsWorkflowCompleted())
	require.EqualError(t, env.GetWorkflowError(),
		"workflow execution error (type: Workflow, workflowID: default-test-workflow-id, runID: default-test-run-id): my custom timer fired: deadline exceeded (type: ScheduleToClose)",
	)
}
