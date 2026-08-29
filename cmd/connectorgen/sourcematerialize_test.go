package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestSourceMaterialize_HappyBundleAndByteIdenticalCheck(t *testing.T) {
	defsRoot, lockRaw, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{})
	stderr := new(bytes.Buffer)
	if code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, new(bytes.Buffer), stderr, fetcher); code != 0 {
		t.Fatalf("materialize exit code = %d stderr=%s", code, stderr.String())
	}
	bundleDir := filepath.Join(defsRoot, "alpha")
	for _, name := range []string{"metadata.json", "spec.json", "streams.json", "writes.json", "operations.json", "api_surface.json", "cli_surface.json", "docs.md", "missing-foundation.json", filepath.Join("sources", "alpha-operation-descriptor.json")} {
		if _, err := os.Stat(filepath.Join(bundleDir, name)); err != nil {
			t.Fatalf("generated %s: %v", name, err)
		}
	}
	bundle, err := engine.Load(os.DirFS(defsRoot), "alpha")
	if err != nil {
		t.Fatalf("generated bundle does not load: %v", err)
	}
	if bundle.HTTP.Check == nil || !reflect.DeepEqual(bundle.HTTP.Check.SuccessStatuses, []string{"200"}) || bundle.HTTP.Check.MaxBytes != 4096 {
		t.Fatalf("generated base check policy = %#v, want exact [200] and max_bytes 4096", bundle.HTTP.Check)
	}
	if findings, err := validateOperationalContractPath(bundleDir, "alpha", "declared"); err != nil || len(findings) != 0 {
		t.Fatalf("declared operational contract = findings=%+v err=%v", findings, err)
	}
	before := sourceMaterializeOwnedBytes(t, bundleDir)
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	if code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot, "--check"}, stdout, stderr, fetcher); code != 0 {
		t.Fatalf("byte-identical check exit = %d stderr=%s", code, stderr.String())
	}
	after := sourceMaterializeOwnedBytes(t, bundleDir)
	if !bytes.Equal(before, after) {
		t.Fatalf("--check changed owned output bytes\nbefore=%q\nafter=%q\nlock=%s", before, after, lockRaw)
	}
}

func TestSourceMaterialize_EdgeUsesExactRetainedSourceURLs(t *testing.T) {
	defsRoot, _, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{})
	var requested []string
	recordingFetcher := sourceImportFetchFunc(func(ctx context.Context, sourceURL string) ([]byte, error) {
		requested = append(requested, sourceURL)
		return fetcher.Fetch(ctx, sourceURL)
	})
	stderr := new(bytes.Buffer)
	if code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, new(bytes.Buffer), stderr, recordingFetcher); code != 0 {
		t.Fatalf("materialize exit = %d stderr=%s", code, stderr.String())
	}
	want := []string{
		"https://fixtures.polymetrics.invalid/get.openapi.json",
		"https://fixtures.polymetrics.invalid/post.openapi.json",
	}
	if !reflect.DeepEqual(requested, want) {
		t.Fatalf("retained source URLs = %#v, want %#v", requested, want)
	}
}

func TestSourceMaterialize_StatusSpellingsPreserveSourceEvidenceAndNormalizeRuntime(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		wantFailure string
	}{
		{name: "happy canonical 200", status: "200"},
		{name: "happy plus 200", status: "+200"},
		{name: "edge zero-padded 0200", status: "0200"},
		{name: "edge blocks exact non-2xx base check", status: "404", wantFailure: "runtime gap: exact response-status non-2xx execution"},
		{name: "bad uninterpretable status", status: "provider-success", wantFailure: `source operation "alpha.rest.get.shared" success status "provider-success" must be an unambiguous numeric HTTP status`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defsRoot, _, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{GetSuccessStatus: tt.status})
			stderr := new(bytes.Buffer)
			code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, new(bytes.Buffer), stderr, fetcher)
			if tt.wantFailure != "" {
				if code == 0 || !strings.Contains(stderr.String(), tt.wantFailure) {
					t.Fatalf("uninterpretable status result = %d/%q, want failure containing %q", code, stderr.String(), tt.wantFailure)
				}
				return
			}
			if code != 0 {
				t.Fatalf("materialize exit code = %d stderr=%s", code, stderr.String())
			}
			bundle, err := engine.Load(os.DirFS(defsRoot), "alpha")
			if err != nil {
				t.Fatalf("load materialized bundle: %v", err)
			}
			if bundle.HTTP.Check == nil || !reflect.DeepEqual(bundle.HTTP.Check.SuccessStatuses, []string{"200"}) {
				t.Fatalf("materialized check statuses = %#v, want exact runtime [200]", bundle.HTTP.Check)
			}
			if len(bundle.Operations) != 1 || bundle.Operations[0].REST == nil || bundle.Operations[0].REST.Response == nil || !reflect.DeepEqual(bundle.Operations[0].REST.Response.SuccessStatuses, []string{"200"}) {
				t.Fatalf("materialized operation statuses = %#v, want exact runtime [200]", bundle.Operations)
			}
			descriptor, err := os.ReadFile(filepath.Join(defsRoot, "alpha", "sources", "alpha-operation-descriptor.json"))
			if err != nil {
				t.Fatalf("read source descriptor: %v", err)
			}
			if !bytes.Contains(descriptor, []byte(`"status": "`+tt.status+`"`)) {
				t.Fatalf("source descriptor did not retain original status spelling %q: %s", tt.status, descriptor)
			}
		})
	}
}

func TestSourceMaterialize_EdgePreservesRawEquivalentStatusesAndDeduplicatesRuntime(t *testing.T) {
	defsRoot, _, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{
		GetSuccessStatus:           "200",
		GetAdditionalSuccessStatus: []string{"0200"},
		GetBindingSuccessStatuses:  []string{"200", "0200"},
	})
	stderr := new(bytes.Buffer)
	if code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, new(bytes.Buffer), stderr, fetcher); code != 0 {
		t.Fatalf("materialize equivalent statuses = %d stderr=%s", code, stderr.String())
	}
	bundle, err := engine.Load(os.DirFS(defsRoot), "alpha")
	if err != nil {
		t.Fatalf("load materialized bundle: %v", err)
	}
	if bundle.HTTP.Check == nil || !reflect.DeepEqual(bundle.HTTP.Check.SuccessStatuses, []string{"200"}) {
		t.Fatalf("materialized check statuses = %#v, want deduplicated exact runtime [200]", bundle.HTTP.Check)
	}
	if len(bundle.Operations) != 1 || bundle.Operations[0].REST == nil || bundle.Operations[0].REST.Response == nil || !reflect.DeepEqual(bundle.Operations[0].REST.Response.SuccessStatuses, []string{"200"}) {
		t.Fatalf("materialized operation statuses = %#v, want deduplicated exact runtime [200]", bundle.Operations)
	}
	descriptor, err := os.ReadFile(filepath.Join(defsRoot, "alpha", "sources", "alpha-operation-descriptor.json"))
	if err != nil {
		t.Fatalf("read source descriptor: %v", err)
	}
	for _, rawStatus := range []string{"200", "0200"} {
		if !bytes.Contains(descriptor, []byte(`"status": "`+rawStatus+`"`)) {
			t.Fatalf("source descriptor did not retain original status spelling %q: %s", rawStatus, descriptor)
		}
	}
}

