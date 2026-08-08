package main

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// jiraOperations is Jira Cloud platform v3's documented operation count,
// re-derived 2026-08-07 from Atlassian's own OpenAPI description at
// https://dac-static.atlassian.com/cloud/jira/platform/swagger-v3.v3.json
// -- openapi 3.0.1, 421 path keys, 617 (method, path) operations.
//
// THE SWEEP'S BYTE-COUNT CHECK CANNOT BE SATISFIED FOR THIS CONNECTOR, AND
// SAYING SO IS THE POINT. Every earlier connector proved its artifact by
// re-fetching identical bytes. Atlassian serves a ROLLING SNAPSHOT: the
// version-pinned URL the ledger records (?_v=1.8516.72) 404s, the unpinned URL
// serves whatever is current, and `info.version` is
// 1001.0.0-SNAPSHOT-<git sha>. Between the master plan's derivation and this
// one -- the SAME calendar day -- the document went 2,445,625 -> 2,449,760
// bytes, 420 -> 421 path keys and 616 -> 617 operations, gaining exactly one
// GET. So a byte match proves nothing here and its absence disproves nothing.
//
// What replaces it is the artifact's sha256, recorded in DERIVED-OPERATIONS.json
// and in api_surface.json's own scope prose:
// 5a51740d7ab3c77c521fc8895a7a58b4ff684bc0d2ebeb830135e8320b063ced.
// A future worker who re-fetches and gets a different sha has learned that the
// snapshot moved, which is a finding about Atlassian rather than about this
// bundle -- exactly the distinction the byte check exists to draw.
//
// The derivation was run through THIS TEST'S OWN RULES before its number was
// adopted (finding 34): zero paths contain "?", "*" or a space, zero (method,
// path) pairs repeat templated, and zero repeat once path variables are
// normalised away. Neither collapse that shrank workday-rest applies.
const jiraOperations = 617

// jiraMethodSplit is the distribution of those 617, counted from (method, path)
// keys under the document's top-level `paths`, never from section headings
// (finding 17). Jira publishes no PATCH, HEAD, OPTIONS or TRACE operation.
var jiraMethodSplit = map[string]int{
	"GET":    276,
	"POST":   134,
	"PUT":    118,
	"DELETE": 89,
}

// jiraCoveredRows and jiraBlockedRows partition the 617. Parity is
// reachability, not inventory (finding 40): the shipped bundle passed for a
// complete review while carrying FIFTEEN rows, twelve of which were
// comma-joined or "and similar" wildcard families standing for hundreds of
// endpoints. A wildcard row is not an operation (finding 24), so this is a
// restructure and the count moves from 15 to 617.
const (
	jiraCoveredRows = 590
	jiraBlockedRows = 27
)

// jiraBlockedClasses is the whole remaining gap, by cause. Jira's spec is
// exhaustive about paths and specific about payloads, which is why the blocked
// set is small -- and every one of the 27 is blocked by something a reader can
// go and check in engine/, not by a shared-foundation issue number.
//
//	unbounded_body   12  bodies Atlassian declares as arbitrary JSON: nine
//	                     entity-property PUTs whose requestBody is literally
//	                     `"schema": {}`, and three JSON-Patch plan updates whose
//	                     schema is an object with no properties. Deriving a
//	                     record contract here means inventing a payload the
//	                     provider never published -- the generic HTTP write
//	                     AGENTS.md forbids.
//	dynamic_key_map   5  field-scheme bodies declared `additionalProperties:
//	                     <object>` and keyed by custom field or scheme id.
//	                     engine's dynamic_fields region is the one declared
//	                     capability for a dynamic-key payload and
//	                     validateDynamicFields accepts SCALAR value_types only,
//	                     so an object-valued map has no bounded contract.
//	raw_binary_body   3  avatar uploads declared `*/*` with no schema at all.
//	                     engine's write body types are json, form, none,
//	                     graphql, json_array, multipart and base64_upload; none
//	                     emits a raw byte stream, and inline raw bytes are
//	                     banned outright.
//	empty_contract    3  no body, no path variable, no required query parameter.
//	                     engine.PreflightWriteAction refuses a record_schema
//	                     admitting only {}, correctly: the action would plan a
//	                     mutation with no inputs.
//	scalar_body       2  bodies that are a bare JSON string. buildJSONBody
//	                     assembles an object from record fields and json_array
//	                     covers a top-level array; nothing emits a top-level
//	                     scalar, and body_type none would send the request with
//	                     the documented value silently dropped.
//	unbindable_read_body
//	                  2  read-shaped POSTs whose required body field `payload`
//	                     is an object. cli_surface flag types are boolean,
//	                     string, integer, number, enum and string_array, so an
//	                     object has no flag form; validate then refuses the
//	                     `implemented` claim and covered_by.direct_read accepts
//	                     only an implemented command. A WRITE may be `partial`
//	                     and still cover its row -- a read may not, which is why
//	                     the same missing capability blocks here and downgrades
//	                     there.
//
// A DELETE is deliberately NOT blocked for "no request body": it is addressed
// by its path, so that is its normal shape, not a missing contract -- the same
// judgement zendesk-support made for the same reason.
var jiraBlockedClasses = map[string]int{
	"unbounded_body":       12,
	"dynamic_key_map":      5,
	"raw_binary_body":      3,
	"empty_contract":       3,
	"scalar_body":          2,
	"unbindable_read_body": 2,
}

