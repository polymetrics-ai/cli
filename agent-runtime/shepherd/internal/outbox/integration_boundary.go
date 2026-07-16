//go:build integration

package outbox

import "os"

func integrationClaimedBoundary() {
	if os.Getenv("SHEPHERD_INTEGRATION_OUTBOX_CRASH") == "post_claim" {
		os.Exit(97)
	}
}

func integrationExecutionStartedBoundary() {
	if os.Getenv("SHEPHERD_INTEGRATION_OUTBOX_CRASH") == "post_execution_start" {
		os.Exit(96)
	}
}

func integrationPostSendBoundary() {
	if os.Getenv("SHEPHERD_INTEGRATION_OUTBOX_CRASH") == "post_send" {
		os.Exit(98)
	}
}
