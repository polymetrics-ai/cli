package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// Required REST path parameters are executable command inputs. This
// repository-wide invariant prevents a surface from advertising one as
// optional and deferring the failure to path interpolation or provider I/O.
func TestRequiredRESTPathParametersAlwaysMapToRequiredCLIFlags(t *testing.T) {
	definitionsRoot := filepath.Join(repoRootForCertificationTest(t), "internal", "connectors", "defs")
	entries, err := os.ReadDir(definitionsRoot)
	if err != nil {
		t.Fatalf("read connector definitions: %v", err)
	}

	var violations []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		bundle, err := engine.Load(os.DirFS(definitionsRoot), entry.Name())
		if err != nil {
			t.Fatalf("load connector %q: %v", entry.Name(), err)
		}
		if bundle.CLISurface == nil {
			continue
		}
		operations := make(map[string]engine.OperationSpec, len(bundle.Operations))
		for _, operation := range bundle.Operations {
			operations[operation.ID] = operation
		}
		for _, command := range bundle.CLISurface.Commands {
			if defect := requiredPathFlagDefect(command, operations[command.Operation]); defect != nil {
				violations = append(violations, entry.Name()+" "+defect.Command+": "+defect.Reason)
			}
		}
	}
	if len(violations) != 0 {
		t.Fatalf("optional CLI flags mapped to required REST path parameters (%d):\n%s", len(violations), strings.Join(violations, "\n"))
	}
}

func TestCertificationSweepForGitHubIsSurfaceDerivedAndExhaustive(t *testing.T) {
	root := repoRootForCertificationTest(t)
	sweep, err := buildCertificationSweep(root, "github")
	if err != nil {
		t.Fatalf("buildCertificationSweep() error = %v", err)
	}
	if sweep.Connector != "github" {
		t.Fatalf("connector = %q, want github", sweep.Connector)
	}
	if sweep.Source != "connector declarations" {
		t.Fatalf("source = %q, want connector declarations", sweep.Source)
	}
	if sweep.DeclaredCommands != 1612 || sweep.DeclaredRows != 1616 || len(sweep.Commands) != 1616 {
		t.Fatalf("declared commands/rows/commands = %d/%d/%d, want 1612/1616/1616", sweep.DeclaredCommands, sweep.DeclaredRows, len(sweep.Commands))
	}
	bundle, err := engine.Load(os.DirFS(filepath.Join(root, "internal", "connectors", "defs")), "github")
	if err != nil {
		t.Fatalf("load github bundle: %v", err)
	}
	assertions, err := certificationSweepAssertions(&bundle)
	if err != nil {
		t.Fatalf("certificationSweepAssertions() error = %v", err)
	}
	wantGeneratedTrialCandidates := map[string]int{}
	wantActiveAssertions := 0
	commandsByPath := map[string]engine.CLICommand{}
	for _, command := range bundle.CLISurface.Commands {
		commandsByPath[command.Path] = command
	}
	for path, assertion := range assertions {
		if commandsByPath[path].Availability != "implemented" {
			continue
		}
		wantActiveAssertions++
		if assertion.Cohort != "" {
			wantGeneratedTrialCandidates[assertion.Cohort]++
		}
	}

	seen := make(map[string]bool, len(sweep.Commands))
	statusTotal := 0
	for _, command := range sweep.Commands {
		if command.Path == "" || seen[command.Path] {
			t.Fatalf("commands contain missing or duplicate path %q", command.Path)
		}
		seen[command.Path] = true
		if command.Status == "" || command.Reason == "" {
			t.Fatalf("command %q has status=%q reason=%q; want concrete non-pass accounting", command.Path, command.Status, command.Reason)
		}
		if command.Status == "pass" {
			t.Fatalf("unexecuted generated command %q was promoted to pass", command.Path)
		}
		statusTotal++
	}
	if sweep.StatusTotal != statusTotal {
		t.Fatalf("status total = %d, want %d", sweep.StatusTotal, statusTotal)
	}

	var issuesList *certificationSweepCommand
	for index := range sweep.Commands {
		if sweep.Commands[index].Path == "issue list" {
			issuesList = &sweep.Commands[index]
			break
		}
	}
	if issuesList == nil {
		t.Fatal("generated sweep is missing declared command issue list")
	}
	if issuesList.Stream != "issues" || len(issuesList.APISurface) != 1 || issuesList.APISurface[0].Method != "GET" || issuesList.APISurface[0].Path != "/repos/{owner}/{repo}/issues" {
		t.Fatalf("issue list source metadata = %#v, want stream and GET api_surface from cli_surface", issuesList)
	}
	var stateFlag *certificationSweepFlag
	for index := range issuesList.Flags {
		if issuesList.Flags[index].Name == "state" {
			stateFlag = &issuesList.Flags[index]
			break
		}
	}
	if stateFlag == nil || stateFlag.Type != "enum" || stateFlag.Required || strings.Join(stateFlag.Values, ",") != "open,closed,all" {
		t.Fatalf("issue list state flag = %#v, want declared optional enum values", stateFlag)
	}

	eligible := 0
	assertionOverlays := 0
	generatedTrialCandidates := map[string]int{}
	for _, command := range sweep.Commands {
		if command.AssertionSource != "" {
			assertionOverlays++
			if len(command.OutputAssertions) == 0 {
				t.Fatalf("assertion overlay command %q has no produced-value assertion", command.Path)
			}
		}
		if command.CertificationCohort != "" {
			if command.AssertionSource != "certification.json direct_read_candidates generated from cli_surface.json" {
				t.Fatalf("generated cohort command %q has assertion source %q", command.Path, command.AssertionSource)
			}
			generatedTrialCandidates[command.CertificationCohort]++
		}
		if command.Status == certificationSweepEligiblePendingLive {
			eligible++
			if len(command.OutputAssertions) == 0 {
				t.Fatalf("eligible command %q has no produced-value assertion", command.Path)
			}
		}
	}
	if eligible == 0 {
		t.Fatal("generated sweep has no eligible assertion-bearing commands")
	}
	if assertionOverlays != wantActiveAssertions {
		t.Fatalf("assertion overlays = %d, want %d active declaration-owned overlays", assertionOverlays, wantActiveAssertions)
	}
	if !reflect.DeepEqual(generatedTrialCandidates, wantGeneratedTrialCandidates) {
		t.Fatalf("generated trial cohort candidates = %#v, want %#v active declaration-owned candidates", generatedTrialCandidates, wantGeneratedTrialCandidates)
	}
}

