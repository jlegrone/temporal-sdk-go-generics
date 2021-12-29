package workflow

import "go.temporal.io/sdk/workflow"

// We alias a bunch of types here to make it easier to test whether the workflow
// package is a "drop-in" replacement for existing code.

type (
	Context              = workflow.Context
	ActivityOptions      = workflow.ActivityOptions
	LocalActivityOptions = workflow.LocalActivityOptions
	ChildWorkflowOptions = workflow.ChildWorkflowOptions
	ContinueAsNewError   = workflow.ContinueAsNewError
)

var (
	Go                       = workflow.Go
	GetLogger                = workflow.GetLogger
	WithActivityOptions      = workflow.WithActivityOptions
	WithLocalActivityOptions = workflow.WithLocalActivityOptions
	WithChildOptions         = workflow.WithChildOptions
	WithTaskQueue            = workflow.WithTaskQueue
	WithWorkflowTaskQueue    = workflow.WithWorkflowTaskQueue
	WithCancel               = workflow.WithCancel
	WithRetryPolicy          = workflow.WithRetryPolicy
	WithDataConverter        = workflow.WithDataConverter
	NewSelector              = workflow.NewSelector
	NewTimer                 = workflow.NewTimer
	NewDisconnectedContext   = workflow.NewDisconnectedContext
	NewWaitGroup             = workflow.NewWaitGroup
	NewContinueAsNewError    = workflow.NewContinueAsNewError
)

var (
	ErrDeadlineExceeded = workflow.ErrDeadlineExceeded
	ErrCanceled         = workflow.ErrCanceled
	ErrSessionFailed    = workflow.ErrSessionFailed
)