func TestSourceMaterialize_EdgeAccountsFor2XXSourceStatusAsBlocked(t *testing.T) {
	defsRoot, _, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{PostSuccessStatus: "2XX", BlockPostForStatusRange: true})
	stderr := new(bytes.Buffer)
	if code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, new(bytes.Buffer), stderr, fetcher); code != 0 {
		t.Fatalf("blocked 2XX materialization = %d stderr=%s", code, stderr.String())
	}
	var report struct {
		Operations []sourceMaterializeFoundationRow `json:"operations"`
	}
	reportBytes, err := os.ReadFile(filepath.Join(defsRoot, "alpha", "missing-foundation.json"))
	if err != nil {
		t.Fatalf("read missing-foundation report: %v", err)
	}
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		t.Fatalf("decode missing-foundation report: %v", err)
	}
	var blocked *sourceMaterializeFoundationRow
	for i := range report.Operations {
		if report.Operations[i].SourceID == "alpha.rest.post.shared" {
			blocked = &report.Operations[i]
			break
		}
	}
	if blocked == nil || blocked.State != "blocked" || !strings.Contains(blocked.Reason, "runtime gap: exact response-status range execution") {
		t.Fatalf("2XX source accounting = %#v, want named blocked runtime gap", blocked)
	}
	descriptor, err := os.ReadFile(filepath.Join(defsRoot, "alpha", "sources", "alpha-operation-descriptor.json"))
	if err != nil {
		t.Fatalf("read source descriptor: %v", err)
	}
	if !bytes.Contains(descriptor, []byte(`"status": "2XX"`)) {
		t.Fatalf("source descriptor did not retain 2XX provider spelling: %s", descriptor)
	}
	writes, err := os.ReadFile(filepath.Join(defsRoot, "alpha", "writes.json"))
	if err != nil {
		t.Fatalf("read writes: %v", err)
	}
	if bytes.Contains(writes, []byte("create_widget")) {
		t.Fatalf("2XX source operation was widened into an executable write: %s", writes)
	}
}

func TestSourceMaterializeWrite_EdgeNamesExactDeclared404RuntimeGap(t *testing.T) {
	defsRoot, _, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{PostSuccessStatus: "404"})
	stderr := new(bytes.Buffer)
	if code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, new(bytes.Buffer), stderr, fetcher); code == 0 || !strings.Contains(stderr.String(), "runtime gap: exact response-status non-2xx execution") {
		t.Fatalf("materialize exact declared 404 = %d stderr=%s, want named runtime gap", code, stderr.String())
	}
	lock, err := os.ReadFile(filepath.Join(defsRoot, "alpha", "sources", "alpha-operation-source-lock.json"))
	if err != nil {
		t.Fatalf("read source lock: %v", err)
	}
	if !bytes.Contains(lock, []byte(`"404"`)) {
		t.Fatalf("source lock did not retain 404 provider spelling: %s", lock)
	}
}

func TestSourceMaterialize_BadUninterpretableRetainedResponseIsValidationError(t *testing.T) {
	defsRoot, _, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{GetAdditionalSuccessStatus: []string{"provider-status"}})
	stderr := new(bytes.Buffer)
	if code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, new(bytes.Buffer), stderr, fetcher); code == 0 || !strings.Contains(stderr.String(), `source operation "alpha.rest.get.shared" response status "provider-status" must be an unambiguous numeric HTTP status`) {
		t.Fatalf("uninterpretable retained response result = %d stderr=%s, want source-context validation error", code, stderr.String())
	}
}

func TestSourceMaterialize_RejectsBadV4Plans(t *testing.T) {
	tests := []struct {
		name      string
		options   sourceMaterializeFixtureOptions
		wantError string
	}{
		{name: "legacy lock requires v4", options: sourceMaterializeFixtureOptions{Legacy: true}, wantError: "schema v4"},
		{name: "unknown materialization field", options: sourceMaterializeFixtureOptions{UnknownMaterializationField: true}, wantError: "unknown field"},
		{name: "duplicate source accounting", options: sourceMaterializeFixtureOptions{DuplicateAccounting: true}, wantError: "duplicate source operation ID"},
		{name: "non-api integration type", options: sourceMaterializeFixtureOptions{IntegrationType: "database"}, wantError: "metadata.integration_type \"database\" is not supported"},
		{name: "required body input lacks binding", options: sourceMaterializeFixtureOptions{MissingBodyBinding: true}, wantError: "required input \"body.title\""},
		{name: "unselected request media arm", options: sourceMaterializeFixtureOptions{WrongRequestMedia: true, MultipleMedia: true}, wantError: "not the selected JSON mutation contract"},
		{name: "blocked row cannot carry executable inputs", options: sourceMaterializeFixtureOptions{BlockedWriteWithInputs: true}, wantError: "must not declare inputs"},
		{name: "blocked check row cannot become executable", options: sourceMaterializeFixtureOptions{BlockedCheck: true}, wantError: "must select a materialized"},
		{name: "runtime-blocked check row cannot become executable", options: sourceMaterializeFixtureOptions{RuntimeBlockedCheck: true}, wantError: "not runtime-admissible"},
		{name: "edge wildcard check status names range execution gap", options: sourceMaterializeFixtureOptions{WildcardCheckStatus: true}, wantError: "runtime gap: exact response-status range execution"},
		{name: "edge numeric status range names range execution gap", options: sourceMaterializeFixtureOptions{PostSuccessStatus: "200-299"}, wantError: "runtime gap: exact response-status range execution"},
		{name: "owned output symlink escape", options: sourceMaterializeFixtureOptions{OutputSymlink: true}, wantError: "must not traverse a symlink"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defsRoot, _, fetcher := sourceMaterializeFixture(t, tt.options)
			stderr := new(bytes.Buffer)
			code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, new(bytes.Buffer), stderr, fetcher)
			if code == 0 || !strings.Contains(stderr.String(), tt.wantError) {
				t.Fatalf("exit/stderr = %d/%q, want failure containing %q", code, stderr.String(), tt.wantError)
			}
		})
	}
}

func TestValidate_OperationalContractRejectsMismatchedBundleTarget(t *testing.T) {
	defsRoot, _, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{})
	if code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, new(bytes.Buffer), new(bytes.Buffer), fetcher); code != 0 {
		t.Fatalf("materialize exit = %d", code)
	}
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	code := run([]string{"validate", filepath.Join(defsRoot, "alpha"), "--connector", "other", "--require-operational-contract", "write", "--json"}, stdout, stderr)
	if code != 2 || !strings.Contains(stderr.String(), "does not match bundle target") || stdout.Len() != 0 {
		t.Fatalf("mismatched target = %d stdout=%q stderr=%q, want typed target mismatch", code, stdout.String(), stderr.String())
	}
}

