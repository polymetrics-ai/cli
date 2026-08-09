package main

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// workdayRESTOperations is Workday REST's documented operation count, derived
// 2026-08-07 from Workday's own service directory manifest at
// https://community.workday.com/sites/default/files/file-hosting/restapi/services2026.30.json
// (HTTP 200, 617,538 bytes — byte-identical to the sweep derivation, so the
// manifest is reproduced rather than trusted) and the 52 production service
// specs it names, each fetched individually from the same host.
//
// The manifest is a DIRECTORY, not a spec. There is no single API version:
// every service module is independently versioned (v1 through v7) and mounted
// at its own base path, so an operation's identity is the resolved
// (method, base+path) pair, never the service-relative path alone.
//
// 920 is the RAW row count across the 52 specs, and it collapses TWICE.
//
//	920 raw
//	 -4  the same endpoint published by two service modules
//	     (workdayRESTDuplicatedAcrossServices)
//	 -9  query-string variants of an endpoint already counted
//	     (workdayRESTQueryStringVariants)
//	=907 documented operations
const workdayRESTOperations = 907

// workdayRESTRawRows is the count before either dedup, kept so both deltas are
// written down rather than silently absorbed. A future re-derivation that lands
// on 920 has not found thirteen more operations; it has failed to dedup. One
// that lands on 916 has deduped across services and missed the query strings —
// which is exactly what this connector's first derivation did.
const workdayRESTRawRows = 920

// workdayRESTServiceSpecs is how many production service specs the directory
// names. 49 are Swagger 2.0 and 3 are OpenAPI 3.0.1, which matters: the OAS3
// specs carry no `basePath` at all. A reader that looks only at `basePath`
// records an EMPTY base for those three, which is exactly what makes two
// unrelated-looking services collide into four phantom duplicates.
const workdayRESTServiceSpecs = 52

// workdayRESTDuplicatedAcrossServices are the four operations that appear in
// two service specs at once. "Custom Object Data (multi-instance) v2" and
// "Custom Object Data (single-instance) v2" are two directory entries that
// declare the IDENTICAL servers URL (https://<tenantHostname>/customObject/v2)
// and publish the same paths; single- versus multi-instance is a property of
// the custom OBJECT, not of the URL. One endpoint, two doc pages.
//
// They are pinned by name so a re-derivation cannot quietly reintroduce them,
// and so whoever authors these rows re-expresses the losing page's documented
// behaviour rather than dropping it (sweep finding 23).
var workdayRESTDuplicatedAcrossServices = []string{
	"DELETE /customObject/v2/customObjects/{customObjectAlias}/{customObjectID}",
	"GET /customObject/v2/customObjects/{customObjectAlias}/{customObjectID}",
	"POST /customObject/v2/customObjects/{customObjectAlias}",
	"PUT /customObject/v2/customObjects/{customObjectAlias}/{customObjectID}",
}

// workdayRESTLegacyCarriedRows are the rows the shipped bundle already holds,
// and they are NOT among the 916. The bundle's three legacy streams point at
// /ccx/api/hcm/v1/{tenant}/... -- an older Workday HCM REST shape. The current
// service directory publishes no "hcm" service and no /ccx/ path anywhere: its
// worker resources live under staffing/v7, absenceManagement/v5, compensation/v3
// and so on. Nor are they in the archived list, which only holds older versions
// of the 52 listed services.
//
// So they cannot simply be counted as documented, and they must not simply
// vanish either: deleting them would delete three shipped, schema- and
// fixture-backed streams inside a parity commit. They are pinned here and each
// must carry its own disposition, so whoever authors this surface has to make
// that call deliberately -- re-point them at the documented services, or
// disposition them as superseded -- rather than have the count decide it.
var workdayRESTLegacyCarriedRows = []string{
	"GET /ccx/api/hcm/v1/{tenant}/jobs",
	"GET /ccx/api/hcm/v1/{tenant}/organizations",
	"GET /ccx/api/hcm/v1/{tenant}/workers",
	"POST /ccx/api/hcm/v1/{tenant}/workers",
}

// workdayRESTQueryStringVariants are nine rows the provider publishes as their
// own Swagger path keys even though the path is an endpoint already counted
// with a query string bolted on. They are the sweep's recurring double-count
// defect (notion, lever-hiring, help-scout, github) recurring a FIFTH time, and
// this connector's first derivation shipped all nine: it deduped across service
// modules, which caught the four custom-object rows, and never looked for a "?".
//
// Two independent facts settle them as variants rather than operations:
//
//   - Seven carry an EMPTY summary. The provider documents them as an addendum
//     to the base row, not as an operation in their own right. Procurement's
//     base row says so outright — "Retrieves the metadata OR the attachment
//     content of the specified requisition" — one endpoint, two modes.
//   - The two staffing PATCHes carry a summary that describes a BEHAVIOUR of
//     the base endpoint ("...to archived or un-archived"), which is sweep
//     finding 23's shape exactly: a behaviour becomes a flag, never a path.
//
// Every one has its own base-path sibling documented separately, so collapsing
// them loses no endpoint. They are pinned by name because the losing row's
// behaviour must be re-expressed on the surviving command (--type) rather than
// dropped, and because a re-derivation must not quietly reintroduce them.
var workdayRESTQueryStringVariants = []string{
	"GET /accountsPayable/v1/supplierInvoiceRequests/{ID}/attachments/{subresourceID}?type=viewContent",
	"GET /accountsPayable/v1/supplierInvoiceRequests/{ID}/attachments?type=viewContent",
	"GET /procurement/v5/requisitions/{ID}/attachments/{subresourceID}?type=getFileContent",
	"GET /procurement/v5/requisitions/{ID}/attachments?type=getFileContent",
	"GET /recruiting/v4/prospects/{ID}/resumeAttachments/{subresourceID}?type=viewFile",
	"GET /recruiting/v4/prospects/{ID}/resumeAttachments?type=viewFile",
	"PATCH /staffing/v7/workers/{ID}/checkInTopics/{subresourceID}?type=archive",
	"PATCH /staffing/v7/workers/{ID}/checkIns/{subresourceID}?type=archive",
	"POST /api/common/v1/workers/{ID}/businessTitleChanges?type=me",
}