func TestCertificationSweepGraphQLUsesSchemaAndLiveBoundaries(t *testing.T) {
	sweep, err := buildCertificationSweep(repoRootForCertificationTest(t), "github")
	if err != nil {
		t.Fatalf("buildCertificationSweep() error = %v", err)
	}
	counts := map[string]int{}
	for _, command := range sweep.Commands {
		if strings.HasPrefix(command.Path, "graphql ") {
			counts[command.Status]++
			if command.Reason == "" {
				t.Fatalf("GraphQL command %q has no concrete non-pass reason", command.Path)
			}
		}
	}
	if got := counts[certificationSweepSchemaConformant]; got != 29 {
		t.Fatalf("schema-conformant GraphQL commands = %d, want 29", got)
	}
	if got := counts[certificationSweepEligiblePendingLive]; got != 2 {
		t.Fatalf("live-required GraphQL commands = %d, want 2", got)
	}
	if got := counts[certificationSweepFixtureRequired]; got != 274 {
		t.Fatalf("fixture-bound GraphQL commands = %d, want 274", got)
	}
	if got := counts[certificationSweepSchemaConformant] + counts[certificationSweepEligiblePendingLive] + counts[certificationSweepFixtureRequired]; got != 305 {
		t.Fatalf("classified GraphQL commands = %d, want 305", got)
	}
}