func TestValidate_OperationalContractRejectsExplicitEmptyValues(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "empty profile and connector", args: []string{"validate", "cmd/connectorgen/testdata/valid/goodconn", "--require-operational-contract", "", "--connector", ""}},
		{name: "whitespace profile", args: []string{"validate", "cmd/connectorgen/testdata/valid/goodconn", "--connector", "goodconn", "--require-operational-contract", " \t "}},
		{name: "empty connector", args: []string{"validate", "cmd/connectorgen/testdata/valid/goodconn", "--connector", "", "--require-operational-contract", "write"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
			if code := run(tt.args, stdout, stderr); code != 2 || !strings.Contains(stderr.String(), "non-empty value") || stdout.Len() != 0 {
				t.Fatalf("explicit empty value = %d stdout=%q stderr=%q, want closed usage error", code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestValidate_OperationalContractRejectsNonAPIIntegration(t *testing.T) {
	defsRoot, _, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{})
	if code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, new(bytes.Buffer), new(bytes.Buffer), fetcher); code != 0 {
		t.Fatalf("materialize exit = %d", code)
	}
	metadataPath := filepath.Join(defsRoot, "alpha", "metadata.json")
	raw, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatal(err)
	}
	var metadata map[string]any
	if err := json.Unmarshal(raw, &metadata); err != nil {
		t.Fatal(err)
	}
	metadata["integration_type"] = "database"
	changed, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(metadataPath, changed, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := validateOperationalContractPath(filepath.Join(defsRoot, "alpha"), "alpha", "write"); err == nil || !strings.Contains(err.Error(), "metadata.integration_type \"api\"") {
		t.Fatalf("non-api operational contract error = %v", err)
	}
}

func TestSourceMaterialize_AtomicNoWriteOnLateFailure(t *testing.T) {
	defsRoot, _, _ := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{})
	bundleDir := filepath.Join(defsRoot, "alpha")
	sentinel := []byte(`{"sentinel":"original"}\n`)
	if err := os.WriteFile(filepath.Join(bundleDir, "metadata.json"), sentinel, 0o644); err != nil {
		t.Fatalf("seed existing owned output: %v", err)
	}
	// The source-materializer stages generated bytes before calling the same
	// loader/check suite used by production. Invalid staged metadata therefore
	// reaches a genuinely late failure, after every staged write but before the
	// bundle-directory rename.
	err := sourceMaterializePublish(bundleDir, "alpha", []sourceMaterializeOutput{{RelativePath: "metadata.json", Bytes: []byte(`{}`)}}, new(bytes.Buffer))
	if err == nil || !strings.Contains(err.Error(), "load staged bundle") {
		t.Fatalf("late staged failure = %v, want staged loader failure", err)
	}
	got, err := os.ReadFile(filepath.Join(bundleDir, "metadata.json"))
	if err != nil || !bytes.Equal(got, sentinel) {
		t.Fatalf("atomic failure changed original metadata: got=%q err=%v", got, err)
	}
}

func TestSourceMaterialize_RejectsExplicitEmptyDefsWithoutWrite(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "empty separate value", args: []string{"source-materialize", "alpha", "--defs", "", "--check"}},
		{name: "whitespace separate value", args: []string{"source-materialize", "alpha", "--defs", " \t ", "--check"}},
		{name: "empty equals value", args: []string{"source-materialize", "alpha", "--defs=", "--check"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defsRoot, _, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{})
			stderr := new(bytes.Buffer)
			if code := runSourceMaterializeWithFetcher(tt.args, new(bytes.Buffer), stderr, fetcher); code != 2 || !strings.Contains(stderr.String(), "non-empty value") {
				t.Fatalf("empty --defs = %d/%q, want closed usage error", code, stderr.String())
			}
			if _, err := os.Stat(filepath.Join(defsRoot, "alpha", "metadata.json")); !os.IsNotExist(err) {
				t.Fatalf("rejected --defs wrote metadata: %v", err)
			}
		})
	}
}

func TestSourceMaterializeWrite_RejectsOpenProviderObjectWithoutSynthesis(t *testing.T) {
	source := sourceOperationDescriptor{
		SourceID: "alpha.rest.post.open-widget",
		Protocol: "rest",
		Method:   "POST",
		Path:     "/widgets",
		Request: sourceRequestDescriptor{
			MediaType: "application/json",
			Body: &sourceRequestBodyDescriptor{Required: true, Schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"title": map[string]any{"type": "string"}},
			}},
		},
		Responses: []sourceResponseDescriptor{{Status: "201"}},
	}
	row := sourceMaterializationOperation{SourceID: source.SourceID, Inputs: []sourceMaterializationInputBinding{{Source: "body.title", Target: "record.title"}}}
	binding := sourceMaterializationOperationBind{Kind: "write", ID: "create_widget", RequestMedia: "application/json", WriteKind: "create", MutationClass: "create", Approval: "none", Risk: "high", SuccessStatuses: []string{"201"}}
	if _, _, err := sourceMaterializeWrite(source, row, binding); err == nil || !strings.Contains(err.Error(), "mark it blocked") {
		t.Fatalf("open provider object error = %v, want actionable blocked mapping", err)
	}
}

func TestSourceMaterializeDirectRead_RejectsRuntimeUnrepresentableContracts(t *testing.T) {
	base := sourceOperationDescriptor{
		SourceID:  "alpha.rest.get.widgets",
		Protocol:  "rest",
		Method:    "GET",
		Path:      "/widgets",
		Request:   sourceRequestDescriptor{Path: []sourceParameterDescriptor{}, Query: []sourceParameterDescriptor{}, Header: []sourceParameterDescriptor{}},
		Responses: []sourceResponseDescriptor{{Status: "200"}},
		Output:    sourceOutputDescriptor{Class: sourceOutputJSON},
	}
	row := sourceMaterializationOperation{SourceID: base.SourceID}
	binding := sourceMaterializationOperationBind{Kind: "direct_read", ID: "list_widgets", CommandPath: "widgets list", CommandSummary: "List widgets", OutputPolicy: "json_redacted", MaxResponseBytes: 4096, SuccessStatuses: []string{"200"}, Risk: "low"}
	tests := []struct {
		name   string
		mutate func(*sourceOperationDescriptor)
		want   string
	}{
		{name: "descriptor merge blocked", mutate: func(source *sourceOperationDescriptor) { source.Runtime.MergeBlocked = true }, want: "runtime reachability is blocked"},
		{name: "declared auth scope", mutate: func(source *sourceOperationDescriptor) { source.AuthScopes.Declared = true }, want: "declared authentication scope"},
		{name: "declared pagination", mutate: func(source *sourceOperationDescriptor) { source.Pagination = map[string]any{"type": "cursor"} }, want: "declared pagination"},
		{name: "declared server routing", mutate: func(source *sourceOperationDescriptor) { source.Servers.Root.Declared = true }, want: "declared server routing"},
		{name: "response byte limit mismatch", mutate: func(source *sourceOperationDescriptor) { source.ByteLimits.Response = 128 }, want: "response byte limit"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := base
			tt.mutate(&source)
			if _, _, _, err := sourceMaterializeDirectRead(source, row, binding); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("direct read error = %v, want fail-closed error containing %q", err, tt.want)
			}
		})
	}
}

