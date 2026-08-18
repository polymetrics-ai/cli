package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
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

func TestBuildGeneratedMutationCandidatesDerivesDeclaredExecutionContract(t *testing.T) {
	commands := []engine.CLICommand{
		{
			Path:         "widgets archive",
			Intent:       "direct_write",
			Availability: "implemented",
			Operation:    "acme.archive_widget",
			Flags: []engine.CLIFlag{{
				Name: "input", Type: "json", Required: true, MapsTo: "body.input",
			}},
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/graphql"}},
		},
		{
			Path:         "widgets add-seat",
			Intent:       "reverse_etl",
			Availability: "implemented",
			Write:        "add_widget_seat",
			Flags: []engine.CLIFlag{{
				Name: "org", Type: "string", Required: true, MapsTo: "record.org",
			}},
		},
	}
	operations := []engine.OperationSpec{{
		ID:   "acme.archive_widget",
		Kind: "graphql_mutation",
		GraphQL: &engine.GraphQLOperationSpec{VariablesSchema: json.RawMessage(`{
			"type":"object","required":["input"],"properties":{"input":{"type":"object"}}
		}`)},
	}}
	writes := []engine.WriteAction{{
		Name:   "add_widget_seat",
		Kind:   "create",
		Method: "PUT",
		Path:   "/orgs/{{ record.org }}/widgets/seats",
		RecordSchema: json.RawMessage(`{
			"type":"object","required":["org","seat"],
			"properties":{"org":{"type":"string"},"seat":{"type":"integer"}}
		}`),
	}}
	generation := engine.CertificationMutationCandidateGeneration{
		Cohort: engine.CertificationMutationCandidateCohort{
			Name: "fixture_mutations", CommandCount: 2, Intents: []string{"direct_write", "reverse_etl"},
		},
		Unassessed: engine.CertificationMutationClassification{
			Code: "unassessed", Evidence: "no connector-owned family matched",
		},
		Families: []engine.CertificationMutationClassificationFamily{
			{
				ID:             "contained_archive",
				Classification: engine.CertificationMutationClassification{Code: "contained", Evidence: "disposable widget container"},
				Operations:     []string{"acme.archive_widget"},
			},
			{
				ID:             "paid_widget_seat",
				Classification: engine.CertificationMutationClassification{Code: "real_money", Evidence: "adds a paid seat"},
				Writes:         []string{"add_widget_seat"},
			},
		},
	}

	generated, err := buildGeneratedMutationCandidates("acme", commands, operations, writes, generation)
	if err != nil {
		t.Fatalf("buildGeneratedMutationCandidates() error = %v", err)
	}
	if len(generated) != 2 {
		t.Fatalf("generated mutation candidates = %#v, want two", generated)
	}
	byCommand := mutationCandidatesByCommand(t, generated)
	archive := byCommand["widgets archive"]
	if got, want := archive.CommandTokens, []string{"widgets", "archive"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("archive command tokens = %#v, want %#v", got, want)
	}
	if archive.CredentialFlag != "--credential" {
		t.Fatalf("archive credential contract = %#v, want the source --credential flag", archive)
	}
	if archive.JSONMode == nil || !*archive.JSONMode {
		t.Fatalf("archive JSON mode = %#v, want the declared --json mode", archive.JSONMode)
	}
	if got, want := archive.Declaration, (engine.CertificationMutationDeclaration{
		Kind: "operation", ID: "acme.archive_widget", Executor: "direct_write",
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("archive declaration = %#v, want %#v", got, want)
	}
	if got, want := archive.Address, (engine.CertificationMutationAddress{
		Source: "cli_surface", Transport: "graphql", Method: "POST", Path: "/graphql",
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("archive address = %#v, want %#v", got, want)
	}
	if got := archive.Fixture; got.Strategy != "named_exception" || got.ExceptionCode != "graphql_transport_not_collection" {
		t.Fatalf("archive fixture = %#v, want explicit GraphQL collection exception", got)
	}
	if got := archive.Classification; got.Code != "contained" || got.Evidence != "disposable widget container" {
		t.Fatalf("archive classification = %#v, want contained disposable evidence", got)
	}
	if got, want := archive.RequiredFlags, []string{"--input"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("archive required flags = %#v, want %#v", got, want)
	}
	if !hasMutationSlot(archive.InputSlots, "body.input", "json", true) {
		t.Fatalf("archive input slots = %#v, want required body.input JSON slot", archive.InputSlots)
	}

	seat := byCommand["widgets add-seat"]
	if got, want := seat.Declaration, (engine.CertificationMutationDeclaration{
		Kind: "write_action", ID: "add_widget_seat", Executor: "reverse_plan",
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("seat declaration = %#v, want %#v", got, want)
	}
	if got, want := seat.Address, (engine.CertificationMutationAddress{
		Source: "write_action", Transport: "rest", Method: "PUT", Path: "/orgs/{{ record.org }}/widgets/seats",
	}); !reflect.DeepEqual(got, want) {
		t.Fatalf("seat address = %#v, want %#v", got, want)
	}
	if got := seat.Fixture; got.Strategy != "named_exception" || got.ExceptionCode != "missing_api_surface" {
		t.Fatalf("seat fixture = %#v, want explicit missing API surface exception", got)
	}
	if got := seat.Classification; got.Code != "real_money" || got.Evidence != "adds a paid seat" {
		t.Fatalf("seat classification = %#v, want real_money", got)
	}
	if !hasMutationSlot(seat.InputSlots, "record.seat", "integer", true) {
		t.Fatalf("seat input slots = %#v, want required record.seat slot from the write schema", seat.InputSlots)
	}
}

func TestBuildGeneratedMutationCandidatesClassifiesEscapeAndFailsClosed(t *testing.T) {
	commands := []engine.CLICommand{
		{Path: "widgets contained", Intent: "direct_write", Availability: "implemented", Operation: "acme.contained", APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/graphql"}}},
		{Path: "widgets paid", Intent: "direct_write", Availability: "implemented", Operation: "acme.paid", APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/graphql"}}},
		{Path: "widgets invite", Intent: "direct_write", Availability: "implemented", Operation: "acme.invite", APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/graphql"}}},
		{Path: "widgets publish", Intent: "direct_write", Availability: "implemented", Operation: "acme.publish", APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/graphql"}}},
		{Path: "widgets third-party", Intent: "direct_write", Availability: "implemented", Operation: "acme.third_party", APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/graphql"}}},
		{Path: "widgets unknown", Intent: "direct_write", Availability: "implemented", Operation: "acme.unknown", APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/graphql"}}},
	}
	operations := make([]engine.OperationSpec, 0, len(commands))
	for _, command := range commands {
		operations = append(operations, engine.OperationSpec{ID: command.Operation, Kind: "graphql_mutation", GraphQL: &engine.GraphQLOperationSpec{}})
	}
	generation := engine.CertificationMutationCandidateGeneration{
		Cohort:     engine.CertificationMutationCandidateCohort{Name: "mutation_test", CommandCount: len(commands), Intents: []string{"direct_write"}},
		Unassessed: engine.CertificationMutationClassification{Code: "unassessed", Evidence: "no connector-owned containment evidence"},
		Families: []engine.CertificationMutationClassificationFamily{
			{ID: "contained", Classification: engine.CertificationMutationClassification{Code: "contained", Evidence: "owned disposable container"}, Operations: []string{"acme.contained"}},
			{ID: "money", Classification: engine.CertificationMutationClassification{Code: "real_money", Evidence: "paid-seat mutation"}, Operations: []string{"acme.paid"}},
			{ID: "people", Classification: engine.CertificationMutationClassification{Code: "real_people", Evidence: "outside invitation"}, Operations: []string{"acme.invite"}},
			{ID: "public", Classification: engine.CertificationMutationClassification{Code: "public_visibility", Evidence: "public publication"}, Operations: []string{"acme.publish"}},
			{ID: "third-party", Classification: engine.CertificationMutationClassification{Code: "third_party_scope", Evidence: "third-party repository target"}, Operations: []string{"acme.third_party"}},
		},
	}

	generated, err := buildGeneratedMutationCandidates("acme", commands, operations, nil, generation)
	if err != nil {
		t.Fatalf("buildGeneratedMutationCandidates() error = %v", err)
	}
	got := map[string]string{}
	for _, candidate := range generated {
		got[candidate.Command] = candidate.Classification.Code
	}
	want := map[string]string{
		"widgets contained":   "contained",
		"widgets paid":        "real_money",
		"widgets invite":      "real_people",
		"widgets publish":     "public_visibility",
		"widgets third-party": "third_party_scope",
		"widgets unknown":     "unassessed",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mutation classification = %#v, want %#v", got, want)
	}
	generation.Families[0].Operations = []string{"acme.not-declared"}
	if _, err := buildGeneratedMutationCandidates("acme", commands, operations, nil, generation); err == nil {
		t.Fatal("buildGeneratedMutationCandidates() accepted a containment selector outside the declared cohort")
	}
}

func TestBuildGeneratedMutationCandidatesDerivesRESTCollectionCycleInURLDepthOrder(t *testing.T) {
	commands := []engine.CLICommand{
		{Path: "widgets create", Intent: "reverse_etl", Availability: "implemented", Write: "create_widget", APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/orgs/{org}/widgets"}}},
		{Path: "widgets update", Intent: "reverse_etl", Availability: "implemented", Write: "update_widget", APISurface: []engine.CLISurfaceEndpointRef{{Method: "PATCH", Path: "/orgs/{org}/widgets/{widget_id}"}}},
		{Path: "widget-labels create", Intent: "reverse_etl", Availability: "implemented", Write: "create_widget_label", APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/orgs/{org}/widgets/{widget_id}/labels"}}},
		{Path: "widget-labels delete", Intent: "reverse_etl", Availability: "implemented", Write: "delete_widget_label", APISurface: []engine.CLISurfaceEndpointRef{{Method: "DELETE", Path: "/orgs/{org}/widgets/{widget_id}/labels/{name}"}}},
		{Path: "orphans delete", Intent: "reverse_etl", Availability: "implemented", Write: "delete_orphan", APISurface: []engine.CLISurfaceEndpointRef{{Method: "DELETE", Path: "/orgs/{org}/orphans/{orphan_id}"}}},
	}
	writes := []engine.WriteAction{
		{Name: "create_widget", Kind: "create"},
		{Name: "update_widget", Kind: "update"},
		{Name: "create_widget_label", Kind: "create"},
		{Name: "delete_widget_label", Kind: "delete"},
		{Name: "delete_orphan", Kind: "delete"},
	}
	generation := engine.CertificationMutationCandidateGeneration{
		Cohort:     engine.CertificationMutationCandidateCohort{Name: "mutation_test", CommandCount: len(commands), Intents: []string{"reverse_etl"}},
		Unassessed: engine.CertificationMutationClassification{Code: "unassessed", Evidence: "no containment classification"},
		Families: []engine.CertificationMutationClassificationFamily{{
			ID:             "contained",
			Classification: engine.CertificationMutationClassification{Code: "contained", Evidence: "owned disposable container"},
			Intents:        []string{"reverse_etl"},
		}},
	}

	generated, err := buildGeneratedMutationCandidates("acme", commands, nil, writes, generation)
	if err != nil {
		t.Fatalf("buildGeneratedMutationCandidates() error = %v", err)
	}
	byCommand := mutationCandidatesByCommand(t, generated)
	for _, command := range []string{"widgets create", "widgets update"} {
		fixture := byCommand[command].Fixture
		if got, want := fixture.Strategy, "derived_collection_cycle"; got != want {
			t.Fatalf("%s fixture strategy = %q, want %q: %#v", command, got, want, fixture)
		}
		if got, want := fixture.Collection, "/orgs/{org}/widgets"; got != want {
			t.Fatalf("%s collection = %q, want %q", command, got, want)
		}
		if got, want := fixture.ProvisionerCommands, []string{"widgets create"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("%s provisioners = %#v, want %#v", command, got, want)
		}
	}
	for _, command := range []string{"widget-labels create", "widget-labels delete"} {
		fixture := byCommand[command].Fixture
		if got, want := fixture.Collection, "/orgs/{org}/widgets/{widget_id}/labels"; got != want {
			t.Fatalf("%s collection = %q, want %q", command, got, want)
		}
		if got, want := fixture.CollectionDepth, 5; got != want {
			t.Fatalf("%s collection depth = %d, want %d", command, got, want)
		}
	}
	if fixture := byCommand["orphans delete"].Fixture; fixture.Strategy != "named_exception" || fixture.ExceptionCode != "collection_without_creator" {
		t.Fatalf("orphan fixture = %#v, want explicit no-creator exception", fixture)
	}
	for index := 1; index < len(generated); index++ {
		previous, current := generated[index-1].Fixture, generated[index].Fixture
		if previous.Strategy == "derived_collection_cycle" && current.Strategy == "derived_collection_cycle" && previous.CollectionDepth > current.CollectionDepth {
			t.Fatalf("fixture order regressed from depth %d to %d: %#v", previous.CollectionDepth, current.CollectionDepth, generated)
		}
	}
}

func TestGitHubFixtureRequiredMutationCohortGeneratesEveryCandidate(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(defs.FS, github) error = %v", err)
	}
	generation := bundle.Certification.MutationGeneration
	if generation == nil {
		t.Fatal("github has no mutation generation declaration")
	}
	generated, err := buildGeneratedMutationCandidates("github", bundle.CLISurface.Commands, bundle.Operations, bundle.Writes, *generation)
	if err != nil {
		t.Fatalf("buildGeneratedMutationCandidates() error = %v", err)
	}
	byIntent := map[string]int{}
	byClassification := map[string]int{}
	byFixtureStrategy := map[string]int{}
	for _, candidate := range generated {
		byIntent[candidate.Intent]++
		byClassification[candidate.Classification.Code]++
		byFixtureStrategy[candidate.Fixture.Strategy]++
		if candidate.Classification.Code == "" || candidate.Classification.Evidence == "" {
			t.Fatalf("candidate %q lacks an explicit classification: %#v", candidate.Command, candidate.Classification)
		}
		if candidate.Fixture.Strategy == "" || candidate.Fixture.Evidence == "" {
			t.Fatalf("candidate %q lacks explicit fixture provenance: %#v", candidate.Command, candidate.Fixture)
		}
		if candidate.JSONMode == nil || !*candidate.JSONMode {
			t.Fatalf("candidate %q lacks the declared --json mode", candidate.Command)
		}
		if candidate.Address.Transport != "rest" && candidate.Address.Transport != "graphql" {
			t.Fatalf("candidate %q lacks a derived address transport: %#v", candidate.Command, candidate.Address)
		}
	}
	if got, want := len(generated), 865; got != want {
		t.Fatalf("generated mutation candidates = %d, want %d", got, want)
	}
	if got, want := byIntent, map[string]int{"direct_write": 282, "reverse_etl": 583}; !reflect.DeepEqual(got, want) {
		t.Fatalf("generated candidates by intent = %#v, want %#v", got, want)
	}
	if total := sumMutationClassifications(byClassification); total != len(generated) {
		t.Fatalf("mutation classification buckets total = %d, want %d", total, len(generated))
	}
	if got, want := byFixtureStrategy, map[string]int{"derived_collection_cycle": 495, "named_exception": 370}; !reflect.DeepEqual(got, want) {
		t.Fatalf("mutation fixture provenance = %#v, want %#v", got, want)
	}
	committed := committedMutationCandidates(t)
	if got := len(committed); got != len(generated) {
		t.Fatalf("committed mutation candidates = %d, want %d generated candidates", got, len(generated))
	}
	if got, want := committed, generated; !reflect.DeepEqual(got, want) {
		t.Fatal("committed mutation candidates differ from the declared-surface projection")
	}
}

func TestGitHubMutationInventoryIsNotEmbeddedInTheRuntimeBundle(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("Load(defs.FS, github) error = %v", err)
	}
	if got := len(bundle.Certification.MutationCandidates); got != 0 {
		t.Fatalf("runtime bundle mutation candidates = %d, want 0; exhaustive candidate validation belongs to connectorgen", got)
	}
	if _, err := defs.FS.ReadFile(filepath.Join("github", mutationCertificationCandidatesFile)); err == nil {
		t.Fatalf("runtime defs.FS embeds %s; generator-only mutation inventory must not affect fresh pm startup", mutationCertificationCandidatesFile)
	}
}

func committedMutationCandidates(t *testing.T) []engine.CertificationMutationCandidate {
	t.Helper()
	path := filepath.Join(repoRootForCertificationTest(t), "internal", "connectors", "defs", "github", mutationCertificationCandidatesFile)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read mutation candidate artifact: %v", err)
	}
	var artifact mutationCertificationCandidatesArtifact
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&artifact); err != nil {
		t.Fatalf("decode mutation candidate artifact: %v", err)
	}
	if artifact.SchemaVersion != 1 {
		t.Fatalf("mutation candidate artifact schema_version = %d, want 1", artifact.SchemaVersion)
	}
	return artifact.MutationCandidates
}

func mutationCandidatesByCommand(t *testing.T, candidates []engine.CertificationMutationCandidate) map[string]engine.CertificationMutationCandidate {
	t.Helper()
	byCommand := make(map[string]engine.CertificationMutationCandidate, len(candidates))
	for _, candidate := range candidates {
		if _, duplicate := byCommand[candidate.Command]; duplicate {
			t.Fatalf("duplicate mutation candidate %q", candidate.Command)
		}
		byCommand[candidate.Command] = candidate
	}
	return byCommand
}

func hasMutationSlot(slots []engine.CertificationMutationInputSlot, path, valueType string, required bool) bool {
	for _, slot := range slots {
		if slot.Path == path && slot.Type == valueType && slot.Required == required {
			return true
		}
	}
	return false
}

func sumMutationClassifications(buckets map[string]int) int {
	keys := make([]string, 0, len(buckets))
	for key := range buckets {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	total := 0
	for _, key := range keys {
		total += buckets[key]
	}
	return total
}

func TestMergeGeneratedMutationCandidatesRequiresNamedExactOverride(t *testing.T) {
	generated := []engine.CertificationMutationCandidate{
		{Command: "widgets archive", Generated: true},
		{Command: "widgets remove", Generated: true},
	}
	manual := engine.CertificationMutationCandidate{
		Command:        "widgets archive",
		OverrideReason: "the declared fixture needs a produced identifier assertion",
	}
	merged, err := mergeGeneratedMutationCandidates([]engine.CertificationMutationCandidate{manual}, generated)
	if err != nil {
		t.Fatalf("mergeGeneratedMutationCandidates() error = %v", err)
	}
	if got, want := merged, []engine.CertificationMutationCandidate{manual, generated[1]}; !reflect.DeepEqual(got, want) {
		t.Fatalf("merged mutation candidates = %#v, want %#v", got, want)
	}
	_, err = mergeGeneratedMutationCandidates([]engine.CertificationMutationCandidate{{
		Command: "widgets invented", OverrideReason: "must not create an unbounded manual candidate",
	}}, generated)
	if err == nil {
		t.Fatal("mergeGeneratedMutationCandidates() accepted a manual candidate that does not shadow an exact generated command")
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