func TestCertificationSweepSeparatesProductDefectsAndProviderRefusals(t *testing.T) {
	sweep, err := buildCertificationSweep(repoRootForCertificationTest(t), "github")
	if err != nil {
		t.Fatalf("buildCertificationSweep() error = %v", err)
	}

	if len(sweep.ProductDefects) != 0 {
		t.Fatalf("product defects = %#v, want none after required-path flag derivation", sweep.ProductDefects)
	}
	for _, refusal := range sweep.ProviderRefusals {
		found := false
		for _, command := range sweep.Commands {
			if command.Path == refusal.Command {
				found = true
				if command.Status != certificationSweepProviderRefused {
					t.Fatalf("provider refusal command %q status = %q, want %q", command.Path, command.Status, certificationSweepProviderRefused)
				}
			}
		}
		if !found {
			t.Fatalf("provider refusal command %q is missing from generated sweep", refusal.Command)
		}
	}

	missingFlag := requiredPathFlagDefect(engine.CLICommand{Path: "widgets view", Operation: "widgets_view"}, engine.OperationSpec{
		REST: &engine.RESTOperationSpec{Parameters: []engine.OperationParameter{{Name: "widget_id", In: "path", Required: true}}},
	})
	if missingFlag == nil || missingFlag.Flag != "<missing>" || missingFlag.PathParameter != "widget_id" {
		t.Fatalf("missing required path flag defect = %#v, want concrete missing widget_id flag finding", missingFlag)
	}

	if err := validateCertificationSweep(certificationSweep{
		SchemaVersion:    certificationSweepSchemaVersion,
		Connector:        "acme",
		Source:           "cli_surface.json",
		DeclaredCommands: 1,
		StatusTotal:      1,
		Commands: []certificationSweepCommand{{
			Summary: "View widget", Path: "widgets view", Intent: "direct_read", Availability: "implemented", Status: certificationSweepProviderRefused,
			Reason: "provider refused request",
		}},
	}); err == nil {
		t.Fatal("validateCertificationSweep accepted a provider refusal without a concrete provider status")
	}
	if err := validateCertificationSweep(certificationSweep{
		SchemaVersion:    certificationSweepSchemaVersion,
		Connector:        "acme",
		Source:           "cli_surface.json",
		DeclaredCommands: 1,
		StatusTotal:      1,
		Commands: []certificationSweepCommand{{
			Summary: "View widget", Path: "widgets view", Intent: "direct_read", Availability: "implemented", Status: certificationSweepStatusProductDefect,
			Reason: "required REST path parameter is missing its mapped CLI flag",
		}},
	}); err == nil {
		t.Fatal("validateCertificationSweep accepted a product-defect command without a concrete product-defect record")
	}
}

func TestCertificationSweepRetainsProviderRefusalObservationWhenTheCommandIsImplemented(t *testing.T) {
	sweep, err := buildCertificationSweep(repoRootForCertificationTest(t), "github")
	if err != nil {
		t.Fatalf("buildCertificationSweep() error = %v", err)
	}
	foundRefusal := false
	for _, refusal := range sweep.ProviderRefusals {
		if refusal.Command == "actions fork-pr-contributor-approval view" {
			foundRefusal = true
			if refusal.ProviderStatus != 422 {
				t.Fatalf("provider refusal status = %d, want 422", refusal.ProviderStatus)
			}
		}
	}
	if !foundRefusal {
		t.Fatal("implemented command lost its provider-refusal observation")
	}
	for _, command := range sweep.Commands {
		if command.Path == "actions fork-pr-contributor-approval view" && command.Status != certificationSweepProviderRefused {
			t.Fatalf("implemented command status = %q, want %q", command.Status, certificationSweepProviderRefused)
		}
	}
}