func TestSourceMaterializeDirectRead_RejectsQueryWithoutExactWireRepresentation(t *testing.T) {
	source := sourceOperationDescriptor{
		SourceID: "alpha.rest.get.widgets",
		Protocol: "rest",
		Method:   "GET",
		Path:     "/widgets",
		Request: sourceRequestDescriptor{
			Path:   []sourceParameterDescriptor{},
			Header: []sourceParameterDescriptor{},
			Query:  []sourceParameterDescriptor{{Name: "tags", Schema: map[string]any{"type": "array"}, Wire: sourceParameterWireDescriptor{Style: "form"}}},
		},
		Responses: []sourceResponseDescriptor{{Status: "200"}},
		Output:    sourceOutputDescriptor{Class: sourceOutputJSON},
	}
	row := sourceMaterializationOperation{SourceID: source.SourceID, Inputs: []sourceMaterializationInputBinding{{Source: "query.tags", Target: "config.base_url"}}}
	binding := sourceMaterializationOperationBind{Kind: "direct_read", ID: "list_widgets", CommandPath: "widgets list", CommandSummary: "List widgets", OutputPolicy: "json_redacted", MaxResponseBytes: 4096, SuccessStatuses: []string{"200"}, Risk: "low"}
	if _, _, _, err := sourceMaterializeDirectRead(source, row, binding); err == nil || !strings.Contains(err.Error(), "query wire") {
		t.Fatalf("query wire error = %v, want explicit blocked mapping", err)
	}
}

func TestSourceMaterializeWrite_RejectsVendorJSONWhenRuntimeSendsApplicationJSON(t *testing.T) {
	source := sourceOperationDescriptor{
		SourceID: "alpha.rest.post.widget",
		Protocol: "rest",
		Method:   "POST",
		Path:     "/widgets",
		Request: sourceRequestDescriptor{
			MediaType: "application/vnd.alpha+json",
			Body: &sourceRequestBodyDescriptor{Required: true, Schema: map[string]any{
				"type": "object", "additionalProperties": false,
				"required": []any{"title"}, "properties": map[string]any{"title": map[string]any{"type": "string"}},
			}},
		},
		Responses: []sourceResponseDescriptor{{Status: "201"}},
	}
	row := sourceMaterializationOperation{SourceID: source.SourceID, Inputs: []sourceMaterializationInputBinding{{Source: "body.title", Target: "record.title"}}}
	binding := sourceMaterializationOperationBind{Kind: "write", ID: "create_widget", RequestMedia: "application/vnd.alpha+json", WriteKind: "create", MutationClass: "create", Approval: "none", Risk: "high", SuccessStatuses: []string{"201"}}
	if _, _, err := sourceMaterializeWrite(source, row, binding); err == nil || !strings.Contains(err.Error(), "application/json") {
		t.Fatalf("vendor JSON error = %v, want exact-media blocked mapping", err)
	}
}

func TestSourceMaterialize_HappyAccountsForOpenProviderWriteAsBlocked(t *testing.T) {
	defsRoot, _, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{OpenWriteBody: true, BlockOpenWrite: true})
	stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
	if code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, stdout, stderr, fetcher); code != 0 {
		t.Fatalf("blocked open-write materialization = %d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}

	var report struct {
		Operations []sourceMaterializeFoundationRow `json:"operations"`
	}
	reportBytes, err := os.ReadFile(filepath.Join(defsRoot, "alpha", "missing-foundation.json"))
	if err != nil {
		t.Fatalf("read missing-foundation report: %v", err)
	}
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		t.Fatalf("decode missing-foundation report: %v", err)
	}
	var blocked *sourceMaterializeFoundationRow
	for i := range report.Operations {
		if report.Operations[i].SourceID == "alpha.rest.post.shared" {
			blocked = &report.Operations[i]
			break
		}
	}
	if blocked == nil || blocked.State != "blocked" || !strings.Contains(blocked.Reason, "open provider object") {
		t.Fatalf("open write accounting = %#v, want blocked row with runtime reason", blocked)
	}
	apiSurface, err := os.ReadFile(filepath.Join(defsRoot, "alpha", "api_surface.json"))
	if err != nil {
		t.Fatalf("read API surface: %v", err)
	}
	if !bytes.Contains(apiSurface, []byte(`"status": "blocked"`)) || !bytes.Contains(apiSurface, []byte("open provider object")) {
		t.Fatalf("blocked API surface = %s", apiSurface)
	}
	writes, err := os.ReadFile(filepath.Join(defsRoot, "alpha", "writes.json"))
	if err != nil {
		t.Fatalf("read writes: %v", err)
	}
	if bytes.Contains(writes, []byte("create_widget")) {
		t.Fatalf("blocked write was emitted as executable: %s", writes)
	}
}

func TestSourceMaterialize_GraphQLBlockedRowsRetainExactDocumentCitation(t *testing.T) {
	_, lockRaw, restFetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{})
	graphqlRaw := []byte("type Query { viewer: String }\n")
	digest := sha256.Sum256(graphqlRaw)
	graphql := sourceImportGraphQL{
		sourceImportArtifact: sourceImportArtifact{
			SourceURL: "https://fixtures.polymetrics.invalid/alpha.graphql",
			SHA256:    hex.EncodeToString(digest[:]),
			Bytes:     int64(len(graphqlRaw)),
		},
		QueryFields: []sourceGraphQLField{{
			Root: "Query", Name: "viewer", Line: 1, Signature: "viewer: String",
			Arguments: []sourceGraphQLArgument{}, ReturnType: sourceGraphQLTypeRef{Kind: "named", Name: "String"},
		}},
		TypeSystem: sourceGraphQLTypeSystem{Enums: []sourceGraphQLNamedType{}, InputObjects: []sourceGraphQLNamedType{}, Interfaces: []sourceGraphQLNamedType{}, Objects: []sourceGraphQLNamedType{}, Scalars: []string{"String"}, Unions: []sourceGraphQLNamedType{}},
	}
	projection, err := canonicalSourceGraphQLProjection(graphql)
	if err != nil {
		t.Fatal(err)
	}
	projectionDigest := sha256.Sum256(projection)
	graphql.ProjectionSHA256 = hex.EncodeToString(projectionDigest[:])
	graphql.ProjectionBytes = int64(len(projection))
	graphqlRawPlan, err := json.Marshal(graphql)
	if err != nil {
		t.Fatal(err)
	}
	var graphqlPlan map[string]any
	if err := json.Unmarshal(graphqlRawPlan, &graphqlPlan); err != nil {
		t.Fatal(err)
	}
	graphqlPlan["document_id"] = "graphql-schema"

	var raw map[string]any
	if err := json.Unmarshal(lockRaw, &raw); err != nil {
		t.Fatal(err)
	}
	raw["graphql"] = graphqlPlan
	raw["counts"] = map[string]any{"rest": 2, "graphql_query": 1, "graphql_mutation": 0, "total": 3}
	plan := raw["materialization"].(map[string]any)
	plan["operations"] = append(plan["operations"].([]any), map[string]any{
		"source_id": "alpha.graphql.query.viewer", "state": "blocked",
		"citation": map[string]any{"document_id": "graphql-schema", "location": `graphql.query_fields["viewer"]@line:1`},
		"reason":   "No fixed GraphQL document materialization section is declared.",
	})
	withGraphQL, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := parseSourceImportLock(withGraphQL, "alpha")
	if err != nil {
		t.Fatalf("parse v4 GraphQL lock: %v", err)
	}
	result, err := importSourceLockResult(context.Background(), lock, sourceImportFetchFunc(func(ctx context.Context, sourceURL string) ([]byte, error) {
		if sourceURL == graphql.SourceURL {
			return graphqlRaw, nil
		}
		return restFetcher.Fetch(ctx, sourceURL)
	}), defaultSourceImportLimits())
	if err != nil {
		t.Fatalf("import v4 GraphQL lock: %v", err)
	}
	var imported *sourceOperationDescriptor
	for index := range result.Operations {
		if result.Operations[index].SourceID == "alpha.graphql.query.viewer" {
			imported = &result.Operations[index]
			break
		}
	}
	if imported == nil || imported.Source.DocumentID != "graphql-schema" {
		t.Fatalf("GraphQL document identity = %#v, want retained graphql-schema identity", imported)
	}
	if _, err := sourceMaterializeOutputs(lock, result); err != nil {
		t.Fatalf("all-accounted blocked GraphQL materialization: %v", err)
	}
}