// jiraStreamRows are the three operations the shipped bundle already covers as
// ETL streams, each with a hand-authored schema and fixture behind it. They
// stay streams. Converting them to direct reads inside a parity commit would
// delete shipped, contract-backed functionality (finding 21), and asserting
// they are still streams is what stops a regeneration from quietly doing it.
var jiraStreamRows = map[string]string{
	"GET /rest/api/3/search":         "issues",
	"GET /rest/api/3/project/search": "projects",
	"GET /rest/api/3/users/search":   "users",
}

// jiraBinaryDownloads are the ONLY operations whose documented success response
// declares an image media type, and all three are GETs whose own summary says
// they return an avatar image. Binary is read out of the artifact, never
// guessed from the path (finding 35), and it is GET-only (finding 45) -- a
// "declares a non-JSON media type" rule alone would be wrong here in BOTH
// directions, because Atlassian attaches the same content map to every response
// code including 401/403/404.
var jiraBinaryDownloads = []string{
	"GET /rest/api/3/universal_avatar/view/type/{type}",
	"GET /rest/api/3/universal_avatar/view/type/{type}/avatar/{id}",
	"GET /rest/api/3/universal_avatar/view/type/{type}/owner/{entityId}",
}

// jiraBinaryTrap is the counterpart assertion, and it is where jira's binary
// judgement could have gone wrong. POST /rest/api/3/universal_avatar/type/
// {type}/owner/{entityId} UPLOADS an avatar: it is the same resource family,
// its request body is a raw image, and a rule keyed on "avatar" or on a
// non-JSON media type anywhere in the operation would ship it as a download and
// silently drop the mutation. It must be blocked, not covered as a read.
const jiraBinaryTrap = "POST /rest/api/3/universal_avatar/type/{type}/owner/{entityId}"

// jiraReadShapedPOSTs are the 24 POSTs that READ. Each searches, evaluates,
// validates, previews or bulk-fetches and persists nothing, so modelling them
// as reverse-ETL writes would put a plan/preview/approval gate in front of a
// lookup. They are pinned by name because a keyword pass got three of them
// wrong: `bulkSetIssuesPropertiesList` matches "list" and SETS properties,
// while `analyseExpression` and `evaluateJiraExpression` match no read keyword
// at all and mutate nothing. Each was checked against its own description text.
var jiraReadShapedPOSTs = []string{
	"POST /rest/api/3/app/field/context/configuration/list",
	"POST /rest/api/3/changelog/bulkfetch",
	"POST /rest/api/3/comment/list",
	"POST /rest/api/3/expression/analyse",
	"POST /rest/api/3/expression/eval",
	"POST /rest/api/3/expression/evaluate",
	"POST /rest/api/3/issue/bulkfetch",
	"POST /rest/api/3/issue/{issueIdOrKey}/changelog/list",
	"POST /rest/api/3/jql/autocompletedata",
	"POST /rest/api/3/jql/function/computation/search",
	"POST /rest/api/3/jql/match",
	"POST /rest/api/3/jql/parse",
	"POST /rest/api/3/jql/pdcleaner",
	"POST /rest/api/3/jql/sanitize",
	"POST /rest/api/3/priorityscheme/mappings",
	"POST /rest/api/3/search",
	"POST /rest/api/3/search/approximate-count",
	"POST /rest/api/3/search/jql",
	"POST /rest/api/3/workflow/history/list",
	"POST /rest/api/3/workflows/create/validation",
	"POST /rest/api/3/workflows/preview",
	"POST /rest/api/3/workflows/update/validation",
	"POST /rest/api/3/worklog/list",
	"POST /rest/atlassian-connect/1/migration/workflow/rule/search",
}

