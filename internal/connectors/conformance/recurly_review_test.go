package conformance

import (
	"os"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestRecurlyReviewBundleConformance(t *testing.T) {
	bundle, err := engine.Load(os.DirFS("../defs"), "recurly")
	if err != nil {
		t.Fatalf("load Recurly bundle: %v", err)
	}
	report := RunBundle(bundle)
	if !report.Passed {
		t.Fatalf("Recurly conformance failed: %+v", report.Checks)
	}
}
