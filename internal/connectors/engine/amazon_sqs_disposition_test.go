package engine

import (
	"testing"

	"polymetrics.ai/internal/connectors/defs"
)

func TestAmazonSQSDispositionBundleLoads(t *testing.T) {
	bundle, err := Load(defs.FS, "amazon-sqs")
	if err != nil {
		t.Fatalf("Load(amazon-sqs): %v", err)
	}
	if bundle.CLISurface == nil {
		t.Fatal("amazon-sqs CLI surface is missing")
	}
	availability := make(map[string]string, len(bundle.CLISurface.Commands))
	for _, command := range bundle.CLISurface.Commands {
		availability[command.Path] = command.Availability
	}
	for _, path := range []string{"queue delete", "queue purge"} {
		if got := availability[path]; got != "unsupported_with_provider_evidence" {
			t.Fatalf("%s availability = %q, want evidence-backed unsupported disposition", path, got)
		}
	}
}
