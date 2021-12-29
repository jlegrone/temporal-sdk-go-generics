package workflow_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	gworkflow "github.com/jlegrone/sdk-go-generics/workflow"
)

// None represents an empty value. Useful for constructing futures that will only return error.
type None struct{}

// Workflow is a Hello World workflow definition, significantly over-engineered to exercise
// more of the workflow package than would usually be necessary.
func Workflow(ctx workflow.Context, name string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	})

	// Start a timer to handle workflow timeout. Normally we'd just register the
	// future returned by workflow.NewTimer with a selector directly, but this
	// allows us to test generic channel sending/receiving.
	timoutChannel := gworkflow.NewNamedChannel[None](ctx, "timeout")
	workflow.Go(ctx, func(ctx workflow.Context) {
		if err := workflow.NewTimer(ctx, time.Minute).Get(ctx, nil); err != nil {
			panic(err)
		}
		timoutChannel.Send(ctx, None{})
	})

	punctuation, err := gworkflow.SideEffect(ctx, func(ctx workflow.Context) string {
		return "!"
	})
	if err != nil {
		return "", err
	}

	var (
		greeting       string
		selectErr      error
		selector       = workflow.NewSelector(ctx)
		activityFuture = gworkflow.ExecuteActivityFunc(ctx, Activity, name)
	)
	selector = gworkflow.SelectorAddFuture(ctx, selector, activityFuture, func(val string, err error) {
		if err != nil {
			selectErr = err
			return
		}
		workflow.GetLogger(ctx).Info("Received activity result", "result", val)
		greeting = val
	})
	selector = gworkflow.SelectorAddReceive(selector, timoutChannel.ReceiveChannel, func(_ None, more bool) {
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