func TestSourceImportV4_RejectsGraphQLDocumentIDCollisionWithREST(t *testing.T) {
	_, lockRaw, _ := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{})
	graphqlRaw := []byte("type Query { viewer: String }\n")
	digest := sha256.Sum256(graphqlRaw)
	graphql := sourceImportGraphQL{
		sourceImportArtifact: sourceImportArtifact{
			SourceURL: "https://fixtures.polymetrics.invalid/alpha.graphql",
			SHA256:    hex.EncodeToString(digest[:]),
			Bytes:     int64(len(graphqlRaw)),
		},
		QueryFields: []sourceGraphQLField{{
			Root: "Query", Name: "viewer", Line: 1, Signature: "viewer: String",
			Arguments: []sourceGraphQLArgument{}, ReturnType: sourceGraphQLTypeRef{Kind: "named", Name: "String"},
		}},
		TypeSystem: sourceGraphQLTypeSystem{Enums: []sourceGraphQLNamedType{}, InputObjects: []sourceGraphQLNamedType{}, Interfaces: []sourceGraphQLNamedType{}, Objects: []sourceGraphQLNamedType{}, Scalars: []string{"String"}, Unions: []sourceGraphQLNamedType{}},
	}
	projection, err := canonicalSourceGraphQLProjection(graphql)
	if err != nil {
		t.Fatal(err)
	}
	projectionDigest := sha256.Sum256(projection)
	graphql.ProjectionSHA256 = hex.EncodeToString(projectionDigest[:])
	graphql.ProjectionBytes = int64(len(projection))
	graphqlRawPlan, err := json.Marshal(graphql)
	if err != nil {
		t.Fatal(err)
	}
	var graphqlPlan map[string]any
	if err := json.Unmarshal(graphqlRawPlan, &graphqlPlan); err != nil {
		t.Fatal(err)
	}
	graphqlPlan["document_id"] = "get"

	var raw map[string]any
	if err := json.Unmarshal(lockRaw, &raw); err != nil {
		t.Fatal(err)
	}
	raw["graphql"] = graphqlPlan
	raw["counts"] = map[string]any{"rest": 2, "graphql_query": 1, "graphql_mutation": 0, "total": 3}
	colliding, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parseSourceImportLock(colliding, "alpha"); err == nil || !strings.Contains(err.Error(), "collides with REST source document ID") {
		t.Fatalf("GraphQL document ID collision error = %v, want cited REST collision", err)
	}
}

func TestSourceMaterialize_EdgeRejectsSymlinkedBundleTarget(t *testing.T) {
	defsRoot, _, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{})
	alphaDir := filepath.Join(defsRoot, "alpha")
	betaDir := filepath.Join(defsRoot, "beta")
	if err := os.Rename(alphaDir, betaDir); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("beta", alphaDir); err != nil {
		t.Skipf("symlink support unavailable: %v", err)
	}
	stderr := new(bytes.Buffer)
	if code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, new(bytes.Buffer), stderr, fetcher); code == 0 || !strings.Contains(stderr.String(), "bundle must not be a symlink") {
		t.Fatalf("symlinked bundle = %d/%q, want fail-closed bundle identity error", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(betaDir, "metadata.json")); !os.IsNotExist(err) {
		t.Fatalf("symlinked target received generated output: %v", err)
	}
}

func TestSourceMaterializePublish_RollsBackInstallRenameFailure(t *testing.T) {
	bundleDir := sourceMaterializeGeneratedBundle(t)
	metadataPath := filepath.Join(bundleDir, "metadata.json")
	before, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatal(err)
	}
	renames := 0
	err = sourceMaterializePublishWithOps(bundleDir, "alpha", nil, sourceMaterializePublishOps{
		Rename: func(oldPath, newPath string) error {
			renames++
			if renames == 2 {
				return errors.New("injected staged-install rename failure")
			}
			return os.Rename(oldPath, newPath)
		},
		Remove:    os.Remove,
		RemoveAll: os.RemoveAll,
	})
	if err == nil || !strings.Contains(err.Error(), "publish staged bundle") {
		t.Fatalf("install rename failure = %v", err)
	}
	after, readErr := os.ReadFile(metadataPath)
	if readErr != nil || !bytes.Equal(after, before) {
		t.Fatalf("rollback did not preserve live bundle: metadata=%q err=%v", after, readErr)
	}
}

func TestSourceMaterializePublish_CleanupFailureDoesNotFailInstalledBundle(t *testing.T) {
	bundleDir := sourceMaterializeGeneratedBundle(t)
	warnings := new(bytes.Buffer)
	err := sourceMaterializePublishWithOps(bundleDir, "alpha", nil, sourceMaterializePublishOps{
		Rename: os.Rename,
		Remove: os.Remove,
		RemoveAll: func(path string) error {
			if strings.Contains(filepath.Base(path), "-source-materialize-backup-") {
				return errors.New("injected backup cleanup failure")
			}
			return os.RemoveAll(path)
		},
		Warn: func(message string) { logln(warnings, message) },
	})
	if err != nil {
		t.Fatalf("successful install reported cleanup failure: %v", err)
	}
	if _, err := os.Stat(filepath.Join(bundleDir, "metadata.json")); err != nil {
		t.Fatalf("successful install removed live bundle: %v", err)
	}
	if !strings.Contains(warnings.String(), "backup cleanup") || !strings.Contains(warnings.String(), "injected backup cleanup failure") {
		t.Fatalf("backup cleanup warning = %q", warnings.String())
	}
}

