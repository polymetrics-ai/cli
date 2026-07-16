//go:build !integration

package outbox

import "testing"

func TestReleaseBuildIgnoresIntegrationOutboxCrashes(t *testing.T) {
	for _, boundary := range []string{"post_claim", "post_execution_start", "post_send"} {
		t.Setenv("SHEPHERD_INTEGRATION_OUTBOX_CRASH", boundary)
		integrationClaimedBoundary()
		integrationExecutionStartedBoundary()
		integrationPostSendBoundary()
	}
}
