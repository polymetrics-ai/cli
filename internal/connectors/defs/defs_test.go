package defs

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestProductionEmbedLoadsRuntimeBundles(t *testing.T) {
	bundles, err := engine.LoadAll(FS)
	if err != nil {
		t.Fatalf("LoadAll(FS): %v", err)
	}
	if len(bundles) == 0 {
		t.Fatal("LoadAll(FS) returned zero bundles")
	}

	var github *engine.Bundle
	for i := range bundles {
		if bundles[i].Name == "github" {
			github = &bundles[i]
			break
		}
	}
	if github == nil {
		t.Fatal("LoadAll(FS) missing github bundle")
	}
	if github.Metadata.Name != "github" {
		t.Fatalf("github metadata name = %q", github.Metadata.Name)
	}
	if len(github.Streams) == 0 {
		t.Fatal("github bundle has zero streams")
	}
	if github.Docs == "" {
		t.Fatal("github bundle docs are empty")
	}
	if github.Surface != nil {
		t.Fatal("production embed should not include api_surface.json")
	}
	if github.Fixtures != nil {
		t.Fatal("production embed should not include fixtures")
	}
}

func TestProductionEmbedExcludesConformanceArtifacts(t *testing.T) {
	for _, path := range []string{"github/api_surface.json", "github/fixtures"} {
		if _, err := fs.Stat(FS, path); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("fs.Stat(%q) err = %v, want fs.ErrNotExist", path, err)
		}
	}
}

func TestLinearWriteRequiredFieldsAreNonNullable(t *testing.T) {
	linear := mustProductionBundle(t, "linear")

	for _, action := range linear.Writes {
		if action.GraphQL == nil || len(action.RecordSchema) == 0 {
			continue
		}
		var schema struct {
			Required   []string `json:"required"`
			Properties map[string]struct {
				Type any `json:"type"`
			} `json:"properties"`
		}
		if err := json.Unmarshal(action.RecordSchema, &schema); err != nil {
			t.Fatalf("%s record_schema: %v", action.Name, err)
		}
		for _, field := range schema.Required {
			property, ok := schema.Properties[field]
			if !ok {
				t.Fatalf("%s required field %q missing from properties", action.Name, field)
			}
			switch typ := property.Type.(type) {
			case string:
				if typ == "null" {
					t.Fatalf("%s required field %q is nullable", action.Name, field)
				}
			case []any:
				for _, entry := range typ {
					if entry == "null" {
						t.Fatalf("%s required field %q is nullable", action.Name, field)
					}
				}
			default:
				t.Fatalf("%s required field %q has unsupported type shape %T", action.Name, field, property.Type)
			}
		}
	}
}

func TestLinearStateReducingWritesRequireDestructiveConfirmation(t *testing.T) {
	linear := mustProductionBundle(t, "linear")

	for _, action := range linear.Writes {
		if !linearStateReducingWrite(action.Name) {
			continue
		}
		if action.Confirm != "destructive" {
			t.Fatalf("%s confirm = %q, want destructive", action.Name, action.Confirm)
		}
		if !strings.Contains(action.Risk, "destructive") {
			t.Fatalf("%s risk = %q, want destructive risk text", action.Name, action.Risk)
		}
	}
}

func TestLinearCustomerMergeRequiresDestructiveConfirmation(t *testing.T) {
	linear := mustProductionBundle(t, "linear")

	var action *engine.WriteAction
	for i := range linear.Writes {
		if linear.Writes[i].Name == "customer_merge" {
			action = &linear.Writes[i]
			break
		}
	}
	if action == nil {
		t.Fatal("linear customer_merge action missing")
	}
	if action.Confirm != "destructive" {
		t.Fatalf("customer_merge confirm = %q, want destructive", action.Confirm)
	}
	for _, want := range []string{"destructive", "archives the source customer", "typed confirmation"} {
		if !strings.Contains(action.Risk, want) {
			t.Fatalf("customer_merge risk = %q, want %q", action.Risk, want)
		}
	}

	command := mustLinearCLICommand(t, "customer_merge")
	if !strings.Contains(command.Approval, "typed destructive confirmation") {
		t.Fatalf("customer_merge CLI approval = %q, want typed destructive confirmation", command.Approval)
	}
	if !strings.Contains(command.Risk, "archives the source customer") {
		t.Fatalf("customer_merge CLI risk = %q, want source customer archive warning", command.Risk)
	}
}