func TestSourceMaterializePublish_StageCleanupFailureWarnsButDoesNotFailInstalledBundle(t *testing.T) {
	bundleDir := sourceMaterializeGeneratedBundle(t)
	warnings := new(bytes.Buffer)
	err := sourceMaterializePublishWithOps(bundleDir, "alpha", nil, sourceMaterializePublishOps{
		Rename: os.Rename,
		Remove: os.Remove,
		RemoveAll: func(path string) error {
			base := filepath.Base(path)
			if strings.Contains(base, "-source-materialize-") && !strings.Contains(base, "-source-materialize-backup-") {
				return errors.New("injected stage cleanup failure")
			}
			return os.RemoveAll(path)
		},
		Warn: func(message string) { logln(warnings, message) },
	})
	if err != nil {
		t.Fatalf("successful install reported stage cleanup failure: %v", err)
	}
	if _, err := os.Stat(filepath.Join(bundleDir, "metadata.json")); err != nil {
		t.Fatalf("successful install removed live bundle: %v", err)
	}
	if !strings.Contains(warnings.String(), "staging cleanup") || !strings.Contains(warnings.String(), "injected stage cleanup failure") {
		t.Fatalf("stage cleanup warning = %q", warnings.String())
	}
}

func TestSourceMaterializePublish_PrePublishCleanupFailurePreservesPrimaryErrorAndWarns(t *testing.T) {
	bundleDir := sourceMaterializeGeneratedBundle(t)
	warnings := new(bytes.Buffer)
	err := sourceMaterializePublishWithOps(bundleDir, "alpha", []sourceMaterializeOutput{{RelativePath: "metadata.json", Bytes: []byte(`{}`)}}, sourceMaterializePublishOps{
		Rename: os.Rename,
		Remove: os.Remove,
		RemoveAll: func(path string) error {
			if strings.Contains(filepath.Base(path), "-source-materialize-") {
				return errors.New("injected pre-publish cleanup failure")
			}
			return os.RemoveAll(path)
		},
		Warn: func(message string) { logln(warnings, message) },
	})
	if err == nil || !strings.Contains(err.Error(), "load staged bundle") {
		t.Fatalf("pre-publish primary error = %v", err)
	}
	if !strings.Contains(warnings.String(), "after failed publish") || !strings.Contains(warnings.String(), "injected pre-publish cleanup failure") {
		t.Fatalf("pre-publish cleanup warning = %q", warnings.String())
	}
}

func TestSourceMaterializePublish_RollbackCleanupFailurePreservesPrimaryErrorAndWarns(t *testing.T) {
	bundleDir := sourceMaterializeGeneratedBundle(t)
	warnings := new(bytes.Buffer)
	renames := 0
	err := sourceMaterializePublishWithOps(bundleDir, "alpha", nil, sourceMaterializePublishOps{
		Rename: func(oldPath, newPath string) error {
			renames++
			if renames == 2 {
				return errors.New("injected staged-install rename failure")
			}
			return os.Rename(oldPath, newPath)
		},
		Remove: os.Remove,
		RemoveAll: func(path string) error {
			if strings.Contains(filepath.Base(path), "-source-materialize-") {
				return errors.New("injected rollback cleanup failure")
			}
			return os.RemoveAll(path)
		},
		Warn: func(message string) { logln(warnings, message) },
	})
	if err == nil || !strings.Contains(err.Error(), "publish staged bundle") {
		t.Fatalf("rollback primary error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(bundleDir, "metadata.json")); err != nil {
		t.Fatalf("successful rollback did not restore live bundle: %v", err)
	}
	if !strings.Contains(warnings.String(), "after successful rollback") || !strings.Contains(warnings.String(), "injected rollback cleanup failure") {
		t.Fatalf("rollback cleanup warning = %q", warnings.String())
	}
}

func TestSourceMaterializePublish_ReportsUnrecoverableRollbackBoundary(t *testing.T) {
	bundleDir := sourceMaterializeGeneratedBundle(t)
	renames := 0
	err := sourceMaterializePublishWithOps(bundleDir, "alpha", nil, sourceMaterializePublishOps{
		Rename: func(oldPath, newPath string) error {
			renames++
			if renames >= 2 {
				return errors.New("injected commit-window rename failure")
			}
			return os.Rename(oldPath, newPath)
		},
		Remove:    os.Remove,
		RemoveAll: os.RemoveAll,
	})
	if err == nil || !strings.Contains(err.Error(), "previous bundle retained at") || !strings.Contains(err.Error(), "manual recovery") {
		t.Fatalf("unrecoverable rollback error = %v", err)
	}
	if _, statErr := os.Stat(bundleDir); !os.IsNotExist(statErr) {
		t.Fatalf("unrecoverable boundary expected live path absent, stat err=%v", statErr)
	}
}

func TestSourceMaterializePublish_BackupReservationCleanupBoundary(t *testing.T) {
	tests := []struct {
		name          string
		removeAllFail bool
	}{
		{name: "cleans reserved backup after preparation failure"},
		{name: "names reserved backup when cleanup also fails", removeAllFail: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundleDir := sourceMaterializeGeneratedBundle(t)
			var reserved string
			err := sourceMaterializePublishWithOps(bundleDir, "alpha", nil, sourceMaterializePublishOps{
				Rename: os.Rename,
				Remove: func(path string) error {
					if strings.Contains(filepath.Base(path), "-source-materialize-backup-") {
						reserved = path
						return errors.New("injected backup reservation removal failure")
					}
					return os.Remove(path)
				},
				RemoveAll: func(path string) error {
					if path == reserved && tt.removeAllFail {
						return errors.New("injected reserved backup cleanup failure")
					}
					return os.RemoveAll(path)
				},
			})
			if err == nil || !strings.Contains(err.Error(), "injected backup reservation removal failure") {
				t.Fatalf("backup reservation error = %v", err)
			}
			if reserved == "" {
				t.Fatal("publish did not reserve a backup path")
			}
			if tt.removeAllFail {
				if !strings.Contains(err.Error(), reserved) || !strings.Contains(err.Error(), "injected reserved backup cleanup failure") {
					t.Fatalf("unrecovered reservation error = %q, want path and cleanup context", err)
				}
				return
			}
			if _, statErr := os.Stat(reserved); !os.IsNotExist(statErr) {
				t.Fatalf("reserved backup was left behind at %q: %v", reserved, statErr)
			}
		})
	}
}

func sourceMaterializeGeneratedBundle(t *testing.T) string {
	t.Helper()
	defsRoot, _, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{})
	stderr := new(bytes.Buffer)
	if code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, new(bytes.Buffer), stderr, fetcher); code != 0 {
		t.Fatalf("materialize generated bundle = %d stderr=%s", code, stderr.String())
	}
	return filepath.Join(defsRoot, "alpha")
}

