package main

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// zendeskSupportOperations is Zendesk Support's documented operation count,
// re-derived 2026-08-07 from the provider's own OpenAPI description at
// https://developer.zendesk.com/zendesk/oas.yaml -- 1,701,930 bytes, openapi
// 3.0.3, info.title "Support API", info.version 2.0.0. The re-fetch returned
// that byte count exactly, so the artifact is reproduced rather than trusted,
// and the derivation was re-run from it rather than carried forward.
//
// THIS IS THE FIRST CONNECTOR IN THE SWEEP WHOSE PROVIDER LEDGER RECONCILES
// EXACTLY. The ledger has been wrong six times (notion +1, intercom -93,
// lever-hiring ~-40, help-scout's 146 query-string double-count, workday-rest
// 920 -> 916 -> 907), which is why every count is re-derived. Zendesk publishes
// a machine-readable spec with one operation per (method, path) key, and the
// derivation lands on the recorded 625 with a zero delta. A reconciling total
// is evidence; it does not make the next ledger trustworthy.
//
// The derivation was run through THIS TEST'S OWN RULES before the number was
// adopted, which is finding 34's lesson applied rather than restated: zero rows
// contain "?", "*" or a space, and zero (method, path) pairs repeat. Neither
// collapse that shrank workday-rest applies here -- the provider publishes no
// query-string variant path keys and no endpoint under two modules.
const zendeskSupportOperations = 625

// zendeskSupportMethodSplit is the distribution of those 625 operations,
// counted from "HTTP Request" path keys under the spec's top-level `paths`,
// never from section headings (finding 17).
var zendeskSupportMethodSplit = map[string]int{
	"GET":    325,
	"POST":   111,
	"PUT":    89,
	"DELETE": 86,
	"PATCH":  14,
}

// zendeskSupportCarriedRows are the six rows the shipped bundle carries that
// are NOT in the Support OAS. They back existing fixture-backed streams over
// Zendesk Guide, community and business-hours resources, whose endpoints live
// in separate Zendesk specs outside this Support-scoped artifact. They are
// counted APART from the documented 625 -- exactly as workday-rest's legacy
// /ccx/ rows are -- so carrying them can never inflate the parity number, and
// so a new one cannot be smuggled in as a documented operation.
var zendeskSupportCarriedRows = []string{
	"GET /api/v2/business_hours/schedules.json",
	"GET /api/v2/community/posts",
	"GET /api/v2/community/topics",
	"GET /api/v2/help_center/categories",
	"GET /api/v2/help_center/incremental/articles",
	"GET /api/v2/help_center/sections",
}

// zendeskSupportCredentialBodyRead is the one read-shaped POST that must stay
// BLOCKED, and it is blocked for a reason no shared foundation will clear.
// POST /api/v2/any_channel/validate_token validates and stores nothing, so it
// classifies as a read — but its documented request body carries a channel
// token and an account push id. Promoting it means authoring a --token flag,
// which is precisely the inline-credential input AGENTS.md forbids ("Add
// credentials from environment variables or stdin, not prompt text"; "Never
// request, print, summarize, or store secret values").
//
// This test therefore requires it to be blocked AND to name the missing
// runtime capability, rather than exempting it. A row that became covered here
// would mean a secret had been turned into a command-line flag.
const zendeskSupportCredentialBodyRead = "POST /api/v2/any_channel/validate_token"

// zendeskSupportReadShapedPOSTs are the eight POSTs that READ. Each queries,
// validates or previews and creates no stored resource, so modelling them as
// reverse-ETL writes would put a plan/preview/approval gate in front of a
// lookup. They are pinned by name so the read-vs-write judgement cannot drift
// silently, and each was re-checked against the artifact when the pin was
// written (finding 32: a spot-pin can name an endpoint the provider does not
// document).
var zendeskSupportReadShapedPOSTs = []string{
	"POST /api/v2/any_channel/validate_token",
	"POST /api/v2/autocomplete/tags",
	"POST /api/v2/custom_objects/{custom_object_key}/records/search",
	"POST /api/v2/it_asset_management/assets/search",
	"POST /api/v2/problems/autocomplete",
	"POST /api/v2/users/autocomplete",
	"POST /api/v2/views/preview",
	"POST /api/v2/views/preview/count",
}