// workdayRESTMethodSplit is the distribution of the 907 documented operations
// after both dedups, counting only rows the current directory documents.
// Workday is read-heavy: 648 GETs against 259 mutations.
var workdayRESTMethodSplit = map[string]int{
	"GET":    648,
	"POST":   152,
	"PATCH":  56,
	"DELETE": 32,
	"PUT":    19,
}

func TestWorkdayRESTDocumentedSurfaceIsComplete(t *testing.T) {
	raw, err := os.ReadFile("../../internal/connectors/defs/workday-rest/api_surface.json")
	if err != nil {
		t.Fatalf("read workday-rest api_surface.json: %v", err)
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
				Notes            string `json:"notes"`
				DuplicateOf      string `json:"duplicate_of"`
			} `json:"operation"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("unmarshal workday-rest api_surface.json: %v", err)
	}

	if surface.OperationLedgerVersion != 1 {
		t.Errorf("operation_ledger_version = %d, want 1", surface.OperationLedgerVersion)
	}

	legacy := map[string]bool{}
	for _, row := range workdayRESTLegacyCarriedRows {
		legacy[row] = true
	}

	byMethod := map[string]int{}
	seen := map[string]bool{}
	var blank, malformed []string
	total, covered, blocked, legacyExcluded, carried := 0, 0, 0, 0, 0

	for _, ep := range surface.Endpoints {
		key := ep.Method + " " + ep.Path
		if seen[key] {
			t.Errorf("duplicate endpoint %q", key)
		}
		seen[key] = true
		// Legacy /ccx/ rows are counted apart from the documented surface, so
		// carrying them can never inflate the parity number.
		if legacy[key] || strings.HasPrefix(ep.Path, "/ccx/") {
			carried++
		} else {
			total++
			byMethod[ep.Method]++
		}

		// The defect class that has now recurred in lever-hiring, help-scout and
		// github: a query-string variant, a wildcard family, or a behaviour
		// encoded into the path is not an endpoint.
		if strings.ContainsAny(ep.Path, " ?*") {
			malformed = append(malformed, key)
		}
		// Workday mounts every service at its own base, so a service-relative
		// path is ambiguous across 52 modules. Rows carry the resolved path.
		if !strings.HasPrefix(ep.Path, "/") {
			malformed = append(malformed, key+" (path is not resolved from the service base)")
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
	if total != workdayRESTOperations {
		t.Errorf("documented endpoints = %d, want %d (%d raw rows across %d service specs, minus %d "+
			"published by two service modules, minus %d query-string variants of an endpoint already "+
			"counted; %d legacy /ccx/ row(s) counted apart)", total, workdayRESTOperations,
			workdayRESTRawRows, workdayRESTServiceSpecs, len(workdayRESTDuplicatedAcrossServices),
			len(workdayRESTQueryStringVariants), carried)
	}
	if carried > len(workdayRESTLegacyCarriedRows) {
		t.Errorf("legacy /ccx/ rows = %d, want at most %d — a new one is not a documented operation",
			carried, len(workdayRESTLegacyCarriedRows))
	}
	if covered+blocked != total+carried {
		t.Errorf("covered(%d)+blocked(%d) = %d, want %d — every row needs a disposition, legacy included",
			covered, blocked, covered+blocked, total+carried)
	}
	if !reflect.DeepEqual(byMethod, workdayRESTMethodSplit) {
		t.Errorf("byMethod = %+v, want %+v", byMethod, workdayRESTMethodSplit)
	}

	// Each collapsed pair must appear exactly once. Asserting presence alone
	// would pass on a surface that reintroduced both copies.
	for _, want := range workdayRESTDuplicatedAcrossServices {
		if !seen[want] {
			t.Errorf("expected the collapsed custom-object row %q exactly once", want)
		}
	}

	// Each query-string variant must be ABSENT (the malformed check above
	// enforces that) while the endpoint it collapses into is PRESENT. Checking
	// only absence would pass on a surface that dropped the endpoint entirely,
	// which is how a double-count gets "fixed" into a missing operation.
	for _, variant := range workdayRESTQueryStringVariants {
		method, rest, ok := strings.Cut(variant, " ")
		if !ok {
			t.Fatalf("malformed variant pin %q", variant)
		}
		base := method + " " + strings.SplitN(rest, "?", 2)[0]
		if !seen[base] {
			t.Errorf("expected %q — the endpoint that %q collapses into", base, variant)
		}
	}

	// One row per service base that the legacy bundle never enumerated. The
	// shipped surface holds four rows against three HCM streams, so a partial
	// re-expansion cannot pass by filling the worker surface again.
	for _, want := range []string{
		"GET /absenceManagement/v5/workers",                       // an independently-versioned v5 service
		"GET /accountsPayable/v1/supplierInvoiceRequests",         // a financials service, not HCM
		"GET /api/prismAnalytics/v3/{tenant}/tables",              // the one service mounted under /api with a tenant variable
		"POST /customObject/v2/customObjects/{customObjectAlias}", // an OAS3 service with no basePath at all
	} {
		if !seen[want] {
			t.Errorf("expected %q — the shipped bundle enumerated only the three legacy HCM read streams", want)
		}
	}
}
