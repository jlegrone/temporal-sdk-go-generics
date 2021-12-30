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
helloworld.go:11:55: cannot use 42 (untyped int constant) as string value in argument to workflow.NewActivityClient(Activity).Run
```

The code to get the activity's return value can also now be more terse.

```go
package helloworld

import (
	"context"

	"github.com/jlegrone/temporal-sdk-go-generics/workflow"
)

func Workflow(ctx workflow.Context) (string, error) {
	// The activity should be passed a string, not an integer.
	return workflow.NewActivityClient(Activity).Run(ctx, 42)
}

func Activity(ctx context.Context, name string) (string, error) {
	return "Hello " + name + "!", nil
}
```
