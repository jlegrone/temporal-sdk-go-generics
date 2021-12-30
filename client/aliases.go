package client

import (
	"go.temporal.io/sdk/client"
)

// We alias a bunch of types here to make it easier to test whether the client
// package is a "drop-in" replacement for existing code.

type (
	Client               = client.Client
	StartWorkflowOptions = client.StartWorkflowOptions
)
