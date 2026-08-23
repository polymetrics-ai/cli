package conformance

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"maps"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestGitHubExhaustiveProviderDouble(t *testing.T) {
	report, err := runGitHubExhaustiveProviderDouble(t)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load GitHub bundle: %v", err)
	}
	if report.Streams != len(bundle.Streams) || report.WriteActions != len(bundle.Writes) || report.Operations != len(bundle.Operations) {
		t.Fatalf("provider-double totals = streams=%d writes=%d operations=%d, want source bundle %d/%d/%d",
			report.Streams, report.WriteActions, report.Operations, len(bundle.Streams), len(bundle.Writes), len(bundle.Operations))
	}
	if report.Failed != 0 {
		t.Errorf("provider-double report has %d failed rows: %v", report.Failed, report.Failures)
	}
	wantGenericStreams, wantGenericWrites := githubCommandlessDeclarationSets(bundle)
	gotGenericStreams, gotGenericWrites := githubReportedGenericRouteSets(report)
	if !reflect.DeepEqual(gotGenericStreams, wantGenericStreams) || !reflect.DeepEqual(gotGenericWrites, wantGenericWrites) {
		t.Errorf("generic route rows = streams=%v writes=%v, want commandless declarations streams=%v writes=%v", gotGenericStreams, gotGenericWrites, wantGenericStreams, wantGenericWrites)
	}
	for _, row := range report.Rows {
		if !strings.HasPrefix(row.Name, "github.graphql.") {
			continue
		}
		if row.State != "exercised" {
			t.Errorf("fixed GraphQL operation %q state = %q, want exercised (%s)", row.Name, row.State, row.Reason)
		}
	}
}

func githubCommandlessDeclarationSets(bundle engine.Bundle) (streams, writes []string) {
	for _, stream := range bundle.Streams {
		if !githubStreamHasCommand(bundle, stream.Name) {
			streams = append(streams, stream.Name)
		}
	}
	for _, action := range bundle.Writes {
		if !githubWriteActionHasCommand(bundle, action.Name) {
			writes = append(writes, action.Name)
		}
	}
	sort.Strings(streams)
	sort.Strings(writes)
	return streams, writes
}

func githubReportedGenericRouteSets(report githubProviderDoubleReport) (streams, writes []string) {
	for _, row := range report.Rows {
		switch row.Kind {
		case "stream":
			if strings.HasPrefix(row.Route, "pm etl generic route") {
				streams = append(streams, row.Name)
			}
		case "write_action":
			if strings.HasPrefix(row.Route, "pm reverse generic route") {
				writes = append(writes, row.Name)
			}
		}
	}
	sort.Strings(streams)
	sort.Strings(writes)
	return streams, writes
}

func TestSyntheticGitHubSecretSetRecordUsesSchemaValidCiphertext(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load GitHub bundle: %v", err)
	}
	var action *engine.WriteAction
	for index := range bundle.Writes {
		if bundle.Writes[index].Name == "actions_secrets_secret_name3" {
			action = &bundle.Writes[index]
			break
		}
	}
	if action == nil {
		t.Fatal("GitHub actions secret-set write action is absent")
	}
	record, err := syntheticGitHubRecord(*action)
	if err != nil {
		t.Fatalf("syntheticGitHubRecord() error = %v", err)
	}
	value, ok := record["encrypted_value"].(string)
	if !ok || value == "" {
		t.Fatalf("synthetic encrypted_value = %#v, want base64 ciphertext", record["encrypted_value"])
	}
	if value == "provider-double" {
		t.Fatal("synthetic secret-set value is plaintext rather than ciphertext-shaped input")
	}
	if _, err := base64.StdEncoding.DecodeString(value); err != nil {
		t.Fatalf("synthetic encrypted_value is not base64: %v", err)
	}
}

