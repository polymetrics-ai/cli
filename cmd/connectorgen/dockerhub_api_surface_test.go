package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// dockerhubDocumentedOperations is the operation total re-verified 2026-08-08
// against Docker Hub's own OpenAPI 3.0.3 document
// (https://docs.docker.com/reference/api/hub/latest.yaml — 148,322 bytes,
// sha256 99d9d53c2d93656a3c66d604885abd153dc5df285abc0ecb13802a3bc53d0756):
// 54 unique (METHOD, path) operations across 35 paths (GET 24, POST 12,
// DELETE 6, PATCH 5, PUT 4, HEAD 3).
//
// This artifact's hash and (method, path) set are byte-for-byte identical to
// the "dockerhub-artifact-2026-08-06" artifact already recorded in
// api_surface.json before this phase began — the pre-existing 54-row
// inventory independently re-derives correct, it was not merely trusted (the
// provider ledger has been wrong before, in both directions, per AGENTS.md).
//
// Unlike gorgias/notion (operation_ledger_version 1), dockerhub already
// carries operation_ledger_version 2: every row's provenance lives in a
// per-endpoint `provenance{artifact,source_url}` block plus a top-level
// `artifacts[]`, not in `operation.source_url`. This test's assertions are
// written against that v2 shape, not copy-pasted from the v1 gorgias/notion
// template.
const (
	dockerhubDocumentedOperations = 54
	dockerhubDocumentedGET        = 24
	dockerhubDocumentedPOST       = 12
	dockerhubDocumentedDELETE     = 6
	dockerhubDocumentedPATCH      = 5
	dockerhubDocumentedPUT        = 4
	dockerhubDocumentedHEAD       = 3
)

