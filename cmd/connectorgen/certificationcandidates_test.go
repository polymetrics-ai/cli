package main

import (
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestBuildGeneratedReadCandidatesDerivesCommandAndProducedValueAssertion(t *testing.T) {
	commands := []engine.CLICommand{
		{
			Path:         "widgets list",
			Intent:       "direct_read",
			Availability: "implemented",
			Flags: []engine.CLIFlag{
				{Name: "org", Type: "string", Required: true},
				{Name: "state", Type: "enum", Values: []string{"open", "closed"}},
			},
		},
		{Path: "widgets create", Intent: "reverse_etl", Availability: "implemented"},
		{Path: "widgets hidden", Intent: "direct_read", Availability: "partial"},
	}

	generated, err := buildGeneratedReadCandidates("acme", commands, readCandidateGeneration{
		Cohorts: []engine.CertificationReadCandidateCohort{{
			Name:         "trial",
			CommandCount: 1,
			Commands:     []string{"widgets list"},
		}},
		RequiredFlagDefaults: map[string]string{"org": "acme-org"},
	})
	if err != nil {
		t.Fatalf("buildGeneratedReadCandidates() error = %v", err)
	}
	if len(generated) != 1 {
		t.Fatalf("generated candidates = %#v, want exactly widgets list", generated)
	}
	candidate := generated[0]
	if candidate.Command != "widgets list" || candidate.Cohort != "trial" || !candidate.Generated {
		t.Fatalf("candidate identity = %#v, want generated trial widgets list", candidate)
	}
	if got, want := candidate.Args, []engine.CertificationCommandArg{
		{Connector: true},
		{Literal: "widgets"},
		{Literal: "list"},
		{Literal: "--credential"},
		{SourceCredential: true},
		{Literal: "--org"},
		{ConfigKey: "org", Default: "acme-org"},
		{Literal: "--json"},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("candidate args = %#v, want %#v", got, want)
	}
	if got, want := candidate.OutputAssertions, []engine.CertificationOutputAssertion{{
		JSONPointer: "/response",
		ValueType:   "object_or_array",
	}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("candidate assertions = %#v, want %#v", got, want)
	}
}

func TestBuildGeneratedReadCandidatesRejectsCohortCommandOutsideDeclaredSurface(t *testing.T) {
	_, err := buildGeneratedReadCandidates("acme", []engine.CLICommand{{
		Path: "widgets list", Intent: "direct_read", Availability: "implemented",
	}}, readCandidateGeneration{Cohorts: []engine.CertificationReadCandidateCohort{{
		Name:         "trial",
		CommandCount: 1,
		Commands:     []string{"widgets missing"},
	}}})
	if err == nil {
		t.Fatal("buildGeneratedReadCandidates() accepted a cohort command absent from cli_surface")
	}
}

func TestMergeGeneratedReadCandidatesPreservesManualOverride(t *testing.T) {
	manual := engine.CertificationCommandCandidate{
		StageName: "manual_widgets_list",
		Command:   "widgets list",
		Args:      []engine.CertificationCommandArg{{Literal: "widgets"}},
		OutputAssertions: []engine.CertificationOutputAssertion{{
			JSONPointer: "/response/name",
			Equals:      "fixture-widget",
		}},
	}
	generated := []engine.CertificationCommandCandidate{
		{StageName: "generated_widgets_list", Command: "widgets list", Generated: true},
		{StageName: "generated_widgets_all", Command: "widgets all", Generated: true},
	}

	merged, err := mergeGeneratedReadCandidates([]engine.CertificationCommandCandidate{manual}, generated)
	if err != nil {
		t.Fatalf("mergeGeneratedReadCandidates() error = %v", err)
	}
	if got, want := merged, []engine.CertificationCommandCandidate{manual, generated[1]}; !reflect.DeepEqual(got, want) {
		t.Fatalf("merged candidates = %#v, want %#v", got, want)
	}
}

func TestGitHubTrialCohortIsCompleteAndGeneratesOnlyReads(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(defs.FS, github) error = %v", err)
	}
	generation := bundle.Certification.DirectReadGeneration
	if generation == nil {
		t.Fatal("github has no direct-read generation declaration")
	}
	wantCohorts := map[string]int{
		"trial_advanced_security": 50,
		"trial_codespaces":        46,
		"trial_copilot":           50,
		"trial_enterprise":        46,
	}
	gotCohorts := map[string]int{}
	total := 0
	for _, cohort := range generation.Cohorts {
		gotCohorts[cohort.Name] = cohort.CommandCount
		total += cohort.CommandCount
	}
	if got, want := gotCohorts, wantCohorts; !reflect.DeepEqual(got, want) {
		t.Fatalf("trial cohorts = %#v, want %#v", got, want)
	}
	if total != 192 {
		t.Fatalf("trial cohort command total = %d, want 192", total)
	}

	generated, err := buildGeneratedReadCandidates("github", bundle.CLISurface.Commands, readCandidateGeneration{
		RequiredFlagDefaults: generation.RequiredFlagDefaults,
		Cohorts:              generation.Cohorts,
	})
	if err != nil {
		t.Fatalf("buildGeneratedReadCandidates() error = %v", err)
	}
	if len(generated) != 97 {
		t.Fatalf("generated direct reads = %d, want 97", len(generated))
	}
}
