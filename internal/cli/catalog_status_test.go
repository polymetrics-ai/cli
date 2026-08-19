package cli

import (
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestCatalogStatusMessageMakesFreshnessAndPartialityExplicit(t *testing.T) {
	tests := []struct {
		name   string
		status *connectors.DiscoveryStatus
		want   string
	}{
		{name: "static", want: ""},
		{name: "current", status: &connectors.DiscoveryStatus{Complete: true}, want: "catalog status: current"},
		{name: "partial", status: &connectors.DiscoveryStatus{Complete: false}, want: "catalog status: partial; refresh after the provider issue is resolved before relying on this schema"},
		{name: "stale", status: &connectors.DiscoveryStatus{Complete: true, Stale: true}, want: "catalog status: stale; run pm catalog refresh --connection <name> before using this schema"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := catalogStatusMessage(test.status); got != test.want {
				t.Fatalf("catalogStatusMessage() = %q, want %q", got, test.want)
			}
		})
	}
}
