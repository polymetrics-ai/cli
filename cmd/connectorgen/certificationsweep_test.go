package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

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
	if sweep.Source != "cli_surface.json" {
		t.Fatalf("source = %q, want cli_surface.json", sweep.Source)
	}
	if sweep.DeclaredCommands != 1571 || len(sweep.Commands) != 1571 {
		t.Fatalf("declared/commands = %d/%d, want 1571/1571", sweep.DeclaredCommands, len(sweep.Commands))
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
	if assertionOverlays != 122 {
		t.Fatalf("assertion overlays = %d, want 25 hand-authored plus 97 generated overlays", assertionOverlays)
	}
	if got, want := generatedTrialCandidates, map[string]int{
		"trial_advanced_security": 31,
		"trial_codespaces":        22,
		"trial_copilot":           23,
		"trial_enterprise":        21,
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("generated trial cohort candidates = %#v, want %#v", got, want)
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
	var providerRefusal *certificationSweepProviderRefusal
	for index := range sweep.ProviderRefusals {
		if sweep.ProviderRefusals[index].Command == "actions fork-pr-contributor-approval view" {
			providerRefusal = &sweep.ProviderRefusals[index]
			break
		}
	}
	if providerRefusal == nil || providerRefusal.ProviderStatus != 422 || !strings.Contains(providerRefusal.Reason, "does not apply") {
		t.Fatalf("provider refusals = %#v, want named HTTP 422 provider refusal", sweep.ProviderRefusals)
	}
	foundProviderRefusalCommand := false
	for _, command := range sweep.Commands {
		if command.Path == providerRefusal.Command && command.Status != certificationSweepProviderRefused {
			t.Fatalf("provider refusal command status = %q, want %q", command.Status, certificationSweepProviderRefused)
		}
		if command.Path == providerRefusal.Command {
			foundProviderRefusalCommand = true
		}
	}
	if !foundProviderRefusalCommand {
		t.Fatalf("provider refusal command %q is missing from generated sweep", providerRefusal.Command)
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
