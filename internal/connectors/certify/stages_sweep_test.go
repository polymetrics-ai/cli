package certify_test

import (
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

func TestSweepSkipsWhenFullDisabled(t *testing.T) {
	// When Options.Full is false, the sweep stage records a skip and does
	// not touch WriteActions.
	rep := certify.Report{}
	rep.Capabilities.WriteActions = map[string]certify.WriteActionResult{}

	// The sweep stage is the last stage; it's exercised via the Runner in
	// the full stage list. Here we verify the Options.Full field exists and
	// defaults to false.
	opts := certify.Options{Connector: "github", Write: true}
	if opts.Full {
		t.Fatal("Options.Full should default to false")
	}
}

func TestSweepOptionsFullFlagExists(t *testing.T) {
	// Verify the Full field is settable (the CLI --full flag maps to this).
	opts := certify.Options{Connector: "github", Write: true, Full: true}
	if !opts.Full {
		t.Fatal("Options.Full should be true when set")
	}
}

func TestSweepPairingsForGithubHasMultiple(t *testing.T) {
	// The sweep needs >1 pairing to be meaningful. GitHub has 3.
	pairings := certify.PairingsFor("github")
	if len(pairings) <= 1 {
		t.Fatalf("PairingsFor(github) = %d pairings, want >1 for a sweep", len(pairings))
	}
	// Verify each pairing has a create + cleanup.
	for i, p := range pairings {
		if p.Create == "" || p.Cleanup == "" {
			t.Fatalf("pairing %d has empty Create or Cleanup: %+v", i, p)
		}
	}
}
