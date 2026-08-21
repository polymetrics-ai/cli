package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
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
	byCommand := map[string]engine.CertificationMutationCandidate{}
	for _, candidate := range generated {
		if _, duplicate := byCommand[candidate.Command]; duplicate {
			t.Fatalf("generated mutation candidate duplicates command %q", candidate.Command)
		}
		byCommand[candidate.Command] = candidate
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
	if got, want := len(generated), 906; got != want {
		t.Fatalf("generated mutation candidates = %d, want %d", got, want)
	}
	if got, want := byIntent, map[string]int{"direct_write": 283, "reverse_etl": 623}; !reflect.DeepEqual(got, want) {
		t.Fatalf("generated candidates by intent = %#v, want %#v", got, want)
	}
	userDraft, ok := byCommand["projects create-draft-item-for-authenticated-user"]
	if !ok {
		t.Fatal("generated mutation candidates omit the authenticated-user project draft command")
	}
	if userDraft.Intent != "direct_write" || userDraft.Declaration.Kind != "operation" ||
		userDraft.Declaration.ID != "github.graphql.mutation.add-project-v2-draft-issue" ||
		userDraft.Declaration.Executor != "direct_write" || userDraft.Address.Transport != "graphql" ||
		userDraft.Address.Method != "POST" || userDraft.Address.Path != "/graphql" ||
		userDraft.Fixture.Strategy != "named_exception" || userDraft.Fixture.ExceptionCode != "graphql_transport_not_collection" {
		t.Fatalf("authenticated-user project draft candidate = %#v, want the fixed GraphQL direct-write operation", userDraft)
	}
	if total := sumMutationClassifications(byClassification); total != len(generated) {
		t.Fatalf("mutation classification buckets total = %d, want %d", total, len(generated))
	}
	if got, want := byFixtureStrategy, map[string]int{"derived_collection_cycle": 528, "named_exception": 378}; !reflect.DeepEqual(got, want) {
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

// githubFoundationMutationDelta is the frozen source-identity crosswalk for
// the 32 commands added by the current-foundations Group-1 projection. The
// cohort count is evidence only after every entry below proves it is a unique,
// closed, bounded declaration backed by the immutable GitHub REST artifact.
var githubFoundationMutationDelta = map[string]string{
	"api agents set-selected-repos-for-org-secret":   "agents/set-selected-repos-for-org-secret",
	"api agents set-selected-repos-for-org-variable": "agents/set-selected-repos-for-org-variable",
	"api agents update-org-variable":                 "agents/update-org-variable",
	"api code-scanning update-alert":                 "code-scanning/update-alert",
	"api dependabot update-alert":                    "dependabot/update-alert",
	"api git create-ref":                             "git/create-ref",
	"api git update-ref":                             "git/update-ref",
	"api issues add-assignees":                       "issues/add-assignees",
	"api issues add-labels":                          "issues/add-labels",
	"api issues create-milestone":                    "issues/create-milestone",
	"api issues remove-assignees":                    "issues/remove-assignees",
	"api issues set-labels":                          "issues/set-labels",
	"api issues update-comment":                      "issues/update-comment",
	"api issues update-milestone":                    "issues/update-milestone",
	"api pulls create-review-comment":                "pulls/create-review-comment",
	"api pulls dismiss-review":                       "pulls/dismiss-review",
	"api pulls request-reviewers":                    "pulls/request-reviewers",
	"api pulls submit-review":                        "pulls/submit-review",
	"api pulls update-review-comment":                "pulls/update-review-comment",
	"api repos add-collaborator":                     "repos/add-collaborator",
	"api repos create-commit-comment":                "repos/create-commit-comment",
	"api repos create-deployment":                    "repos/create-deployment",
	"api repos create-or-update-environment":         "repos/create-or-update-environment",
	"api repos create-or-update-file-contents":       "repos/create-or-update-file-contents",
	"api repos create-webhook":                       "repos/create-webhook",
	"api repos delete-file":                          "repos/delete-file",
	"api repos merge":                                "repos/merge",
	"api repos replace-all-topics":                   "repos/replace-all-topics",
	"api repos update-commit-comment":                "repos/update-commit-comment",
	"api repos update-release-asset":                 "repos/update-release-asset",
	"api repos update-webhook":                       "repos/update-webhook",
	"api secret-scanning update-alert":               "secret-scanning/update-alert",
}

func TestGitHubFoundationMutationDeltaHasUniqueClosedBoundedSourceCrosswalk(t *testing.T) {
	const (
		pinnedRESTSHA   = "80850db290cde4eb487e0efb587cf27f305e77b6bef96933ed8a09b5169d5b1d"
		pinnedRESTBytes = 12920264
		pinnedRESTURL   = "https://raw.githubusercontent.com/github/rest-api-description/b26c240ded1c8b79cb0fb09dee4a21239061fa23/descriptions/api.github.com/api.github.com.json"
	)
	if got := len(githubFoundationMutationDelta); got != 32 {
		t.Fatalf("frozen Group-1 mutation delta = %d, want exactly 32 entries", got)
	}

	bundle, descriptor := loadInstalledGitHubSourceProjection(t)
	candidates := mutationCandidatesByCommand(t, committedMutationCandidates(t))
	commands := map[string]engine.CLICommand{}
	for _, command := range bundle.CLISurface.Commands {
		commands[command.Path] = command
	}
	actions := map[string]engine.WriteAction{}
	for _, action := range bundle.Writes {
		actions[action.Name] = action
	}
	sources := map[string]sourceOperationDescriptor{}
	for _, source := range descriptor.Operations {
		sources[source.SourceID] = source
	}

	seenWrites := map[string]bool{}
	seenSources := map[string]bool{}
	for commandPath, sourceID := range githubFoundationMutationDelta {
		candidate, found := candidates[commandPath]
		if !found {
			t.Fatalf("frozen Group-1 command %q is absent from the generated candidate artifact", commandPath)
		}
		if candidate.Cohort != "fixture_required_mutations" || candidate.Intent != "reverse_etl" || !candidate.Generated || candidate.Declaration.Kind != "write_action" || candidate.Declaration.Executor != "reverse_plan" {
			t.Fatalf("candidate %q is not the generated reverse-ETL cohort entry: %#v", commandPath, candidate)
		}
		command, found := commands[commandPath]
		if !found || command.Availability != "implemented" || command.Intent != "reverse_etl" || command.Write != candidate.Declaration.ID {
			t.Fatalf("candidate %q does not resolve to one implemented declaration-owned command: candidate=%#v command=%#v", commandPath, candidate, command)
		}
		action, found := actions[candidate.Declaration.ID]
		if !found || seenWrites[action.Name] {
			t.Fatalf("candidate %q has missing or aliased write action %q", commandPath, candidate.Declaration.ID)
		}
		seenWrites[action.Name] = true
		source, found := sources[sourceID]
		if !found || seenSources[sourceID] {
			t.Fatalf("candidate %q has missing or duplicated immutable source identity %q", commandPath, sourceID)
		}
		seenSources[sourceID] = true
		if !strings.EqualFold(source.Method, candidate.Address.Method) || source.Path != candidate.Address.Path || source.Source.URL != pinnedRESTURL || source.Source.SHA256 != pinnedRESTSHA || source.Source.Bytes != pinnedRESTBytes {
			t.Fatalf("candidate %q source crosswalk drift: candidate=%#v source=%#v", commandPath, candidate.Address, source)
		}

		var recordSchema map[string]any
		if err := json.Unmarshal(action.RecordSchema, &recordSchema); err != nil {
			t.Fatalf("candidate %q write action %q record schema: %v", commandPath, action.Name, err)
		}
		additionalProperties, declared := recordSchema["additionalProperties"].(bool)
		if !declared || additionalProperties != false {
			t.Fatalf("candidate %q write action %q must declare additionalProperties false: %s", commandPath, action.Name, action.RecordSchema)
		}
		properties, _ := recordSchema["properties"].(map[string]any)
		flags := map[string]engine.CLIFlag{}
		for _, flag := range command.Flags {
			field, recordMapped := strings.CutPrefix(flag.MapsTo, "record.")
			if !recordMapped {
				continue
			}
			if field == "" || properties[field] == nil {
				t.Fatalf("candidate %q has a CLI flag outside its closed record declaration: %#v", commandPath, flag)
			}
			if err := githubFoundationBoundedRecordFlag(flag, properties[field]); err != nil {
				t.Fatalf("candidate %q field %q is not runner-bounded: %v", commandPath, field, err)
			}
			flags[field] = flag
		}
		for field := range properties {
			if _, found := flags[field]; !found {
				t.Fatalf("candidate %q declaration-owned field %q is unreachable from the command surface", commandPath, field)
			}
		}
		for _, rawRequired := range githubFoundationRecordSchemaRequired(recordSchema) {
			if flag, found := flags[rawRequired]; !found || !flag.Required {
				t.Fatalf("candidate %q required record field %q is not a required CLI flag", commandPath, rawRequired)
			}
		}
	}
	if len(seenWrites) != 32 || len(seenSources) != 32 {
		t.Fatalf("frozen Group-1 mutation delta is not one-to-one: writes=%d sources=%d", len(seenWrites), len(seenSources))
	}
}

func githubFoundationRecordSchemaRequired(recordSchema map[string]any) []string {
	rawRequired, _ := recordSchema["required"].([]any)
	required := make([]string, 0, len(rawRequired))
	for _, raw := range rawRequired {
		if field, ok := raw.(string); ok {
			required = append(required, field)
		}
	}
	sort.Strings(required)
	return required
}

func githubFoundationBoundedRecordFlag(flag engine.CLIFlag, schema any) error {
	switch flag.Type {
	case "boolean", "integer", "number":
		// These flags are parsed into a fixed Go scalar before the write
		// validator runs, so no arbitrary byte sequence reaches the record.
		return nil
	case "enum":
		if len(flag.Values) == 0 {
			return fmt.Errorf("enum has no finite declared values")
		}
		return nil
	case "string":
		if flag.MaxBytes <= 0 {
			return fmt.Errorf("string flag has no runner byte cap")
		}
		return githubFoundationFiniteJSONShape(schema, false)
	case "json":
		if flag.MaxBytes <= 0 || flag.MaxBytes > sourceProjectionDefaultJSONBytes {
			return fmt.Errorf("JSON flag byte cap = %d, want 1..%d", flag.MaxBytes, sourceProjectionDefaultJSONBytes)
		}
		return githubFoundationFiniteJSONShape(schema, false)
	case "string_array":
		if flag.MaxBytes <= 0 || flag.MaxItems <= 0 {
			return fmt.Errorf("string-array flag requires byte and item caps")
		}
		return githubFoundationFiniteJSONShape(schema, false)
	default:
		return fmt.Errorf("unsupported record flag type %q", flag.Type)
	}
}

func githubFoundationFiniteJSONShape(raw any, allowUntyped bool) error {
	schema, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("schema is %T, want object", raw)
	}
	types := githubFoundationSchemaTypes(schema)
	if len(types) == 0 {
		if allowUntyped {
			return nil
		}
		return fmt.Errorf("schema has no type")
	}
	for _, typeName := range types {
		switch typeName {
		case "null", "boolean", "integer", "number":
			// Parsed scalar values are finite at the runner boundary.
		case "string":
			if !githubFoundationPositiveNumber(schema["maxLength"]) {
				return fmt.Errorf("string schema has no maxLength")
			}
		case "array":
			if !githubFoundationPositiveNumber(schema["maxItems"]) {
				return fmt.Errorf("array schema has no maxItems")
			}
			items, exists := schema["items"]
			if !exists {
				return fmt.Errorf("array schema has no items declaration")
			}
			if err := githubFoundationFiniteJSONShape(items, true); err != nil {
				return fmt.Errorf("array items: %w", err)
			}
		case "object":
			additional, declared := schema["additionalProperties"].(bool)
			switch {
			case declared && additional:
				if !githubFoundationPositiveNumber(schema["maxProperties"]) {
					return fmt.Errorf("dynamic object schema has no maxProperties")
				}
			case declared && !additional:
				properties, _ := schema["properties"].(map[string]any)
				for _, name := range sortedAnyMapKeys(properties) {
					if err := githubFoundationFiniteJSONShape(properties[name], false); err != nil {
						return fmt.Errorf("property %q: %w", name, err)
					}
				}
			default:
				return fmt.Errorf("object schema does not declare a closed root or maxProperties")
			}
		default:
			return fmt.Errorf("unknown schema type %q", typeName)
		}
	}
	return nil
}

func githubFoundationSchemaTypes(schema map[string]any) []string {
	rawTypes, exists := schema["type"]
	if !exists {
		return nil
	}
	var types []string
	switch typed := rawTypes.(type) {
	case string:
		types = append(types, typed)
	case []any:
		for _, rawType := range typed {
			if typeName, ok := rawType.(string); ok {
				types = append(types, typeName)
			}
		}
	}
	sort.Strings(types)
	return slices.Compact(types)
}

func githubFoundationPositiveNumber(raw any) bool {
	switch value := raw.(type) {
	case float64:
		return value > 0 && value == math.Trunc(value)
	case json.Number:
		integer, err := value.Int64()
		return err == nil && integer > 0
	case int:
		return value > 0
	case int64:
		return value > 0
	default:
		return false
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