func TestValidate_OperationalContractGateIsOptIn(t *testing.T) {
	defsRoot, _, fetcher := sourceMaterializeFixture(t, sourceMaterializeFixtureOptions{})
	stderr := new(bytes.Buffer)
	if code := runSourceMaterializeWithFetcher([]string{"source-materialize", "alpha", "--defs", defsRoot}, new(bytes.Buffer), stderr, fetcher); code != 0 {
		t.Fatalf("materialize exit = %d stderr=%s", code, stderr.String())
	}
	validateStderr := new(bytes.Buffer)
	validateStdout := new(bytes.Buffer)
	ordinaryCode := run([]string{"validate", filepath.Join(defsRoot, "alpha")}, validateStdout, validateStderr)
	// A v4 lock is intentionally outside the legacy source-projection mapper,
	// so preserve validate's existing source-projection result verbatim. The
	// opt-in flag must neither alter that result nor enable anything globally.
	connectorStdout, connectorStderr := new(bytes.Buffer), new(bytes.Buffer)
	connectorCode := run([]string{"validate", filepath.Join(defsRoot, "alpha"), "--connector", "alpha"}, connectorStdout, connectorStderr)
	if connectorCode != ordinaryCode || connectorStdout.String() != validateStdout.String() || connectorStderr.String() != validateStderr.String() {
		t.Fatalf("connector selection changed ordinary validate: before=%d/%q/%q after=%d/%q/%q", ordinaryCode, validateStdout.String(), validateStderr.String(), connectorCode, connectorStdout.String(), connectorStderr.String())
	}
	gateStdout, gateStderr := new(bytes.Buffer), new(bytes.Buffer)
	gateCode := run([]string{"validate", filepath.Join(defsRoot, "alpha"), "--connector", "alpha", "--require-operational-contract", "write"}, gateStdout, gateStderr)
	if gateCode != ordinaryCode || gateStdout.String() != validateStdout.String() || gateStderr.String() != validateStderr.String() {
		t.Fatalf("satisfied opt-in write gate changed validate result: before=%d/%q/%q after=%d/%q/%q", ordinaryCode, validateStdout.String(), validateStderr.String(), gateCode, gateStdout.String(), gateStderr.String())
	}
	invalidStderr := new(bytes.Buffer)
	if code := run([]string{"validate", filepath.Join(defsRoot, "alpha"), "--connector", "alpha", "--require-operational-contract", "query"}, new(bytes.Buffer), invalidStderr); code != 1 || !strings.Contains(invalidStderr.String(), "operational contract profile") {
		t.Fatalf("invalid profile = %d/%q, want operational-contract failure", code, invalidStderr.String())
	}
}

type sourceMaterializeFixtureOptions struct {
	Legacy                      bool
	UnknownMaterializationField bool
	DuplicateAccounting         bool
	IntegrationType             string
	MissingBodyBinding          bool
	WrongRequestMedia           bool
	MultipleMedia               bool
	OpenWriteBody               bool
	BlockOpenWrite              bool
	BlockedWriteWithInputs      bool
	BlockedCheck                bool
	RuntimeBlockedCheck         bool
	WildcardCheckStatus         bool
	GetSuccessStatus            string
	GetAdditionalSuccessStatus  []string
	GetBindingSuccessStatuses   []string
	PostSuccessStatus           string
	BlockPostForStatusRange     bool
	OutputSymlink               bool
}

