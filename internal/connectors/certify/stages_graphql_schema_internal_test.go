package certify

import (
	"errors"
	"testing"
)

func TestGraphQLCertificationInventoryRejectsIncorrectProducedValueAfterSchemaConformance(t *testing.T) {
	inventory, err := graphQLCertificationInventoryFor("github")
	if err != nil {
		t.Fatalf("graphQLCertificationInventoryFor: %v", err)
	}
	if inventory.SchemaConformant != 29 || inventory.LiveRequired != 2 || inventory.FixtureBound != 274 {
		t.Fatalf("GraphQL classification = schema=%d live=%d fixture=%d, want 29/2/274", inventory.SchemaConformant, inventory.LiveRequired, inventory.FixtureBound)
	}
	if len(inventory.commands) != 305 {
		t.Fatalf("GraphQL command results = %d, want 305", len(inventory.commands))
	}
	resultCounts := map[string]int{}
	for _, result := range inventory.commands {
		if result.Result == "pass" || result.Reason == "" {
			t.Fatalf("unexecuted GraphQL command result = %+v, want a concrete non-pass result", result)
		}
		resultCounts[result.Result]++
	}
	if resultCounts["schema_conformant"] != 29 || resultCounts["pending_live"] != 2 || resultCounts["fixture_required"] != 274 {
		t.Fatalf("GraphQL command result counts = %+v, want schema=29 pending_live=2 fixture=274", resultCounts)
	}
	candidate, found := inventory.liveCandidate("graphql query viewer")
	if !found {
		t.Fatal("GraphQL live candidates are missing graphql query viewer")
	}
	passed, reason := assertDirectReadOutputAssertions(candidate.StageName, CLIResult{Envelope: map[string]any{
		"response": map[string]any{
			"viewer": map[string]any{"__typename": "User"},
		},
	}}, candidate.OutputAssertions)
	if !passed {
		t.Fatalf("produced-value assertion failed after schema conformance: %s", reason)
	}
}

func TestGraphQLCertificationDeclaredButUnexecutableStaysUnexecutable(t *testing.T) {
	rc := &runContext{
		opts: Options{Connector: "github", Full: true},
		graphQLInventory: func(string) (graphQLCertificationInventory, error) {
			return graphQLCertificationInventory{}, errors.New("source lock is unavailable")
		},
	}
	report := Report{}
	if err := stageGraphQLSchemaConformance(rc, &report); err != nil {
		t.Fatalf("stageGraphQLSchemaConformance() error = %v", err)
	}
	if report.Capabilities.GraphQL == nil || report.Capabilities.GraphQL.Result != "unexecutable" {
		t.Fatalf("GraphQL capability = %+v, want declared unexecutable result", report.Capabilities.GraphQL)
	}
	if len(report.Stages) != 1 || report.Stages[0].Status != "unexecutable" || report.Stages[0].Passed {
		t.Fatalf("GraphQL stage = %+v, want one non-passing unexecutable stage", report.Stages)
	}
}

func TestGraphQLReadOnlyCertificateSkipsFixtureBoundInventoryBoundary(t *testing.T) {
	rc := &runContext{
		opts: Options{Connector: "github", Full: true, DirectReadOnly: true},
		graphQLInventory: func(string) (graphQLCertificationInventory, error) {
			return graphQLCertificationInventory{
				SchemaConformant: 29,
				FixtureBound:     274,
			}, nil
		},
	}
	report := Report{}
	if err := stageGraphQLSchemaConformance(rc, &report); err != nil {
		t.Fatalf("stageGraphQLSchemaConformance() error = %v", err)
	}
	if len(report.Stages) != 2 || report.Stages[1].Name != "graphql_inventory_boundary" || report.Stages[1].Status != "skipped" {
		t.Fatalf("read-only GraphQL stages = %+v, want an explicit skipped inventory boundary", report.Stages)
	}
	if !allStagesPassed(report.Stages) {
		t.Fatalf("a read-only GraphQL report must remain passing when only fixture-bound mutations remain: %+v", report.Stages)
	}
}