func TestDockerhubAPISurfaceOperationLedger(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/dockerhub/api_surface.json")
	if err != nil {
		t.Fatalf("read dockerhub api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Artifacts              []struct {
			ID          string `json:"id"`
			URL         string `json:"url"`
			RetrievedAt string `json:"retrieved_at"`
			SHA256      string `json:"sha256"`
		} `json:"artifacts"`
		Endpoints []struct {
			Method     string         `json:"method"`
			Path       string         `json:"path"`
			CoveredBy  map[string]any `json:"covered_by"`
			Excluded   map[string]any `json:"excluded"`
			Provenance *struct {
				Artifact  string `json:"artifact"`
				SourceURL string `json:"source_url"`
			} `json:"provenance"`
			Operation *struct {
				Model            string `json:"model"`
				Status           string `json:"status"`
				Risk             string `json:"risk"`
				BlockedByDefault bool   `json:"blocked_by_default"`
				Reason           string `json:"reason"`
				Notes            string `json:"notes"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal dockerhub api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 2 {
		t.Errorf("operation_ledger_version = %d, want 2 (dockerhub is already v2-provenance)", surface.OperationLedgerVersion)
	}
	if len(surface.Artifacts) == 0 {
		t.Error("artifacts is empty, want at least one v2 provider-artifact record")
	}
	for _, a := range surface.Artifacts {
		if strings.TrimSpace(a.SHA256) == "" {
			t.Errorf("artifact %q has no sha256; the provider artifact must be re-fetched and hashed, not trusted", a.ID)
		}
	}
	if got := len(surface.Endpoints); got != dockerhubDocumentedOperations {
		t.Errorf("api_surface declares %d rows, want %d documented operations", got, dockerhubDocumentedOperations)
	}

	byMethod := map[string]int{}
	seen := map[string]bool{}
	var blank []string
	covered, blocked, legacyExcluded, disallowed := 0, 0, 0, 0

	for _, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Errorf("duplicate endpoint row %q", key)
		}
		seen[key] = true
		byMethod[ep.Method]++

		if ep.Provenance == nil || strings.TrimSpace(ep.Provenance.SourceURL) == "" {
			t.Errorf("%s: v2 row missing provenance.source_url", key)
		}

		dispositions := 0
		if len(ep.CoveredBy) > 0 {
			dispositions++
			covered++
		}
		if ep.Operation != nil {
			dispositions++
			blocked++
			if ep.Operation.Model == "disallowed" {
				disallowed++
				t.Errorf("%s: operation.model is \"disallowed\" — the captain never authorised that classification; every not-yet-implementable operation must carry a real model (direct_read/binary_read/sensitive_reverse_etl/admin_reverse_etl/destructive_action/local_workflow/duplicate/deprecated) plus a named_dependency= note, never a bare disallowed refusal", key)
			}
			if !ep.Operation.BlockedByDefault || ep.Operation.Status != "blocked" {
				t.Errorf("%s: operation row must be blocked_by_default with status blocked", key)
			}
			if strings.TrimSpace(ep.Operation.Reason) == "" {
				t.Errorf("%s: blocked row has no reason", key)
			}
			if !strings.HasPrefix(ep.Operation.Notes, "named_dependency=") {
				t.Errorf("%s: blocked row must name its dependency (notes must start with \"named_dependency=\")", key)
			}
		}
		if len(ep.Excluded) > 0 {
			dispositions++
			legacyExcluded++
		}
		if dispositions == 0 {
			blank = append(blank, key)
		}
		if dispositions > 1 {
			t.Errorf("%s: carries %d dispositions, want exactly 1", key, dispositions)
		}
	}

	if len(blank) > 0 {
		t.Errorf("%d endpoint(s) carry no disposition, want none: %s", len(blank), strings.Join(blank, ", "))
	}
	if legacyExcluded > 0 {
		t.Errorf("%d legacy excluded row(s) remain; operation_ledger_version mode requires operation rows, never the legacy excluded category", legacyExcluded)
	}
	if disallowed > 0 {
		t.Errorf("%d row(s) still carry operation.model \"disallowed\", want 0", disallowed)
	}
	if covered+blocked != dockerhubDocumentedOperations {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d rows", covered, blocked, covered+blocked, dockerhubDocumentedOperations)
	}

	for method, want := range map[string]int{
		"GET":    dockerhubDocumentedGET,
		"POST":   dockerhubDocumentedPOST,
		"DELETE": dockerhubDocumentedDELETE,
		"PATCH":  dockerhubDocumentedPATCH,
		"PUT":    dockerhubDocumentedPUT,
		"HEAD":   dockerhubDocumentedHEAD,
	} {
		if byMethod[method] != want {
			t.Errorf("%s: %d rows, want %d", method, byMethod[method], want)
		}
	}

	// The 4 pre-existing streams plus a representative sample of the newly
	// implemented and newly blocked-with-named-dependency operations.
	for _, key := range []string{
		"GET /v2/namespaces/{namespace}/repositories",
		"GET /v2/namespaces/{namespace}/repositories/{repository}",
		"GET /v2/namespaces/{namespace}/repositories/{repository}/tags",
		"GET /v2/namespaces/{namespace}/repositories/{repository}/tags/{tag}",
		"POST /v2/namespaces/{namespace}/repositories",
		"GET /v2/access-tokens",
		"POST /v2/access-tokens",
		"DELETE /v2/access-tokens/{uuid}",
		"GET /v2/orgs/{name}/access-tokens",
		"GET /v2/auditlogs/{account}",
		"GET /v2/orgs/{org_name}/groups",
		"POST /v2/invites/bulk",
		"GET /v2/orgs/{name}/settings",
		"HEAD /v2/namespaces/{namespace}/repositories/{repository}",
		"POST /v2/auth/token",
		"GET /v2/scim/2.0/Users",
	} {
		if !seen[key] {
			t.Errorf("expected documented endpoint %q", key)
		}
	}

	// Target disposition counts once the phase is GREEN: 4 pre-existing
	// streams + 35 newly implemented (rest_read/write) = 39 covered; 15
	// blocked-with-named-dependency (3 structurally-unexecutable HEAD rows,
	// 3 auth-exchange rows consumed internally by the new dockerhub
	// AuthHook, 9 SCIM rows pending a second SCIM-scoped credential).
	const wantCovered = 39
	const wantBlocked = 15
	if covered != wantCovered {
		t.Errorf("covered = %d, want %d (4 pre-existing streams + 35 newly implemented operations)", covered, wantCovered)
	}
	if blocked != wantBlocked {
		t.Errorf("blocked = %d, want %d (3 HEAD + 3 auth-exchange + 9 SCIM, each carrying a named_dependency= note)", blocked, wantBlocked)
	}
}