func TestSyntheticGitHubRecordsSatisfyConstraintWitnesses(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load GitHub bundle: %v", err)
	}
	want := map[string]bool{
		"dependency_graph_snapshots":                          true,
		"orgs_create_artifact_deployment_record":              true,
		"orgs_create_artifact_storage_record":                 true,
		"pull_request_stacks_create":                          true,
		"users_create_public_ssh_key_for_authenticated_user":  true,
		"users_create_ssh_signing_key_for_authenticated_user": true,
	}
	for _, action := range bundle.Writes {
		if !want[action.Name] {
			continue
		}
		record, err := syntheticGitHubRecord(action)
		if err != nil {
			t.Fatalf("syntheticGitHubRecord(%s): %v", action.Name, err)
		}
		if err := engine.ValidateWrite(context.Background(), bundle, connectors.WriteRequest{Action: action.Name}, []connectors.Record{record}); err != nil {
			t.Fatalf("ValidateWrite(%s, %#v): %v", action.Name, record, err)
		}
		delete(want, action.Name)
	}
	if len(want) != 0 {
		t.Fatalf("constraint action(s) absent from bundle: %v", want)
	}
}

func TestSyntheticGitHubRecordRejectsUnsupportedConstraints(t *testing.T) {
	if _, err := syntheticSchemaValue(json.RawMessage(`{"type":"string","pattern":"^[A-Z]{999}$"}`), "value"); err == nil {
		t.Fatal("syntheticSchemaValue accepted an unsupported pattern")
	}
	if _, err := syntheticSchemaValue(json.RawMessage(`{"type":"array","minItems":65,"items":{"type":"string"}}`), "values"); err == nil {
		t.Fatal("syntheticSchemaValue accepted an excessive minItems witness")
	}
}

func TestSyntheticGitHubGraphQLPaginationUsesExactlyOneDeclaredDirection(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load GitHub bundle: %v", err)
	}
	var operation engine.OperationSpec
	for _, candidate := range bundle.Operations {
		if candidate.ID == "github.graphql.query.search" {
			operation = candidate
			break
		}
	}
	if operation.GraphQL == nil || operation.GraphQL.Pagination == nil {
		t.Fatal("GitHub search operation has no declared bidirectional pagination")
	}
	pagination := operation.GraphQL.Pagination
	variables, err := syntheticGraphQLVariables(operation)
	if err != nil {
		t.Fatalf("syntheticGraphQLVariables() error = %v", err)
	}
	if variables[pagination.PageSizeVariable] != 1 {
		t.Fatalf("forward pagination variable %q = %#v, want 1", pagination.PageSizeVariable, variables[pagination.PageSizeVariable])
	}
	if _, both := variables[pagination.BackwardPageSizeVariable]; both {
		t.Fatalf("synthetic variables selected both pagination directions: %#v", variables)
	}

	response, err := githubGraphQLProviderDoubleResponse(operation)
	if err != nil {
		t.Fatalf("githubGraphQLProviderDoubleResponse() error = %v", err)
	}
	invoke := func(vars map[string]any) (int, error) {
		capture := newGitHubProviderCapture(func(*http.Request) (int, string, []byte) {
			return http.StatusOK, "application/json", response
		})
		defer capture.Close()
		doubleBundle := githubProviderDoubleBundle(bundle, capture.URL)
		_, err := engine.OperationDirectRead(context.Background(), doubleBundle, connectors.OperationDirectReadRequest{
			Operation: operation.ID, Config: githubProviderDoubleConfig(doubleBundle), Body: vars,
			MaxBytes: operation.GraphQL.MaxBytes, OutputPolicy: operation.OutputPolicy,
		}, engine.HooksFor(doubleBundle.Name))
		return len(capture.captured()), err
	}
	if sends, err := invoke(variables); err != nil || sends != 1 {
		t.Fatalf("forward pagination = sends %d, err %v; want one send", sends, err)
	}
	neither := maps.Clone(variables)
	delete(neither, pagination.PageSizeVariable)
	if sends, err := invoke(neither); err == nil || sends != 0 {
		t.Fatalf("neither pagination direction = sends %d, err %v; want zero-send error", sends, err)
	}
	both := maps.Clone(variables)
	both[pagination.BackwardPageSizeVariable] = 1
	if sends, err := invoke(both); err == nil || sends != 0 {
		t.Fatalf("both pagination directions = sends %d, err %v; want zero-send error", sends, err)
	}
	backward := maps.Clone(variables)
	delete(backward, pagination.PageSizeVariable)
	backward[pagination.BackwardPageSizeVariable] = 1
	if sends, err := invoke(backward); err != nil || sends != 1 {
		t.Fatalf("backward pagination = sends %d, err %v; want one send", sends, err)
	}
}
