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

// Workflow is a Hello World workflow definition.
func Workflow(ctx workflow.Context, name string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	punctuation, err := gworkflow.SideEffect(ctx, func(ctx workflow.Context) string {
		return "!"
	})
	if err != nil {
		return "", err
	}

	greeting, err := gworkflow.ExecuteActivityFn(ctx, Activity, name).Get(ctx)
	if err != nil {
		return "", err
	}

	return greeting + punctuation, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	return fmt.Sprintf("Hello %s", name), nil
}

func Test_Workflow(t *testing.T) {
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