func sourceMaterializeFixture(t *testing.T, options sourceMaterializeFixtureOptions) (string, []byte, sourceImportFetchFunc) {
	t.Helper()
	defsRoot := t.TempDir()
	sourcesDir := filepath.Join(defsRoot, "alpha", "sources")
	if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	getStatus := options.GetSuccessStatus
	if getStatus == "" {
		getStatus = "200"
	}
	getRaw := []byte(`{"openapi":"3.0.3","info":{"title":"alpha","version":"1"},"paths":{"/widgets":{"get":{"operationId":"listWidgets","responses":{"200":{"description":"ok","content":{"application/json":{"schema":{"type":"object","additionalProperties":false,"properties":{"items":{"type":"array","items":{"type":"string"}}}}}}}}}}}}`)
	getRaw = bytes.Replace(getRaw, []byte(`"200"`), []byte(`"`+getStatus+`"`), 1)
	if options.RuntimeBlockedCheck {
		var getDocument map[string]any
		if err := json.Unmarshal(getRaw, &getDocument); err != nil {
			t.Fatal(err)
		}
		getDocument["paths"].(map[string]any)["/widgets"].(map[string]any)["get"].(map[string]any)["parameters"] = []any{map[string]any{
			"name": "filter", "in": "query", "style": "deepObject",
			"schema": map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"name": map[string]any{"type": "string"}}},
		}}
		var err error
		getRaw, err = json.Marshal(getDocument)
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(options.GetAdditionalSuccessStatus) != 0 {
		var getDocument map[string]any
		if err := json.Unmarshal(getRaw, &getDocument); err != nil {
			t.Fatal(err)
		}
		responses := getDocument["paths"].(map[string]any)["/widgets"].(map[string]any)["get"].(map[string]any)["responses"].(map[string]any)
		for _, status := range options.GetAdditionalSuccessStatus {
			responses[status] = map[string]any{
				"description": "also ok",
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"items": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}}},
					},
				},
			}
		}
		var err error
		getRaw, err = json.Marshal(getDocument)
		if err != nil {
			t.Fatal(err)
		}
	}
	if options.WildcardCheckStatus {
		getRaw = bytes.Replace(getRaw, []byte(`"200"`), []byte(`"2XX"`), 1)
	}
	postStatus := options.PostSuccessStatus
	if postStatus == "" {
		postStatus = "201"
	}
	postSchema := map[string]any{"type": "object", "additionalProperties": false, "required": []any{"title"}, "properties": map[string]any{"title": map[string]any{"type": "string", "maxLength": 64}}}
	if options.OpenWriteBody {
		delete(postSchema, "additionalProperties")
	}
	postContent := map[string]any{"application/json": map[string]any{"schema": postSchema}}
	if options.MultipleMedia {
		postContent["application/xml"] = map[string]any{"schema": postSchema}
	}
	postDocument := map[string]any{
		"openapi": "3.0.3",
		"info":    map[string]any{"title": "alpha", "version": "1"},
		"paths": map[string]any{
			"/widgets": map[string]any{
				"post": map[string]any{
					"operationId": "createWidget",
					"requestBody": map[string]any{"required": true, "content": postContent},
					"responses": map[string]any{
						postStatus: map[string]any{
							"description": "created",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{"type": "object", "additionalProperties": false},
								},
							},
						},
					},
				},
			},
		},
	}
	postRaw, err := json.Marshal(postDocument)
	if err != nil {
		t.Fatal(err)
	}
	getID, postID := "alpha.rest.get.shared", "alpha.rest.post.shared"
	getBindingStatuses := []any{getStatus}
	if len(options.GetBindingSuccessStatuses) != 0 {
		getBindingStatuses = make([]any, len(options.GetBindingSuccessStatuses))
		for i, status := range options.GetBindingSuccessStatuses {
			getBindingStatuses[i] = status
		}
	}
	plan := map[string]any{
		"metadata": map[string]any{"display_name": "Alpha", "description": "Alpha fixture connector.", "integration_type": "api", "release_stage": "preview", "docs_url": "https://docs.example.invalid/alpha"},
		"config": map[string]any{"properties": []any{
			map[string]any{"name": "base_url", "type": "string", "format": "uri", "description": "Provider origin.", "default": "https://api.example.invalid"},
			map[string]any{"name": "token", "type": "string", "description": "Bearer token.", "secret": true},
		}},
		"auth":   map[string]any{"mode": "bearer", "token": "{{ secrets.token }}"},
		"server": map[string]any{"url": "{{ config.base_url }}", "user_agent": "alpha-source-lock/1"},
		"check":  map[string]any{"source_id": getID, "method": "GET", "path": "/widgets", "success_statuses": getBindingStatuses},
		"operations": []any{
			map[string]any{"source_id": getID, "state": "materialized", "citation": map[string]any{"document_id": "get", "location": `paths["/widgets"].get`}, "binding": map[string]any{"kind": "direct_read", "id": "list_widgets", "command_path": "widgets list", "command_summary": "List widgets", "output_policy": "json_redacted", "max_response_bytes": 4096, "success_statuses": getBindingStatuses, "risk": "low"}},
			map[string]any{"source_id": postID, "state": "materialized", "citation": map[string]any{"document_id": "post", "location": `paths["/widgets"].post`}, "inputs": []any{map[string]any{"source": "body.title", "target": "record.title"}}, "binding": map[string]any{"kind": "write", "id": "create_widget", "success_statuses": []any{postStatus}, "request_media": "application/json", "write_kind": "create", "mutation_class": "create", "approval": "destructive", "risk": "high"}},
		},
	}
	if options.MissingBodyBinding {
		plan["operations"].([]any)[1].(map[string]any)["inputs"] = []any{}
	}
	if options.BlockOpenWrite {
		postOperation := plan["operations"].([]any)[1].(map[string]any)
		postOperation["state"] = "blocked"
		postOperation["reason"] = "Runtime cannot faithfully materialize an open provider object; blocked pending explicit typed body support."
		delete(postOperation, "binding")
		delete(postOperation, "inputs")
	}
	if options.BlockPostForStatusRange {
		postOperation := plan["operations"].([]any)[1].(map[string]any)
		postOperation["state"] = "blocked"
		postOperation["reason"] = "runtime gap: exact response-status range execution"
		delete(postOperation, "binding")
		delete(postOperation, "inputs")
	}
	if options.BlockedWriteWithInputs {
		postOperation := plan["operations"].([]any)[1].(map[string]any)
		postOperation["state"] = "blocked"
		postOperation["reason"] = "Blocked fixture row must not retain executable input bindings."
		delete(postOperation, "binding")
	}
	if options.BlockedCheck {
		getOperation := plan["operations"].([]any)[0].(map[string]any)
		getOperation["state"] = "blocked"
		getOperation["reason"] = "Blocked fixture check is accounted but not executable."
		delete(getOperation, "binding")
	}
	if options.WildcardCheckStatus {
		plan["check"].(map[string]any)["success_statuses"] = []any{"2XX"}
		plan["operations"].([]any)[0].(map[string]any)["binding"].(map[string]any)["success_statuses"] = []any{"2XX"}
	}
	if options.WrongRequestMedia {
		plan["operations"].([]any)[1].(map[string]any)["binding"].(map[string]any)["request_media"] = "application/xml"
	}
	if options.DuplicateAccounting {
		plan["operations"] = append(plan["operations"].([]any), plan["operations"].([]any)[0])
	}
	if options.IntegrationType != "" {
		plan["metadata"].(map[string]any)["integration_type"] = options.IntegrationType
	}
	if options.UnknownMaterializationField {
		plan["unexpected"] = true
	}
	lock := sourceMaterializeV4Lock(t, getRaw, postRaw, plan)
	if options.Legacy {
		lock = sourceImportV3FixtureLock(t, "alpha", []sourceImportV3FixtureDocument{{ID: "get", Path: "/widgets", OperationID: "listWidgets", Artifact: getRaw}})
	}
	if err := os.WriteFile(filepath.Join(sourcesDir, "alpha-operation-source-lock.json"), lock, 0o644); err != nil {
		t.Fatal(err)
	}
	if options.OutputSymlink {
		if err := os.Symlink(filepath.Join(defsRoot, "outside"), filepath.Join(defsRoot, "alpha", "metadata.json")); err != nil {
			t.Fatal(err)
		}
	}
	fetcher := sourceImportFetchFunc(func(_ context.Context, sourceURL string) ([]byte, error) {
		switch sourceURL {
		case "https://fixtures.polymetrics.invalid/get.openapi.json":
			return getRaw, nil
		case "https://fixtures.polymetrics.invalid/post.openapi.json":
			return postRaw, nil
		default:
			return nil, os.ErrNotExist
		}
	})
	return defsRoot, lock, fetcher
}

func sourceMaterializeV4Lock(t *testing.T, getRaw, postRaw []byte, plan map[string]any) []byte {
	t.Helper()
	document := func(id, method, operationID string, raw []byte) map[string]any {
		lockRaw := sourceImportV3FixtureLock(t, "alpha", []sourceImportV3FixtureDocument{{ID: id, Path: "/widgets", Method: method, OperationID: operationID, Artifact: raw}})
		var lock map[string]any
		if err := json.Unmarshal(lockRaw, &lock); err != nil {
			t.Fatal(err)
		}
		return lock["rest"].(map[string]any)["source_documents"].([]any)[0].(map[string]any)
	}
	lock := map[string]any{"schema_version": 4, "connector": "alpha", "rest": map[string]any{"retrieval": "hermetic v4 fixture capture", "openapi": []any{"3.0.3"}, "source_documents": []any{document("get", "GET", "listWidgets", getRaw), document("post", "POST", "createWidget", postRaw)}}, "counts": map[string]any{"rest": 2, "graphql_query": 0, "graphql_mutation": 0, "total": 2}, "materialization": plan}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func sourceMaterializeOwnedBytes(t *testing.T, bundleDir string) []byte {
	t.Helper()
	var out bytes.Buffer
	for _, name := range []string{"metadata.json", "spec.json", "streams.json", "writes.json", "operations.json", "api_surface.json", "cli_surface.json", "docs.md", "missing-foundation.json", filepath.Join("sources", "alpha-operation-descriptor.json")} {
		raw, err := os.ReadFile(filepath.Join(bundleDir, name))
		if err != nil {
			t.Fatal(err)
		}
		out.WriteString(name)
		out.WriteByte(0)
		out.Write(raw)
		out.WriteByte(0)
	}
	return out.Bytes()
}