func TestLinearProviderInternalMutationsAreBlocked(t *testing.T) {
	internalMutations := map[string]string{
		"file_upload_dangerously_delete":          "/graphql#Mutation.fileUploadDangerouslyDelete",
		"integration_salesforce_metadata_refresh": "/graphql#Mutation.integrationSalesforceMetadataRefresh",
		"issue_description_update_from_front":     "/graphql#Mutation.issueDescriptionUpdateFromFront",
		"organization_domain_claim":               "/graphql#Mutation.organizationDomainClaim",
		"passkey_login_start":                     "/graphql#Mutation.passkeyLoginStart",
		"project_reassign_status":                 "/graphql#Mutation.projectReassignStatus",
	}

	linear := mustProductionBundle(t, "linear")
	for _, action := range linear.Writes {
		if _, ok := internalMutations[action.Name]; ok {
			t.Fatalf("provider-internal mutation %s exposed as write action", action.Name)
		}
	}

	commands := mustLinearCLICommands(t)
	for _, command := range commands {
		if _, ok := internalMutations[command.Write]; ok {
			t.Fatalf("provider-internal mutation %s exposed as CLI command %q", command.Write, command.Path)
		}
	}

	surface := mustLinearAPISurface(t)
	for name, path := range internalMutations {
		var endpoint *engine.SurfaceEndpoint
		for i := range surface.Endpoints {
			if surface.Endpoints[i].Path == path {
				endpoint = &surface.Endpoints[i]
				break
			}
		}
		if endpoint == nil {
			t.Fatalf("api_surface row missing for %s", name)
		}
		if endpoint.CoveredBy != nil {
			t.Fatalf("%s covered_by = %+v, want blocked operation", name, endpoint.CoveredBy)
		}
		if endpoint.Operation == nil {
			t.Fatalf("%s operation = nil, want blocked operation", name)
		}
		operation := endpoint.Operation
		if operation.Model != "disallowed" || operation.Status != "blocked" || !operation.BlockedByDefault {
			t.Fatalf("%s operation = %+v, want disallowed blocked by default", name, operation)
		}
		if !strings.Contains(operation.Reason, "[INTERNAL]") || !strings.Contains(operation.Reason, "provider-internal") {
			t.Fatalf("%s reason = %q, want provider-internal evidence", name, operation.Reason)
		}
		if !strings.Contains(operation.SourceURL, "packages/sdk/src/schema.graphql") || !strings.Contains(operation.Notes, "[INTERNAL]") {
			t.Fatalf("%s source evidence = source_url %q notes %q", name, operation.SourceURL, operation.Notes)
		}
	}
}

type linearCLICommand struct {
	Path     string `json:"path"`
	Write    string `json:"write"`
	Risk     string `json:"risk"`
	Approval string `json:"approval"`
}

func mustLinearCLICommand(t *testing.T, write string) linearCLICommand {
	t.Helper()

	for _, command := range mustLinearCLICommands(t) {
		if command.Write == write {
			return command
		}
	}
	t.Fatalf("linear CLI command for write %s missing", write)
	return linearCLICommand{}
}

func mustLinearCLICommands(t *testing.T) []linearCLICommand {
	t.Helper()

	raw, err := os.ReadFile("linear/cli_surface.json")
	if err != nil {
		t.Fatalf("read linear cli_surface.json: %v", err)
	}
	var surface struct {
		Commands []linearCLICommand `json:"commands"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("parse linear cli_surface.json: %v", err)
	}
	return surface.Commands
}

func mustLinearAPISurface(t *testing.T) engine.APISurface {
	t.Helper()

	raw, err := os.ReadFile("linear/api_surface.json")
	if err != nil {
		t.Fatalf("read linear api_surface.json: %v", err)
	}
	var surface engine.APISurface
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("parse linear api_surface.json: %v", err)
	}
	return surface
}

func mustProductionBundle(t *testing.T, name string) *engine.Bundle {
	t.Helper()

	bundles, err := engine.LoadAll(FS)
	if err != nil {
		t.Fatalf("LoadAll(FS): %v", err)
	}
	for i := range bundles {
		if bundles[i].Name == name {
			return &bundles[i]
		}
	}
	t.Fatalf("LoadAll(FS) missing %s bundle", name)
	return nil
}

func linearStateReducingWrite(name string) bool {
	for _, token := range []string{
		"retire",
		"disable",
		"unlink",
		"revoke",
		"delete",
		"archive",
		"remove",
		"rotate",
		"cancel",
		"disconnect",
		"unsync",
		"logout",
		"leave",
	} {
		if name == token || strings.HasPrefix(name, token+"_") || strings.HasSuffix(name, "_"+token) || strings.Contains(name, "_"+token+"_") {
			return true
		}
	}
	return false
}
