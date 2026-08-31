package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeferredVisibilityCLIUsage(t *testing.T) {
	t.Run("help", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if exitCode := run([]string{"deferred-visibility", "--help"}, &stdout, &stderr); exitCode != 0 {
			t.Fatalf("help exit=%d stderr=%q", exitCode, stderr.String())
		}
		if !strings.Contains(stdout.String(), "--check [--json]") || !strings.Contains(stdout.String(), "does not write a descriptor") {
			t.Fatalf("help output = %q", stdout.String())
		}
	})

	t.Run("bad missing check remains validation only", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		if exitCode := run([]string{"deferred-visibility", "cohort.json"}, &stdout, &stderr); exitCode != 2 {
			t.Fatalf("missing-check exit=%d stdout=%q stderr=%q", exitCode, stdout.String(), stderr.String())
		}
		if !strings.Contains(stderr.String(), "--check is required") {
			t.Fatalf("missing-check error = %q", stderr.String())
		}
	})
}

// TestDeferredVisibilityBatchR1Cohort is the red starting point for #4364.
// It asks the real frozen cohort for a source-only report; a successful report
// must make each deferred cell discoverable without creating an executable
// connector artifact.
func TestDeferredVisibilityBatchR1Cohort(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	manifest := filepath.Join(root, "data", "connector-canon", "batch1-source-operation-mapping-cohort.json")

	var stdout, stderr bytes.Buffer
	if exitCode := run([]string{"deferred-visibility", manifest, "--check", "--json"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("deferred-visibility exit=%d stderr=%q", exitCode, stderr.String())
	}
	var secondStdout, secondStderr bytes.Buffer
	if exitCode := run([]string{"deferred-visibility", manifest, "--check", "--json"}, &secondStdout, &secondStderr); exitCode != 0 {
		t.Fatalf("second deferred-visibility exit=%d stderr=%q", exitCode, secondStderr.String())
	}
	if !bytes.Equal(stdout.Bytes(), secondStdout.Bytes()) {
		t.Fatal("deferred-visibility JSON output is not deterministic across identical frozen inputs")
	}

	var report struct {
		SchemaVersion           int  `json:"schema_version"`
		MappingOnly             bool `json:"mapping_only"`
		PrimarySourceOperations int  `json:"primary_source_operations"`
		SupplementalSourceRows  int  `json:"supplemental_source_rows"`
		SourceRows              int  `json:"source_rows"`
		MatrixCells             int  `json:"matrix_cells"`
		DeferredCells           int  `json:"deferred_cells"`
		ExecutableDeclarations  int  `json:"executable_declarations"`
		Entries                 []struct {
			Connector         string `json:"connector"`
			SourceOperationID string `json:"source_operation_id"`
			Lane              string `json:"lane"`
			Visibility        string `json:"visibility"`
			SourceDisposition string `json:"source_disposition"`
			Source            struct {
				SourceLock            string `json:"source_lock"`
				SourceLockOperationID string `json:"source_lock_operation_id"`
				CitationURL           string `json:"citation_url"`
				SourceLocation        string `json:"source_location"`
				Method                string `json:"method"`
				Path                  string `json:"path"`
			} `json:"source"`
			SourceFact json.RawMessage `json:"source_fact"`
			Reason     struct {
				ID     string `json:"id"`
				Kind   string `json:"kind"`
				Detail string `json:"detail"`
			} `json:"reason"`
			Capability struct {
				ID      string `json:"id"`
				AtlasID string `json:"atlas_id"`
			} `json:"capability"`
			RuntimeClaim string `json:"runtime_claim"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode deferred-visibility report: %v\n%s", err, stdout.String())
	}
	if report.SchemaVersion != 1 || !report.MappingOnly || report.ExecutableDeclarations != 0 {
		t.Fatalf("mapping-only report identity = %#v", report)
	}
	if report.PrimarySourceOperations != batch1SourceOperationCount || report.SourceRows != 4343 || report.SupplementalSourceRows != 2 || report.MatrixCells != 30401 {
		t.Fatalf("source denominator accounting = %#v", report)
	}
	if report.DeferredCells == 0 || report.DeferredCells != len(report.Entries) {
		t.Fatalf("deferred report entries=%d deferred_cells=%d", len(report.Entries), report.DeferredCells)
	}

	seen := map[string]bool{}
	sawGitLabDistinctLockIdentity := false
	for _, entry := range report.Entries {
		if entry.Connector == "" || entry.SourceOperationID == "" || !retainedSourceMappingKnownLane(entry.Lane) {
			t.Fatalf("deferred entry has incomplete identity: %#v", entry)
		}
		key := entry.Connector + "\x00" + entry.SourceOperationID + "\x00" + entry.Lane
		if seen[key] {
			t.Fatalf("duplicate deferred entry %q", key)
		}
		seen[key] = true
		if entry.Visibility != "deferred" || (entry.SourceDisposition != "mapped_unproven" && entry.SourceDisposition != "missing_foundation") {
			t.Fatalf("deferred entry has non-deferred disposition: %#v", entry)
		}
		if entry.Source.SourceLock == "" || entry.Source.SourceLockOperationID == "" || entry.Source.CitationURL == "" || entry.Source.SourceLocation == "" || entry.Source.Method == "" || entry.Source.Path == "" || len(entry.SourceFact) == 0 {
			t.Fatalf("deferred entry lost source facts/citation: %#v", entry)
		}
		if entry.Reason.ID == "" || entry.Reason.Kind == "" || entry.Reason.Detail == "" {
			t.Fatalf("deferred entry has unstable or empty reason: %#v", entry)
		}
		if entry.RuntimeClaim != "none" {
			t.Fatalf("deferred entry made a runtime claim: %#v", entry)
		}
		if entry.SourceDisposition == "mapped_unproven" {
			if entry.Reason.ID != "deferred_visibility.mapped_unproven.v1" || entry.Capability.ID != "source.projection-admission.v1" {
				t.Fatalf("mapped-unproven entry has wrong generic authoring prerequisite: %#v", entry)
			}
		} else if entry.Reason.ID != "deferred_visibility.missing_foundation.v1" || entry.Capability.ID == "" || entry.Capability.AtlasID == "" {
			t.Fatalf("missing-foundation entry has no named capability/Atlas reference: %#v", entry)
		}
		if entry.Connector == "gitlab" && entry.SourceOperationID == "gitlab.rest.deleteApiV4AdminActiveContextDeadQueue" && entry.Lane == "direct_write" {
			if entry.Source.SourceLockOperationID != "deleteApiV4AdminActiveContextDeadQueue" {
				t.Fatalf("GitLab source identity did not retain exact cited lock identity: %#v", entry.Source)
			}
			sawGitLabDistinctLockIdentity = true
		}
	}
	if !sawGitLabDistinctLockIdentity {
		t.Fatal("real cohort omitted the GitLab explicit source-to-lock identity proof")
	}
	var rawReport map[string]json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &rawReport); err != nil {
		t.Fatalf("decode raw deferred-visibility report: %v", err)
	}
	for _, forbidden := range []string{"command", "stream", "write", "transport", "credential", "executor", "descriptor"} {
		if _, found := rawReport[forbidden]; found {
			t.Fatalf("source-only report exposes executable root field %q", forbidden)
		}
	}
	var rawEntries []map[string]json.RawMessage
	if err := json.Unmarshal(rawReport["entries"], &rawEntries); err != nil {
		t.Fatalf("decode raw deferred-visibility entries: %v", err)
	}
	for _, entry := range rawEntries {
		for _, forbidden := range []string{"command", "stream", "write", "transport", "credential", "executor", "descriptor", "operation"} {
			if _, found := entry[forbidden]; found {
				t.Fatalf("source-only entry exposes executable field %q", forbidden)
			}
		}
	}
}

func TestDeferredVisibilityResolveSourceOperation(t *testing.T) {
	lock := declarationAdmissionReviewedSourceLock{Operations: map[string]declarationAdmissionReviewedOperation{
		"bitbucket.rest.delete_/addon": {
			ProviderOperationID: "",
			Method:              "DELETE",
			Path:                "/addon",
		},
		"deleteApiV4AdminActiveContextDeadQueue": {
			ProviderOperationID: "deleteApiV4AdminActiveContextDeadQueue",
			Method:              "DELETE",
			Path:                "/api/v4/admin/active_context/dead_queue",
		},
	}}

	t.Run("happy direct lock identity", func(t *testing.T) {
		id, operation, err := deferredVisibilityResolveSourceOperation(deferredVisibilityMatrixRow{
			SourceID: "bitbucket.rest.delete_/addon",
			Method:   "DELETE",
			Path:     "/addon",
		}, lock)
		if err != nil || id != "bitbucket.rest.delete_/addon" || operation.Path != "/addon" {
			t.Fatalf("direct source-lock identity = (%q, %#v, %v)", id, operation, err)
		}
	})

	t.Run("happy declared provider operation identity", func(t *testing.T) {
		id, operation, err := deferredVisibilityResolveSourceOperation(deferredVisibilityMatrixRow{
			SourceID:            "gitlab.rest.deleteApiV4AdminActiveContextDeadQueue",
			ProviderOperationID: "deleteApiV4AdminActiveContextDeadQueue",
			Method:              "DELETE",
			Path:                "/api/v4/admin/active_context/dead_queue",
		}, lock)
		if err != nil || id != "deleteApiV4AdminActiveContextDeadQueue" || operation.ProviderOperationID != "deleteApiV4AdminActiveContextDeadQueue" {
			t.Fatalf("declared provider identity = (%q, %#v, %v)", id, operation, err)
		}
	})

	t.Run("bad absent source has no opaque fallback", func(t *testing.T) {
		_, _, err := deferredVisibilityResolveSourceOperation(deferredVisibilityMatrixRow{
			SourceID: "gitlab.rest.unknown",
			Method:   "DELETE",
			Path:     "/api/v4/admin/active_context/dead_queue",
		}, lock)
		if err == nil || !strings.Contains(err.Error(), "declares no provider operation identity") {
			t.Fatalf("absent source error = %v", err)
		}
	})

	t.Run("edge provider identity cannot cross a route", func(t *testing.T) {
		_, _, err := deferredVisibilityResolveSourceOperation(deferredVisibilityMatrixRow{
			SourceID:            "gitlab.rest.deleteApiV4AdminActiveContextDeadQueue",
			ProviderOperationID: "deleteApiV4AdminActiveContextDeadQueue",
			Method:              "DELETE",
			Path:                "/wrong",
		}, lock)
		if err == nil || !strings.Contains(err.Error(), "route") {
			t.Fatalf("route-mismatch error = %v", err)
		}
	})

	t.Run("edge ambiguous provider identity is rejected", func(t *testing.T) {
		ambiguous := declarationAdmissionReviewedSourceLock{Operations: map[string]declarationAdmissionReviewedOperation{
			"one": {ProviderOperationID: "same", Method: "POST", Path: "/v1/example"},
			"two": {ProviderOperationID: "same", Method: "POST", Path: "/v1/example"},
		}}
		_, _, err := deferredVisibilityResolveSourceOperation(deferredVisibilityMatrixRow{
			SourceID:            "connector.rest.same",
			ProviderOperationID: "same",
			Method:              "POST",
			Path:                "/v1/example",
		}, ambiguous)
		if err == nil || !strings.Contains(err.Error(), "ambiguous") {
			t.Fatalf("ambiguous provider identity error = %v", err)
		}
	})
}

func TestDeferredVisibilitySourceFactsRemainSemanticAndCited(t *testing.T) {
	t.Run("happy explicit source semantic mutation", func(t *testing.T) {
		if !deferredVisibilitySourceSemanticMutation(map[string]any{
			"operation_semantics": map[string]any{"state": "mutation_candidate"},
		}) {
			t.Fatal("explicit source semantic mutation was not recognized")
		}
	})

	t.Run("bad method alone is not a mutation classifier", func(t *testing.T) {
		if deferredVisibilitySourceSemanticMutation(map[string]any{"method": "DELETE"}) {
			t.Fatal("HTTP method alone became a mutation classifier")
		}
	})

	t.Run("edge retained lane source fact is semantic evidence", func(t *testing.T) {
		row := deferredVisibilityMatrixRow{Cells: map[string]deferredVisibilityMatrixCell{
			"direct_write": {Mapping: map[string]any{"source_fact": map[string]any{"classification": "mutation_candidate"}}},
		}}
		if !deferredVisibilityRowSemanticMutation(row) {
			t.Fatal("retained lane source fact was not recognized as semantic mutation evidence")
		}
	})

	t.Run("edge source citation must bind the exact lock node", func(t *testing.T) {
		operation := declarationAdmissionReviewedOperation{
			Protocol:  "rest",
			SourceURL: "https://provider.example.test/openapi.json",
			Location:  `paths["/widgets"].get`,
			Method:    "GET",
			Path:      "/widgets",
		}
		row := deferredVisibilityMatrixRow{
			SourceID:   "provider.rest.listWidgets",
			Method:     "GET",
			Path:       "/widgets",
			SourceFact: map[string]any{"source_summary": "List widgets"},
			Cells:      map[string]deferredVisibilityMatrixCell{},
		}
		if err := deferredVisibilityValidateSourceFact(row, operation); err == nil || !strings.Contains(err.Error(), "citation location") {
			t.Fatalf("missing citation error = %v", err)
		}
		row.SourceFact = map[string]any{"source_location": `paths["/widgets"].get`, "source_url": "https://provider.example.test/openapi.json"}
		if err := deferredVisibilityValidateSourceFact(row, operation); err != nil {
			t.Fatalf("exact source citation rejected: %v", err)
		}
	})
}
