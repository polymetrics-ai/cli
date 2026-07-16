//go:build !integration

package outbox

func integrationClaimedBoundary() {}

func integrationExecutionStartedBoundary() {}

func integrationPostSendBoundary() {}
