# temporal-sdk-go-generics
A go1.18 wrapper to provide a generics based API for Temporal.

## Project Status

This module is strictly a thought experiment and is unlikely to be maintained. Please don't develop dependencies on it.

## Why

### Before Generics

This compiles, but will error at runtime:

```go
package helloworld

import (
	"context"

	"go.temporal.io/sdk/workflow"
)

func Workflow(ctx workflow.Context) (string, error) {
	var result string
	// The activity should be passed a string, not an integer.
	err := workflow.ExecuteActivity(ctx, Activity, 42).Get(ctx, &result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	return "Hello " + name + "!", nil
}
```

### After Generics

This will fail to compile with a human readable error:

```
helloworld.go:12:52: cannot use 42 (untyped int constant) as string value in argument to gworkflow.ExecuteActivityFn
```

```go
package helloworld

import (
	"context"

	gworkflow "github.com/jlegrone/sdk-go-generics/workflow"
	"go.temporal.io/sdk/workflow"
)

func Workflow(ctx workflow.Context) (string, error) {
	// The activity should be passed a string, not an integer.
	return gworkflow.ExecuteActivityFn(ctx, Activity, 42).Get(ctx)
}

func Activity(ctx context.Context, name string) (string, error) {
	return "Hello " + name + "!", nil
}
```