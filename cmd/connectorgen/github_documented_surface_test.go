package main

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// githubWebhookEvents is the count of event payloads GitHub documents under the `x-webhooks`
// VENDOR EXTENSION. Because this spec declares `openapi: 3.0.3` it has no native top-level
// `webhooks` object, so tooling that checks only a literal `webhooks` key records zero events for
// the largest connector in the sweep. Events are excluded from the operation surface by policy;
// this constant exists so the number is written down rather than silently lost.
const githubWebhookEvents = 270

func TestGitHubDocumentedRESTSurfaceIsComplete(t *testing.T) {
	lock := loadGitHubSourceLock(t)
	raw, err := os.ReadFile("../../internal/connectors/defs/github/api_surface.json")
	if err != nil {
		t.Fatalf("read github api_surface.json: %v", err)
	}

	var surface struct {
		OperationLedgerVersion int `json:"operation_ledger_version"`
		Endpoints              []struct {
			Method    string         `json:"method"`
			Path      string         `json:"path"`
			CoveredBy map[string]any `json:"covered_by"`
			Excluded  map[string]any `json:"excluded"`
			Operation *struct {
				Model            string `json:"model"`
				Status           string `json:"status"`
				BlockedByDefault bool   `json:"blocked_by_default"`
				Reason           string `json:"reason"`
				SourceURL        string `json:"source_url"`
				Notes            string `json:"notes"`
				DuplicateOf      string `json:"duplicate_of"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal github api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Errorf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	restByMethod := map[string]int{}
	seen := map[string]bool{}
	var blank, synthetic []string
	rest, legacyGraphQLBindings, generatedGraphQLTransports, covered, blocked, legacyExcluded := 0, 0, 0, 0, 0, 0
	lockedREST := githubRESTOperationKeys(lock)

	for _, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Errorf("duplicate endpoint %q", key)
		}
		seen[key] = true

		switch {
		case ep.Method == "GRAPHQL":
			// api_surface retains only legacy fixed-document bindings. The authoritative
			// GraphQL denominator is the source lock's Query/Mutation root inventory.
			legacyGraphQLBindings++
		case githubGeneratedGraphQLTransport(ep.Method, ep.Path, ep.CoveredBy):
			// The generated roots share one physical POST transport. Treat it as
			// GraphQL coverage rather than inventing a REST OpenAPI operation.
			generatedGraphQLTransports++
		case ep.Method == "WEBHOOK":
			// Webhook EVENTS are excluded from the operation surface by the counting policy.
			// GitHub documents githubWebhookEvents of them under `x-webhooks`; its 28 webhook
			// MANAGEMENT operations are ordinary `paths` entries and are part of the 1220.
			t.Errorf("%q is a webhook EVENT row; GitHub's %d x-webhooks events are excluded "+
				"from the operation surface", key, githubWebhookEvents)
		default:
			rest++
			restByMethod[ep.Method]++
			if !lockedREST[key] {
				t.Errorf("%s is not present in the pinned REST source lock", key)
			}
		}

		// A behaviour variant is not an endpoint. The shipped bundle encoded write-action reuse
		// into the path itself — "PATCH /repos/{owner}/{repo}/issues/{issue_number} (close)" and
		// three siblings — which are not documented paths at all. This is the same defect class as
		// a "?async=true" or "?include=…" row: model the variant as a flag or a duplicate
		// disposition, never as a second path.
		if strings.ContainsAny(ep.Path, " ?*") && ep.Method != "GRAPHQL" {
			synthetic = append(synthetic, key)
		}

		dispositions := 0
		if len(ep.CoveredBy) > 0 {
			dispositions++
			covered++
		}
		if ep.Operation != nil {
			dispositions++
			blocked++
			if strings.TrimSpace(ep.Operation.Reason) == "" {
				t.Errorf("%s: blocked row has no reason", key)
			}
			if !ep.Operation.BlockedByDefault {
				t.Errorf("%s: blocked row is not blocked_by_default", key)
			}
			if ep.Operation.Model == "duplicate" && strings.TrimSpace(ep.Operation.DuplicateOf) == "" {
				t.Errorf("%s: duplicate row has no duplicate_of", key)
			}
			named := strings.Contains(ep.Operation.Notes, "Named dependency:") ||
				strings.Contains(ep.Operation.Reason, "Named dependency:") ||
				ep.Operation.Model == "duplicate" ||
				ep.Operation.Model == "deprecated"
			if !named {
				t.Errorf("%s: blocked row must carry a 'Named dependency:' marker", key)
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

	sort.Strings(synthetic)
	if len(synthetic) > 0 {
		t.Errorf("%d synthetic path row(s) are not documented endpoints: %s",
			len(synthetic), strings.Join(synthetic, ", "))
	}
	sort.Strings(blank)
	if len(blank) > 0 {
		t.Errorf("%d endpoint(s) carry no disposition: %s", len(blank), strings.Join(blank, ", "))
	}
	if legacyExcluded != 0 {
		t.Errorf("%d legacy excluded row(s) remain, want 0", legacyExcluded)
	}
	if rest != lock.Counts.REST {
		t.Errorf("REST endpoints = %d, want %d documented operations from source lock", rest, lock.Counts.REST)
	}
	if legacyGraphQLBindings == 0 {
		t.Error("api_surface has no legacy fixed GraphQL bindings; root inventory lives in the source lock")
	}
	if generatedGraphQLTransports != 1 {
		t.Errorf("generated GraphQL transports = %d, want one shared POST /graphql binding", generatedGraphQLTransports)
	}
	if covered+blocked != len(surface.Endpoints) {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d declared bundle bindings", covered, blocked, covered+blocked, len(surface.Endpoints))
	}
	if !reflect.DeepEqual(restByMethod, githubRESTMethodSplit(lock)) {
		t.Errorf("restByMethod = %+v, want %+v from source lock", restByMethod, githubRESTMethodSplit(lock))
	}

	// Spot-pins across the surfaces the shipped bundle never enumerated, one per scope, so a
	// partial re-expansion cannot pass by filling only the repository surface again.
	for _, want := range []string{
		"GET /orgs/{org}", // organization scope
		"GET /user",       // authenticated-user scope
		// Enterprise scope. This pin previously named
		// "GET /enterprises/{enterprise}/copilot/billing/seats", which the
		// artifact does not document at all — Copilot billing is org-scoped
		// (/orgs/{org}/copilot/billing/seats). A pin for an endpoint the
		// provider never published can never pass once the surface is complete,
		// so it is replaced with a real enterprise-scope GET rather than
		// dropped; the scope it was written to guard is still guarded.
		"GET /enterprises/{enterprise}/code-security/configurations",
		"GET /app/hook/config", // GitHub App scope + webhook management
		"POST /markdown",       // a POST that is semantically a read
		"GET /teams/{team_id}", // a deprecated legacy operation, still counted
	} {
		if !seen[want] {
			t.Errorf("expected %q — the shipped bundle enumerated only /repos/{owner}/{repo}/…", want)
		}
	}
}
