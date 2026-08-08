package dockerhub_test

import (
	"testing"

	connectorDefs "polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestDockerHubRegistryPullRateLimitsAreEmbedded(t *testing.T) {
	bundle, err := engine.Load(connectorDefs.FS, "dockerhub")
	if err != nil {
		t.Fatalf("load Docker Hub bundle: %v", err)
	}
	if bundle.RateLimits == nil {
		t.Fatal("Docker Hub bundle has no embedded provider-cited rate_limits.json")
	}
}