// jiraUnbindableReadPOSTs are the two read-shaped POSTs that must stay BLOCKED,
// and they are blocked for a reason that is not about read-versus-write at all.
// Both /workflows/create/validation and /workflows/update/validation require a
// body field `payload` that is an OBJECT. Every cli_surface flag type is a
// scalar or a string array, so nothing can carry it; validate then refuses the
// `implemented` claim, and covered_by.direct_read accepts only an implemented
// command.
//
// They are pinned separately from jiraReadShapedPOSTs because the two
// assertions differ: those 22 must be covered as reads, these 2 must be blocked
// naming this capability, and neither may EVER be covered by a write. Folding
// them together would let a future regeneration quietly promote a validation
// endpoint behind a reverse-ETL approval gate, or quietly block one of the 22.
var jiraUnbindableReadPOSTs = []string{
	"POST /rest/api/3/workflows/create/validation",
	"POST /rest/api/3/workflows/update/validation",
}

// jiraWriteShapedPOST is the inverse pin, and it is the one the keyword pass
// actually got wrong. POST /rest/api/3/issue/properties matches "list" in its
// operationId (bulkSetIssuesPropertiesList) and its description reads "Sets or
// updates ... on issues". It must be covered by a WRITE.
const jiraWriteShapedPOST = "POST /rest/api/3/issue/properties"

// jiraBlockedDependencyClass maps a blocked row's note to one of the classes
// above. A note matching none of them is itself a failure: it means a block was
// added without stating which runtime component refuses the endpoint.
func jiraBlockedDependencyClass(notes string) string {
	switch {
	case strings.Contains(notes, "no body type emits a top-level JSON scalar"):
		return "scalar_body"
	case strings.Contains(notes, "none of them emits a raw byte stream"):
		return "raw_binary_body"
	case strings.Contains(notes, "accepts scalar value_types only"):
		return "dynamic_key_map"
	case strings.Contains(notes, "admits only an empty object"):
		return "empty_contract"
	case strings.Contains(notes, "unconstrained JSON value"):
		return "unbounded_body"
	case strings.Contains(notes, "structured-document flag type"):
		return "unbindable_read_body"
	default:
		return ""
	}
}

