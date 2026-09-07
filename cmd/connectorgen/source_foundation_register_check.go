package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// Decode the complete authoring report before comparison. Membership is
// checked against retained source inputs, never against the report's counts.
func validateSourceFoundationRegister(ctx context.Context, repo string, raw []byte, inputs sourceFoundationDemandInputs) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var candidate sourceFoundationRegister
	if decodeSourceJSON(raw, &candidate) != nil || decodeStrictJSON(raw, &candidate) != nil ||
		candidate.SchemaVersion != 1 || candidate.Kind != "foundation_demand_register" {
		return fmt.Errorf("foundation register closed document invalid")
	}
	if err := validateSourceFoundationCoverage(candidate.Coverage, inputs.assessments); err != nil {
		return err
	}
	// Reconstruct from the retained source/assessment/baseline and actual
	// confined proof/example readers. The candidate supplies none of these
	// authorities; an internally consistent forged report cannot certify itself.
	expected, err := buildSourceFoundationRegister(ctx, repo, inputs)
	if err != nil {
		return err
	}
	want, err := json.Marshal(expected)
	if err != nil {
		return err
	}
	// Compare the original wire object, not a remarshaled Go struct: decoding
	// null/absent scalars into zero values must not erase required-field drift.
	got, err := canonicalSourceJSON(raw)
	if err != nil {
		return err
	}
	want, err = canonicalSourceJSON(want)
	if err != nil {
		return err
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("foundation register differs from independent retained inputs")
	}
	return nil
}
