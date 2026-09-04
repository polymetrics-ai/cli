package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/syncplan"
)

func TestVNextSemanticAdmissionRejectsSwappedGraphQLOperationBeforeWriting(t *testing.T) {
	lock := graphQLSourceLockForSemanticAdmissionTest()
	lock.Operations[0].Commands[0].Command = json.RawMessage(`{"path":"widgets first","summary":"List first widgets","intent":"direct_read","availability":"implemented","operation":"widgets.second","api_surface":[{"method":"POST","path":"/graphql"}],"output_policy":"json_redacted","flags":[]}`)

	root := t.TempDir()
	connectorRoot := filepath.Join(root, lock.Connector)
	if err := os.MkdirAll(connectorRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatalf("marshal source lock: %v", err)
	}
	if err := os.WriteFile(filepath.Join(connectorRoot, "source.lock.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	const sentinel = "must not be replaced"
	outputPath := filepath.Join(connectorRoot, "metadata.json")
	if err := os.WriteFile(outputPath, []byte(sentinel), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr); code != 1 {
		t.Fatalf("runLockRender() = %d, want semantic-admission refusal; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	for _, want := range []string{`source operation "source:widgets.first"`, "/operations/0/commands/0/operation"} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("semantic refusal lacks %q: %s", want, stderr.String())
		}
	}
	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != sentinel {
		t.Fatalf("lock render replaced output after semantic refusal: %q", got)
	}
}

func graphQLSourceLockForSemanticAdmissionTest() vNextSourceLock {
	lock := minimalVNextLockForTest()
	lock.Lanes["etl"] = "unsupported"
	lock.Lanes["direct_read"] = "implemented"
	lock.CLI = json.RawMessage(`{"usage":"pm acme <command>","tagline":"Acme commands"}`)
	lock.Operations = []vNextOperationDescriptor{
		graphQLSourceOperationForSemanticAdmissionTest("source:widgets.first", "widgets.first", "FirstWidgets", "widgets first", 0),
		graphQLSourceOperationForSemanticAdmissionTest("source:widgets.second", "widgets.second", "SecondWidgets", "widgets second", 1),
	}
	return lock
}

func graphQLSourceOperationForSemanticAdmissionTest(sourceID, operationID, operationName, commandPath string, order int) vNextOperationDescriptor {
	return vNextOperationDescriptor{
		ID:             sourceID,
		Source:         json.RawMessage(`{"provider_operation":"` + operationName + `","method":"POST","path":"/graphql","url":"https://provider.example/graphql"}`),
		OperationOrder: order,
		Operation:      json.RawMessage(`{"id":"` + operationID + `","kind":"graphql_query","summary":"List widgets","risk":"low","approval":"none","output_policy":"json_redacted","graphql":{"document":"query ` + operationName + ` { widgets { id } }","operation_name":"` + operationName + `","path":"/graphql","max_bytes":1024,"variables_schema":{"type":"object","properties":{},"additionalProperties":false}}}`),
		Commands: []vNextCommandDescriptor{{
			Order:   order,
			Command: json.RawMessage(`{"path":"` + commandPath + `","summary":"List widgets","intent":"direct_read","availability":"implemented","operation":"` + operationID + `","api_surface":[{"method":"POST","path":"/graphql"}],"output_policy":"json_redacted","flags":[]}`),
		}},
	}
}

func TestVNextSemanticAdmissionStagesRateIdentityAndManifestIndex(t *testing.T) {
	baselineLock := graphQLSourceLockForSemanticAdmissionTest()
	baselineLock.Operations[0].Source = json.RawMessage(`{"provider_operation":"FirstWidgets","method":"POST","path":"/graphql","url":"https://provider.example/graphql","future_provider_fact":{"opaque":true}}`)
	baselineLock.ConfigSchema = json.RawMessage(`{"type":"object","properties":{"account_id":{"type":"string"}}}`)
	baseline, err := canonicalizeVNextSourceLock(baselineLock)
	if err != nil {
		t.Fatalf("canonicalize baseline: %v", err)
	}

	ratedLock := baselineLock
	ratedLock.Execution = map[string]json.RawMessage{"rate_limits.json": json.RawMessage(validVNextRateLimitsForSemanticAdmissionTest())}
	rated, err := canonicalizeVNextSourceLock(ratedLock)
	if err != nil {
		t.Fatalf("canonicalize rated lock: %v", err)
	}
	if len(rated.Staged.Outputs["rate_limits.json"]) == 0 {
		t.Fatal("semantic stage omitted declared rate_limits.json")
	}
	if rated.Staged.Identity.Digest == baseline.Staged.Identity.Digest {
		t.Fatal("adding declared rate limits did not change staged execution identity")
	}
	indexed, found := rated.Staged.Index.Lookup(rated.Connector)
	if !found || indexed.Digest != rated.Staged.Identity.Digest || indexed.Generation != rated.Staged.Identity.Generation || indexed.Bytes != rated.Staged.Identity.Bytes {
		t.Fatalf("staged manifest index = %#v, found=%t; want loaded execution identity %#v", indexed, found, rated.Staged.Identity)
	}
	repeated, err := canonicalizeVNextSourceLock(ratedLock)
	if err != nil {
		t.Fatalf("canonicalize repeated rated lock: %v", err)
	}
	if repeated.Staged.Identity != rated.Staged.Identity || !reflect.DeepEqual(repeated.Staged.Provenance, rated.Staged.Provenance) {
		t.Fatal("unchanged source lock did not retain deterministic staged identity and provenance")
	}
	if !sameJSON(rated.Graph.Operations[0].Source, baselineLock.Operations[0].Source) {
		t.Fatalf("semantic admission rewrote an unknown provider source fact: %s", rated.Graph.Operations[0].Source)
	}

	changedLock := ratedLock
	changedLock.Execution = map[string]json.RawMessage{"rate_limits.json": json.RawMessage(strings.Replace(validVNextRateLimitsForSemanticAdmissionTest(), `"limit":100`, `"limit":101`, 1))}
	changed, err := canonicalizeVNextSourceLock(changedLock)
	if err != nil {
		t.Fatalf("canonicalize changed-rate lock: %v", err)
	}
	if changed.Staged.Identity.Digest == rated.Staged.Identity.Digest {
		t.Fatal("changing declared rate limits did not change staged execution identity")
	}

	removedLock := ratedLock
	removedLock.Execution = nil
	removed, err := canonicalizeVNextSourceLock(removedLock)
	if err != nil {
		t.Fatalf("canonicalize rate-removed lock: %v", err)
	}
	if removed.Staged.Identity != baseline.Staged.Identity {
		t.Fatalf("removing rate limits did not restore the baseline identity: got %#v want %#v", removed.Staged.Identity, baseline.Staged.Identity)
	}
}

func TestVNextSemanticAdmissionRejectsMalformedRateBeforeWriting(t *testing.T) {
	lock := graphQLSourceLockForSemanticAdmissionTest()
	lock.ConfigSchema = json.RawMessage(`{"type":"object","properties":{"account_id":{"type":"string"}}}`)
	lock.Execution = map[string]json.RawMessage{"rate_limits.json": json.RawMessage(`{"schema_version":1,"state":"declared","policies":[{"id":"requests","source":{"url":"https://docs.example.test/rate-limits","retrieved_at":"2026-08-05"},"selector":{"all":true},"scope":{"subject_kind":"account","subject_config":"missing_config"},"budgets":[{"model":"fixed_window","dimension":"sustained","unit":"requests","limit":100,"window_seconds":60}]}]}`)}

	root := t.TempDir()
	connectorRoot := filepath.Join(root, lock.Connector)
	if err := os.MkdirAll(connectorRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(lock)
	if err != nil {
		t.Fatalf("marshal source lock: %v", err)
	}
	if err := os.WriteFile(filepath.Join(connectorRoot, "source.lock.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}
	const sentinel = "must not be replaced"
	outputPath := filepath.Join(connectorRoot, "metadata.json")
	if err := os.WriteFile(outputPath, []byte(sentinel), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr); code != 1 {
		t.Fatalf("runLockRender() = %d, want rate-admission refusal; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "rate_limits.json") {
		t.Fatalf("rate-admission refusal lacks rate artifact identity: %s", stderr.String())
	}
	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != sentinel {
		t.Fatalf("lock render replaced output after malformed rate refusal: %q", got)
	}
}

func TestVNextSemanticAdmissionRejectsCrossConnectorManifestAndResolvesSuppliedSync(t *testing.T) {
	lock := minimalVNextLockForTest()
	canonical, err := canonicalizeVNextSourceLock(lock)
	if err != nil {
		t.Fatalf("canonicalize stream lock: %v", err)
	}

	crossConnector := canonical.Staged.Manifest
	crossConnector.Connector = "other"
	if _, err := admitVNextCanonicalDescriptor(canonical, vNextSemanticAdmissionInput{Manifest: &crossConnector}); err == nil || !strings.Contains(err.Error(), `source connector "acme" field /connector`) {
		t.Fatalf("cross-connector manifest admission error = %v", err)
	}
	unknownExecutor := canonical.Staged.Manifest
	unknownExecutor.Executor = "unregistered/unknown.v1"
	if _, err := admitVNextCanonicalDescriptor(canonical, vNextSemanticAdmissionInput{Manifest: &unknownExecutor}); err == nil || !strings.Contains(err.Error(), "manifest executor") {
		t.Fatalf("unknown-executor manifest admission error = %v", err)
	}

	axes, found := synccontract.ModeAxes(synccontract.ModeFullOverwrite)
	if !found {
		t.Fatal("full-overwrite axes are unavailable")
	}
	plan := syncplan.Plan{
		ContractVersion:  syncplan.ContractVersion,
		Source:           syncplan.BindingRef{Kind: synccontract.BindingKindStream, ID: "widgets"},
		Target:           syncplan.BindingRef{Kind: synccontract.BindingKindAction, ID: "destination.write"},
		Mode:             synccontract.ModeFullOverwrite,
		Axes:             axes,
		GenerationDigest: canonical.Staged.Identity.Digest,
		ArtifactDigest:   vNextSemanticAdmissionDigest("a"),
		EvidenceDigest:   vNextSemanticAdmissionDigest("b"),
		Executors: []syncplan.ExecutorRef{
			{Role: syncplan.ExecutorRoleDestination, ID: "closed_typed/destination.v1", Digest: vNextSemanticAdmissionDigest("c")},
			{Role: syncplan.ExecutorRoleSource, ID: canonical.Staged.Manifest.Executor, Digest: canonical.Staged.Identity.Digest},
		},
		Foundation: syncplan.FoundationRef{ID: "authoring.source-lock-vnext.v1", Digest: vNextSemanticAdmissionDigest("d"), Available: true, Reference: "docs/connector-canon/foundations/catalog.json"},
	}
	wrongGeneration := plan
	wrongGeneration.GenerationDigest = vNextSemanticAdmissionDigest("e")
	if _, err := admitVNextCanonicalDescriptor(canonical, vNextSemanticAdmissionInput{Sync: []vNextSyncAdmission{{SourceID: "stream:widgets", FieldPath: "/operations/0/stream", Plan: wrongGeneration}}}); err == nil || !strings.Contains(err.Error(), "sync generation digest") {
		t.Fatalf("cross-generation sync admission error = %v", err)
	}
	staged, err := admitVNextCanonicalDescriptor(canonical, vNextSemanticAdmissionInput{Sync: []vNextSyncAdmission{{SourceID: "stream:widgets", FieldPath: "/operations/0/stream", Plan: plan}}})
	if err != nil {
		t.Fatalf("admit supplied sync plan: %v", err)
	}
	if len(staged.Sync) != 1 || staged.Sync[0].Result.Kind != syncplan.ResultKindExecutable {
		t.Fatalf("supplied sync admission = %#v, want one executable resolver result", staged.Sync)
	}
}

func validVNextRateLimitsForSemanticAdmissionTest() string {
	return `{"schema_version":1,"state":"declared","policies":[{"id":"requests","source":{"url":"https://docs.example.test/rate-limits","retrieved_at":"2026-08-05"},"selector":{"all":true},"scope":{"subject_kind":"account","subject_config":"account_id"},"budgets":[{"model":"fixed_window","dimension":"sustained","unit":"requests","limit":100,"window_seconds":60}]}]}`
}

func vNextSemanticAdmissionDigest(character string) string {
	return "sha256:" + strings.Repeat(character, 64)
}

func TestVNextSemanticAdmissionRunsResolverAndPreflightBeforeWriting(t *testing.T) {
	tests := []struct {
		name    string
		command json.RawMessage
		want    string
	}{
		{
			name:    "route binding",
			command: json.RawMessage(`{"path":"widgets get","summary":"Get widgets","intent":"direct_read","availability":"implemented","operation":"widgets.get","api_surface":[{"method":"GET","path":"/other"}],"output_policy":"json_redacted","flags":[]}`),
			want:    "runtime binding resolution",
		},
		{
			name:    "flag binding",
			command: json.RawMessage(`{"path":"widgets get","summary":"Get widgets","intent":"direct_read","availability":"implemented","operation":"widgets.get","api_surface":[{"method":"GET","path":"/widgets"}],"output_policy":"json_redacted","flags":[{"name":"bogus","type":"string","maps_to":"unsupported.bogus"}]}`),
			want:    "runtime preflight",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lock := operationDirectReadLockForSemanticAdmissionTest()
			lock.Operations[0].Commands[0].Command = test.command

			root := t.TempDir()
			connectorRoot := filepath.Join(root, lock.Connector)
			if err := os.MkdirAll(connectorRoot, 0o755); err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(lock)
			if err != nil {
				t.Fatalf("marshal source lock: %v", err)
			}
			if err := os.WriteFile(filepath.Join(connectorRoot, "source.lock.json"), raw, 0o600); err != nil {
				t.Fatal(err)
			}
			const sentinel = "must not be replaced"
			outputPath := filepath.Join(connectorRoot, "metadata.json")
			if err := os.WriteFile(outputPath, []byte(sentinel), 0o600); err != nil {
				t.Fatal(err)
			}

			var stdout, stderr bytes.Buffer
			if code := runLockRender([]string{"lock-render", lock.Connector, "--defs", root}, &stdout, &stderr); code != 1 {
				t.Fatalf("runLockRender() = %d, want admission refusal; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
			}
			for _, want := range []string{test.want, `source operation "operation:widgets.get"`, "/operations/0/commands/0"} {
				if !strings.Contains(stderr.String(), want) {
					t.Fatalf("admission refusal lacks %q: %s", want, stderr.String())
				}
			}
			got, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != sentinel {
				t.Fatalf("lock render replaced output after %s refusal: %q", test.name, got)
			}
		})
	}
}

func operationDirectReadLockForSemanticAdmissionTest() vNextSourceLock {
	lock := minimalVNextLockForTest()
	lock.Lanes["etl"] = "unsupported"
	lock.Lanes["direct_read"] = "implemented"
	lock.Operations = []vNextOperationDescriptor{{
		ID:        "operation:widgets.get",
		Operation: json.RawMessage(`{"id":"widgets.get","kind":"rest_read","summary":"Get widgets","risk":"low","approval":"none","output_policy":"json_redacted","rest":{"method":"GET","path":"/widgets","max_bytes":1024,"response":{"success_statuses":["200"]},"parameters":[]}}`),
		Commands: []vNextCommandDescriptor{{
			Order:   0,
			Command: json.RawMessage(`{"path":"widgets get","summary":"Get widgets","intent":"direct_read","availability":"implemented","operation":"widgets.get","api_surface":[{"method":"GET","path":"/widgets"}],"output_policy":"json_redacted","flags":[]}`),
		}},
	}}
	lock.CLI = json.RawMessage(`{"usage":"pm acme <command>","tagline":"Acme commands"}`)
	return lock
}