// zendeskSupportBinaryDownload is the ONE operation whose documented success
// response declares application/octet-stream. Binary is read out of the
// artifact, never inferred from a path.
const zendeskSupportBinaryDownload = "GET /api/v2/custom_objects/{custom_object_key}/records/{record_id}/attachments/{id}/download"

// zendeskSupportBinaryTrap is this connector's binary-detection trap, and it
// runs the opposite way to help-scout's and workday-rest's. PUT
// /api/v2/brands/{brand_id} is the only other operation in the whole spec with
// a non-JSON success response: it declares image/jpg AND image/png, because
// updating a brand echoes back the brand logo. It is an UPDATE. A rule that
// looked only for "declares a non-JSON media type" would ship it as a
// download and silently drop the mutation.
const zendeskSupportBinaryTrap = "PUT /api/v2/brands/{brand_id}"

// zendeskSupportUnionRequestBodies are the three operations whose JSON request
// body is rooted at oneOf/anyOf. AGENTS.md is explicit that such a schema is
// not one executable command contract -- runtime preflight expands its arms and
// rejects promotion -- so each must be modelled as separately named reachable
// arms or left non-implemented with the missing capability named. Two of the
// three are also read-shaped POSTs, so their union sits in a search filter
// rather than in a record contract.
var zendeskSupportUnionRequestBodies = []string{
	"POST /api/v2/custom_objects/{custom_object_key}/records/search",
	"POST /api/v2/it_asset_management/assets/search",
	"PUT /api/v2/users/update_many",
}