func TestJiraDocumentedSurfaceIsComplete(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/jira/api_surface.json")
	if err != nil {
		t.Fatalf("read jira api_surface.json: %v", err)
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
		t.Fatalf("unmarshal jira api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Errorf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	byMethod := map[string]int{}
	seen := map[string]bool{}
	coveredBy := map[string]map[string]any{}
	blockedByClass := map[string]int{}
	var blank, malformed []string
	covered, blocked, legacyExcluded := 0, 0, 0

	for _, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Errorf("duplicate endpoint %q", key)
		}
		seen[key] = true
		byMethod[ep.Method]++

		// The sweep's recurring double-count defect class, and the exact shape
		// the SHIPPED bundle was in: a comma-joined family, a wildcard, or a
		// query-string variant is not an endpoint. Twelve of its fifteen rows
		// would fail this guard, which is why the row count moves from 15 to 617.
		if strings.ContainsAny(ep.Path, " ?*,") {
			malformed = append(malformed, key)
		}
		// Webhook EVENTS are excluded from the operation surface; Jira's
		// webhook MANAGEMENT endpoints are in scope and counted as operations.
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
			// disposition anyone can check.
			if !strings.Contains(ep.Operation.Notes, "Named dependency:") &&
				!strings.Contains(ep.Operation.Reason, "Named dependency:") {
				t.Errorf("%s: blocked row must carry a 'Named dependency:' marker", key)
			}
			if class := jiraBlockedDependencyClass(ep.Operation.Notes); class != "" {
				blockedByClass[class]++
			} else {
				t.Errorf("%s: blocked row names no recognised missing capability; state which "+
					"runtime component refuses it", key)
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
	// `excluded` is not one of the three dispositions this sweep accepts
	// (finding 18). The shipped bundle disposed twelve wildcard families that
	// way; each becomes a covered or blocked row naming a real endpoint.
	if legacyExcluded != 0 {
		t.Errorf("%d legacy excluded row(s) remain, want 0", legacyExcluded)
	}
	if len(surface.Endpoints) != jiraOperations {
		t.Errorf("documented endpoints = %d, want %d", len(surface.Endpoints), jiraOperations)
	}
	if covered+blocked != len(surface.Endpoints) {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d — every row needs a disposition",
			covered, blocked, covered+blocked, len(surface.Endpoints))
	}
	if !reflect.DeepEqual(byMethod, jiraMethodSplit) {
		t.Errorf("byMethod = %+v, want %+v", byMethod, jiraMethodSplit)
	}

	// Parity is reachability, not inventory. An exact partition, rather than a
	// `covered >= 617` floor, is what makes a regression to blocked, an
	// unexplained unblock, and a move between blocked classes each fail here.
	if covered != jiraCoveredRows {
		t.Errorf("covered rows = %d, want %d", covered, jiraCoveredRows)
	}
	if blocked != jiraBlockedRows {
		t.Errorf("blocked rows = %d, want %d", blocked, jiraBlockedRows)
	}
	if !reflect.DeepEqual(blockedByClass, jiraBlockedClasses) {
		t.Errorf("blocked rows by named dependency class = %+v, want %+v",
			blockedByClass, jiraBlockedClasses)
	}

	// The three shipped streams stay streams.
	for want, stream := range jiraStreamRows {
		cb := coveredBy[want]
		if cb == nil {
			t.Errorf("%s: shipped ETL stream row is not covered", want)
			continue
		}
		if got, _ := cb["stream"].(string); got != stream {
			t.Errorf("%s: covered_by.stream = %q, want %q — converting a shipped, schema- and "+
				"fixture-backed stream to a direct read inside a parity commit deletes "+
				"functionality", want, got, stream)
		}
	}

	// Binary: exactly these three downloads, and the upload must NOT be one.
	for _, want := range jiraBinaryDownloads {
		cb := coveredBy[want]
		if cb == nil {
			t.Errorf("%s: avatar image read is not covered", want)
			continue
		}
		if _, ok := cb["direct_reads"]; !ok {
			t.Errorf("%s: an image/png success response must be covered as a read", want)
		}
	}
	if !seen[jiraBinaryTrap] {
		t.Errorf("expected %q — the avatar UPLOAD, whose raw image body must never be modelled "+
			"as a download", jiraBinaryTrap)
	}
	if cb := coveredBy[jiraBinaryTrap]; cb != nil {
		t.Errorf("%s: covered as %v, but it uploads an avatar; modelling it as a read drops the "+
			"mutation", jiraBinaryTrap, cb)
	}

	// Read-shaped POSTs must never be covered by a WRITE -- that holds for all
	// 24, including the two that cannot be covered at all. Shipping any of them
	// as a write would put a plan/preview/approval gate in front of a lookup.
	unbindableRead := map[string]bool{}
	for _, name := range jiraUnbindableReadPOSTs {
		unbindableRead[name] = true
	}
	for _, want := range jiraReadShapedPOSTs {
		if !seen[want] {
			t.Errorf("expected read-shaped POST %q", want)
			continue
		}
		cb := coveredBy[want]
		if _, ok := cb["write"]; ok {
			t.Errorf("%s: read-shaped POST is covered by a WRITE; it queries and stores nothing", want)
		}
		if _, ok := cb["writes"]; ok {
			t.Errorf("%s: read-shaped POST is covered by a WRITE; it queries and stores nothing", want)
		}
		if unbindableRead[want] {
			if cb != nil {
				t.Errorf("%s: covered by %v, but its required body field `payload` is an object and "+
					"no cli_surface flag type can carry one", want, cb)
			}
			continue
		}
		if cb == nil {
			t.Errorf("%s: read-shaped POST is not covered by a command", want)
		}
	}

	// ... and the two that ARE blocked must say so for this reason, not by
	// inheriting some other class's note.
	for _, want := range jiraUnbindableReadPOSTs {
		if !seen[want] {
			t.Errorf("expected blocked read-shaped POST %q", want)
		}
	}

	// ... and the POST that a keyword pass would have called a read must be a
	// write. Without this the read/write pin is only half a test.
	if cb := coveredBy[jiraWriteShapedPOST]; cb == nil {
		t.Errorf("%s: expected a covered write", jiraWriteShapedPOST)
	} else if _, ok := cb["writes"]; !ok {
		if _, single := cb["write"]; !single {
			t.Errorf("%s: covered as %v, but Atlassian's description reads \"Sets or updates ... "+
				"on issues\"; its operationId matching \"list\" is a naming coincidence",
				jiraWriteShapedPOST, cb)
		}
	}
}