func TestCertificationSweepCommandChecksGeneratedGitHubArtifact(t *testing.T) {
	root := repoRootForCertificationTest(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"certification-sweep", root, "--connector", "github", "--check"}, &stdout, &stderr); code != 0 {
		t.Fatalf("certification-sweep --check exit=%d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
}

func TestCertificationSweepArtifactValidationRejectsUnknownFields(t *testing.T) {
	sweep, err := buildCertificationSweep(repoRootForCertificationTest(t), "github")
	if err != nil {
		t.Fatalf("buildCertificationSweep() error = %v", err)
	}
	raw, err := marshalCertificationSweep(sweep)
	if err != nil {
		t.Fatalf("marshalCertificationSweep() error = %v", err)
	}
	if err := validateCertificationSweepArtifact(raw); err != nil {
		t.Fatalf("validateCertificationSweepArtifact() error = %v", err)
	}
	if err := validateCertificationSweepArtifact([]byte(`{"schema_version":1,"unexpected":true}`)); err == nil {
		t.Fatal("validateCertificationSweepArtifact accepted an unknown field")
	}
}

func TestCertificationParityClassifierProjectsOnlyTheNormalizedTaxonomy(t *testing.T) {
	implementedChangefeed := &connectors.ChangefeedDescriptor{Status: connectors.ChangefeedStatusImplemented}
	managedDestination := &connectors.SyncTransportDescriptor{Destination: &connectors.DestinationTransportDescriptor{}}
	cases := []struct {
		name       string
		input      certificationParityInput
		wantKind   string
		wantClass  string
		wantAction string
	}{
		{
			name: "REST direct read", input: certificationParityInput{
				Command:   &engine.CLICommand{Path: "widgets view", Intent: "direct_read", Operation: "widgets_view"},
				Operation: &engine.OperationSpec{ID: "widgets_view", Kind: "rest_read"},
			}, wantKind: certificationParityKindRESTRead, wantClass: certificationParityClassDirectRead,
		},
		{
			name: "GraphQL direct read normalizes to REST read", input: certificationParityInput{
				Command:   &engine.CLICommand{Path: "graphql query", Intent: "direct_read", Operation: "graphql_query"},
				Operation: &engine.OperationSpec{ID: "graphql_query", Kind: "graphql_query"},
			}, wantKind: certificationParityKindRESTRead, wantClass: certificationParityClassDirectRead,
		},
		{
			name: "REST direct write", input: certificationParityInput{
				Command:   &engine.CLICommand{Path: "widgets update", Intent: "direct_write", Operation: "widgets_update"},
				Operation: &engine.OperationSpec{ID: "widgets_update", Kind: "rest_write", MutationClass: "update", REST: &engine.RESTOperationSpec{Method: "PATCH", Path: "/widgets/{id}"}},
			}, wantKind: certificationParityKindRESTWrite, wantClass: certificationParityClassDirectWrite, wantAction: "update",
		},
		{
			name: "operation-backed REST delete", input: certificationParityInput{
				Command:   &engine.CLICommand{Path: "widgets delete", Intent: "direct_write", Operation: "widgets_delete"},
				Operation: &engine.OperationSpec{ID: "widgets_delete", Kind: "rest_write", MutationClass: "destructive", REST: &engine.RESTOperationSpec{Method: "DELETE", Path: "/widgets/{id}"}},
			}, wantKind: certificationParityKindRESTWrite, wantClass: certificationParityClassDirectWrite, wantAction: "delete",
		},
		{
			name: "operation-backed create mutation class", input: certificationParityInput{
				Command:   &engine.CLICommand{Path: "widgets create", Intent: "direct_write", Operation: "widgets_create"},
				Operation: &engine.OperationSpec{ID: "widgets_create", Kind: "rest_write", MutationClass: "create", REST: &engine.RESTOperationSpec{Method: "POST", Path: "/widgets"}},
			}, wantKind: certificationParityKindRESTWrite, wantClass: certificationParityClassDirectWrite, wantAction: "create",
		},
		{
			name: "operation-backed PUT is upsert", input: certificationParityInput{
				Command:   &engine.CLICommand{Path: "widgets replace", Intent: "direct_write", Operation: "widgets_replace"},
				Operation: &engine.OperationSpec{ID: "widgets_replace", Kind: "rest_write", MutationClass: "admin", REST: &engine.RESTOperationSpec{Method: "PUT", Path: "/widgets/{id}"}},
			}, wantKind: certificationParityKindRESTWrite, wantClass: certificationParityClassDirectWrite, wantAction: "upsert",
		},
		{
			name: "operation-backed GraphQL mutation is custom", input: certificationParityInput{
				Command:   &engine.CLICommand{Path: "widgets archive", Intent: "direct_write", Operation: "widgets_archive"},
				Operation: &engine.OperationSpec{ID: "widgets_archive", Kind: "graphql_mutation", MutationClass: "admin"},
			}, wantKind: certificationParityKindRESTWrite, wantClass: certificationParityClassDirectWrite, wantAction: "custom",
		},
		{
			name: "ETL stream", input: certificationParityInput{
				Command: &engine.CLICommand{Path: "widgets sync", Intent: "etl", Stream: "widgets"},
				Stream:  &engine.StreamSpec{Name: "widgets"},
			}, wantKind: certificationParityKindETL, wantClass: certificationParityClassETL,
		},
		{
			name: "direct write delete action", input: certificationParityInput{
				Command: &engine.CLICommand{Path: "widgets delete", Intent: "direct_write", Write: "delete_widget"},
				Write:   &engine.WriteAction{Name: "delete_widget", Kind: "delete"},
			}, wantKind: certificationParityKindRESTWrite, wantClass: certificationParityClassDirectWrite, wantAction: "delete",
		},
		{
			name: "reverse ETL delete is direct write family", input: certificationParityInput{
				Command: &engine.CLICommand{Path: "widgets delete-through-plan", Intent: "reverse_etl", Write: "delete_widget"},
				Write:   &engine.WriteAction{Name: "delete_widget", Kind: "delete"},
			}, wantKind: certificationParityKindRESTWrite, wantClass: certificationParityClassDirectWrite, wantAction: "delete",
		},
		{
			name: "binary download", input: certificationParityInput{
				Command:   &engine.CLICommand{Path: "widgets download", Intent: "binary_download", Operation: "widgets_download"},
				Operation: &engine.OperationSpec{ID: "widgets_download", Kind: "binary_download"},
			}, wantKind: certificationParityKindBinaryDownload, wantClass: certificationParityClassBinary,
		},
		{
			name: "file upload remains binary class", input: certificationParityInput{
				Command:   &engine.CLICommand{Path: "widgets upload", Intent: "direct_write", Operation: "widgets_upload"},
				Operation: &engine.OperationSpec{ID: "widgets_upload", Kind: "file_upload"},
			}, wantKind: certificationParityKindFileUpload, wantClass: certificationParityClassBinary,
		},
		{
			name: "CDC capability", input: certificationParityInput{Capabilities: &engine.Capabilities{CDC: true}},
			wantKind: certificationParityKindCDC, wantClass: certificationParityClassETL,
		},
		{
			name: "changefeed transport", input: certificationParityInput{Changefeed: implementedChangefeed},
			wantKind: certificationParityKindChangefeed, wantClass: certificationParityClassETL,
		},
		{
			name: "planned changefeed retains its exact ETL subcontract", input: certificationParityInput{Changefeed: &connectors.ChangefeedDescriptor{Status: connectors.ChangefeedStatusPlanned}},
			wantKind: certificationParityKindChangefeed, wantClass: certificationParityClassETL,
		},
		{
			name: "managed database destination", input: certificationParityInput{Transport: managedDestination, TransportRole: certificationParityTransportDestination},
			wantKind: certificationParityKindReverseETL, wantClass: certificationParityClassReverseETL,
		},
		{
			name: "valid non-applicable command", input: certificationParityInput{Command: &engine.CLICommand{Path: "config list", Intent: "config"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := classifyCertificationParity(tc.input)
			if err != nil {
				t.Fatalf("classifyCertificationParity() error = %v", err)
			}
			if got.OperationKind != tc.wantKind || got.OpClass != tc.wantClass || got.WriteActionKind != tc.wantAction {
				t.Fatalf("classification = %#v, want kind=%q class=%q action=%q", got, tc.wantKind, tc.wantClass, tc.wantAction)
			}
		})
	}
}

func TestCertificationParityClassifierRefusesMismatchedOrUnresolvedReferences(t *testing.T) {
	cases := []struct {
		name  string
		input certificationParityInput
	}{
		{
			name: "direct read references write operation",
			input: certificationParityInput{
				Command:   &engine.CLICommand{Path: "widgets view", Intent: "direct_read", Operation: "widgets_write"},
				Operation: &engine.OperationSpec{ID: "widgets_write", Kind: "rest_write"},
			},
		},
		{
			name:  "ETL command lacks stream",
			input: certificationParityInput{Command: &engine.CLICommand{Path: "widgets sync", Intent: "etl", Stream: "widgets"}},
		},
		{
			name:  "reverse ETL command lacks declared write action",
			input: certificationParityInput{Command: &engine.CLICommand{Path: "widgets delete", Intent: "reverse_etl", Write: "delete_widget"}},
		},
		{
			name:  "source transport is not a managed destination",
			input: certificationParityInput{Transport: &connectors.SyncTransportDescriptor{Source: &connectors.SourceTransportDescriptor{}}, TransportRole: certificationParityTransportDestination},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := classifyCertificationParity(tc.input); err == nil {
				t.Fatal("classifyCertificationParity() error = nil, want invalid projection error")
			}
		})
	}
}

func TestCertificationParityClassifierRefusesUndeterminableOperationWriteActionKind(t *testing.T) {
	command := engine.CLICommand{Path: "widgets rotate", Intent: "direct_write", Availability: "implemented", Operation: "widgets_rotate"}
	operation := engine.OperationSpec{ID: "widgets_rotate", Kind: "rest_write", MutationClass: "secret", REST: &engine.RESTOperationSpec{Method: "POST", Path: "/widgets/{id}/rotate"}}
	projection, err := classifyCertificationParity(certificationParityInput{
		Command: &command, Operation: &operation,
	})
	if err == nil {
		t.Fatal("classifyCertificationParity() error = nil, want indeterminate operation write action kind error")
	}
	if !strings.Contains(err.Error(), "cannot determine write action kind") {
		t.Fatalf("classifyCertificationParity() error = %q, want indeterminate action-kind context", err)
	}
	row, defect := classifyCertificationSweepCommand(command, operation, projection, err, certificationSweepAssertionOverlay{}, certificationSweepGraphQLProfile{}, certificationSweepProviderRefusal{})
	if row.Status != certificationSweepStatusProductDefect || defect == nil {
		t.Fatalf("indeterminate action-kind row = status=%q defect=%#v, want product defect", row.Status, defect)
	}
}

func TestCertificationSweepProjectsGitHubAndLegacyAPIReads(t *testing.T) {
	root := repoRootForCertificationTest(t)
	sweep, err := buildCertificationSweep(root, "github")
	if err != nil {
		t.Fatalf("buildCertificationSweep(github) error = %v", err)
	}
	readRows := 0
	for _, command := range sweep.Commands {
		if command.Intent != "direct_read" || command.Availability != "implemented" {
			continue
		}
		readRows++
		if command.OperationKind == nil || *command.OperationKind != certificationParityKindRESTRead || command.OpClass == nil || *command.OpClass != certificationParityClassDirectRead {
			t.Fatalf("implemented read %q projection = kind=%v class=%v, want rest_read/direct_read", command.Path, command.OperationKind, command.OpClass)
		}
	}
	if readRows == 0 {
		t.Fatal("GitHub has no implemented direct-read commands")
	}

	for _, connector := range []string{"zoom", "gitlab"} {
		t.Run(connector, func(t *testing.T) {
			bundle, err := engine.Load(defs.FS, connector)
			if err != nil {
				t.Fatalf("load %s: %v", connector, err)
			}
			if !bundle.Metadata.Capabilities.Read {
				t.Fatalf("%s no longer declares capability:read", connector)
			}
			got, err := classifyCertificationParity(certificationParityInput{Capabilities: &bundle.Metadata.Capabilities, Capability: "read"})
			if err != nil {
				t.Fatalf("classify capability:read: %v", err)
			}
			if got.OperationKind != certificationParityKindRESTRead || got.OpClass != certificationParityClassDirectRead {
				t.Fatalf("capability:read projection = %#v, want rest_read/direct_read", got)
			}
		})
	}
}

func TestCertificationSweepCarriesDeclaredDeleteActionKind(t *testing.T) {
	sweep, err := buildCertificationSweep(repoRootForCertificationTest(t), "github")
	if err != nil {
		t.Fatalf("buildCertificationSweep(github) error = %v", err)
	}
	for _, command := range sweep.Commands {
		if command.WriteActionKind == "delete" {
			if command.OperationKind == nil || *command.OperationKind != certificationParityKindRESTWrite || command.OpClass == nil || *command.OpClass != certificationParityClassDirectWrite {
				t.Fatalf("delete command %q projection = %#v, want rest_write/direct_write with delete action", command.Path, command)
			}
			return
		}
	}
	t.Fatal("generated GitHub sweep has no independently classified delete write action")
}

func TestCertificationParityClassifierTreatsManagedDatabaseDestinationAsReverseETL(t *testing.T) {
	postgres, err := engine.Load(defs.FS, "postgres")
	if err != nil {
		t.Fatalf("load postgres: %v", err)
	}
	mysql, err := engine.Load(defs.FS, "mysql")
	if err != nil {
		t.Fatalf("load mysql: %v", err)
	}
	for _, tc := range []struct {
		name      string
		bundle    *engine.Bundle
		transport *connectors.SyncTransportDescriptor
	}{
		{name: "PostgreSQL declared managed destination", bundle: &postgres, transport: postgres.SyncTransport},
		// This base SHA has no MySQL sync_transport.json. The classifier must
		// nevertheless preserve the same real managed-destination route when
		// that descriptor arrives, rather than consulting generic Write.
		{name: "MySQL managed destination contract", bundle: &mysql, transport: &connectors.SyncTransportDescriptor{Destination: &connectors.DestinationTransportDescriptor{}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.bundle.Metadata.Capabilities.Write {
				t.Fatalf("%s metadata must keep generic Connector.Write unavailable", tc.name)
			}
			got, err := classifyCertificationParity(certificationParityInput{
				Transport: tc.transport, TransportRole: certificationParityTransportDestination,
			})
			if err != nil {
				t.Fatalf("classify managed destination: %v", err)
			}
			if got.OperationKind != certificationParityKindReverseETL || got.OpClass != certificationParityClassReverseETL {
				t.Fatalf("managed destination projection = %#v, want reverse_etl/reverse_etl despite generic write=false", got)
			}
		})
	}
}

func TestCertificationSweepEmitsDeclaredManagedDestinationProjection(t *testing.T) {
	sweep, err := buildCertificationSweep(repoRootForCertificationTest(t), "postgres")
	if err != nil {
		t.Fatalf("buildCertificationSweep(postgres) error = %v", err)
	}
	for _, row := range sweep.Commands {
		if row.Path != "transport destination" {
			continue
		}
		if row.OperationKind == nil || *row.OperationKind != certificationParityKindReverseETL || row.OpClass == nil || *row.OpClass != certificationParityClassReverseETL {
			t.Fatalf("managed destination sweep projection = %#v, want reverse_etl/reverse_etl", row)
		}
		if row.Availability != "implemented" || row.Status != certificationSweepFixtureRequired {
			t.Fatalf("managed destination sweep accounting = availability=%q status=%q, want implemented/fixture_required", row.Availability, row.Status)
		}
		return
	}
	t.Fatal("PostgreSQL managed destination transport is absent from the generated sweep")
}

func TestCertificationSweepProjectsDeclaredAPIReadCapabilities(t *testing.T) {
	for _, connector := range []string{"zoom", "gitlab"} {
		t.Run(connector, func(t *testing.T) {
			sweep, err := buildCertificationSweep(repoRootForCertificationTest(t), connector)
			if err != nil {
				t.Fatalf("buildCertificationSweep(%s) error = %v", connector, err)
			}
			found := false
			for _, row := range sweep.Commands {
				if row.Path != "capability read" || row.Intent != "capability_read" || row.Availability != "implemented" {
					continue
				}
				found = true
				if row.OperationKind == nil || *row.OperationKind != certificationParityKindRESTRead || row.OpClass == nil || *row.OpClass != certificationParityClassDirectRead {
					t.Fatalf("declared read %q projection = kind=%v class=%v, want rest_read/direct_read", row.Path, row.OperationKind, row.OpClass)
				}
			}
			if !found {
				t.Fatal("connector has no declared implemented read capability in its generated sweep")
			}
		})
	}
}