func TestZendeskSupportDocumentedSurfaceIsComplete(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/zendesk-support/api_surface.json")
	if err != nil {
		t.Fatalf("read zendesk-support api_surface.json: %v", err)
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
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal zendesk-support api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Errorf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	carried := map[string]bool{}
	for _, row := range zendeskSupportCarriedRows {
		carried[row] = true
	}

	byMethod := map[string]int{}
	seen := map[string]bool{}
	coveredBy := map[string]map[string]any{}
	blockedNamesCredentialDependency := map[string]bool{}
	var blank, malformed []string
	documented, carriedSeen, covered, blocked, legacyExcluded := 0, 0, 0, 0, 0

	for _, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Errorf("duplicate endpoint %q", key)
		}
		seen[key] = true

		if carried[key] {
			carriedSeen++
		} else {
			documented++
			byMethod[ep.Method]++
		}

		// The sweep's recurring double-count defect class: a query-string
		// variant, a wildcard family, or a behaviour encoded into the path is
		// not an endpoint. Zendesk's spec ships none of these, and this guard
		// is what keeps that true.
		if strings.ContainsAny(ep.Path, " ?*") {
			malformed = append(malformed, key)
		}
		// Webhook EVENTS are excluded from the operation surface. This
		// Support-scoped artifact contains zero webhook paths of any kind, so
		// a WEBHOOK row here would be an invention.
		if ep.Method == "WEBHOOK" {
			malformed = append(malformed, key+" (webhook EVENT rows are excluded from the operation surface)")
		}

		dispositions := 0
		if len(ep.CoveredBy) > 0 {
			dispositions++
			covered++
			coveredBy[key] = ep.CoveredBy
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
			if strings.TrimSpace(ep.Operation.SourceURL) == "" && strings.TrimSpace(ep.Operation.Notes) == "" {
				t.Errorf("%s: blocked row has no source citation", key)
			}
			// A blanket "blocked until some shared foundation lands" is not a
			// disposition anyone can check. Every remaining block must name the
			// runtime component that refuses this endpoint.
			if !strings.Contains(ep.Operation.Notes, "Named dependency:") &&
				!strings.Contains(ep.Operation.Reason, "Named dependency:") {
				t.Errorf("%s: blocked row must carry a 'Named dependency:' marker", key)
			}
			if strings.Contains(ep.Operation.Notes, "secret input") ||
				strings.Contains(ep.Operation.Notes, "inlining a secret") {
				blockedNamesCredentialDependency[key] = true
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

	sort.Strings(malformed)
	if len(malformed) > 0 {
		t.Errorf("%d malformed row(s): %s", len(malformed), strings.Join(malformed, ", "))
	}
	sort.Strings(blank)
	if len(blank) > 0 {
		t.Errorf("%d endpoint(s) carry no disposition: %s", len(blank), strings.Join(blank, ", "))
	}
	if legacyExcluded != 0 {
		t.Errorf("%d legacy excluded row(s) remain, want 0", legacyExcluded)
	}
	if documented != zendeskSupportOperations {
		t.Errorf("documented endpoints = %d, want %d (%d row(s) carried apart from the Support OAS)",
			documented, zendeskSupportOperations, carriedSeen)
	}
	if carriedSeen != len(zendeskSupportCarriedRows) {
		t.Errorf("carried non-OAS rows = %d, want exactly %d — a new one is not a documented operation",
			carriedSeen, len(zendeskSupportCarriedRows))
	}
	if covered+blocked != documented+carriedSeen {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d — every row needs a disposition, carried rows included",
			covered, blocked, covered+blocked, documented+carriedSeen)
	}
	if !reflect.DeepEqual(byMethod, zendeskSupportMethodSplit) {
		t.Errorf("byMethod = %+v, want %+v", byMethod, zendeskSupportMethodSplit)
	}

	// Parity is reachability, not inventory. A surface that disposed every row
	// by blocking it would satisfy every assertion above while shipping nothing
	// runnable, which is the state this connector was already in: 509 of 631
	// rows blocked behind one shared-foundation sentence. The documented
	// operations Zendesk publishes as executable REST must be covered by a
	// command, and only a named runtime gap may block one.
	if covered < zendeskSupportOperations {
		t.Errorf("covered rows = %d, want at least %d — a blocked row is a gap, not a disposition",
			covered, zendeskSupportOperations)
	}

	for _, want := range zendeskSupportCarriedRows {
		if !seen[want] {
			t.Errorf("expected carried non-OAS row %q — it backs a shipped fixture-backed stream", want)
		}
	}

	// Read-shaped POSTs must be covered as READS. Asserting presence alone
	// would pass on a surface that shipped a search behind a reverse-ETL
	// approval gate.
	for _, want := range zendeskSupportReadShapedPOSTs {
		if !seen[want] {
			t.Errorf("expected read-shaped POST %q", want)
			continue
		}
		cb := coveredBy[want]
		if want == zendeskSupportCredentialBodyRead {
			// The security judgement, asserted rather than assumed. Covering
			// this row would mean a channel token became a CLI flag.
			if cb != nil {
				t.Errorf("%s: covered by a command, but its request body carries a channel token; "+
					"promoting it authors an inline --token flag", want)
			}
			if !blockedNamesCredentialDependency[want] {
				t.Errorf("%s: must stay blocked naming the missing secret-input capability, "+
					"not a shared-foundation issue number", want)
			}
			continue
		}
		if cb == nil {
			t.Errorf("%s: read-shaped POST is not covered by a command", want)
			continue
		}
		if _, ok := cb["write"]; ok {
			t.Errorf("%s: read-shaped POST is covered by a WRITE; it queries and stores nothing", want)
		}
		if _, ok := cb["writes"]; ok {
			t.Errorf("%s: read-shaped POST is covered by a WRITE; it queries and stores nothing", want)
		}
	}

	// Binary: exactly one download, and the trap must NOT be one. Checking only
	// that the download exists would pass on a surface that also shipped the
	// brand update as a file fetch.
	if !seen[zendeskSupportBinaryDownload] {
		t.Errorf("expected the custom-object attachment download %q", zendeskSupportBinaryDownload)
	}
	if cb := coveredBy[zendeskSupportBinaryDownload]; cb != nil {
		if _, ok := cb["direct_read"]; !ok {
			t.Errorf("%s: the one application/octet-stream operation must be covered as a read", zendeskSupportBinaryDownload)
		}
	}
	if !seen[zendeskSupportBinaryTrap] {
		t.Errorf("expected %q — the brand update, whose image/* success response is a representation, not a download",
			zendeskSupportBinaryTrap)
	}
	if cb := coveredBy[zendeskSupportBinaryTrap]; cb != nil {
		if _, ok := cb["direct_read"]; ok {
			t.Errorf("%s: modelled as a read because it declares image/jpg and image/png success responses; "+
				"it is a PUT and its mutation would be dropped", zendeskSupportBinaryTrap)
		}
	}

	for _, want := range zendeskSupportUnionRequestBodies {
		if !seen[want] {
			t.Errorf("expected union-bodied operation %q", want)
		}
	}
}
