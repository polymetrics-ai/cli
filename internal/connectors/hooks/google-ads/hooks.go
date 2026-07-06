// Package googleads implements the google-ads bundle's StreamHook
// (docs.md "Streams notes"): auth (Bearer + developer-token + optional
// login-customer-id) is fully declarative (streams.json's base.auth/
// base.headers) -- this hook exists purely because two independent shapes
// of the read path cannot be expressed in streams.json alone:
//
//  1. accessible_customers' real response, {"resourceNames": ["customers/123",
//     ...]}, is a bare JSON array of SCALAR STRINGS, not objects --
//     connsdk.RecordsAt (engine/paginate.go, engine/read.go's only
//     record-extraction primitive) silently drops any array element that
//     does not decode as a JSON object, yielding zero records for this shape
//     via the declarative path (the identical gap conventions.md's
//     parity-deviation ledger entry 12 documents for ip2whois's
//     nameservers field).
//  2. campaigns/ad_groups page via a GAQL search POST whose JSON request
//     body carries {"query","pageSize","pageToken"} -- engine/read.go's
//     declarative read path always issues its request with a nil body
//     (engine/bundle.go's StreamSpec.Body is declared but never read), and
//     every one of the engine's 6 pagination types advances by adding a
//     QUERY parameter, never a body field, so an in-body pageToken cannot be
//     expressed regardless of body support.
//
// This ports internal/connectors/google-ads/google_ads.go's
// readAccessibleCustomers/search/mapRecord functions verbatim, reusing
// rt.Requester (the engine's already-built *connsdk.Requester: base URL,
// Bearer auth, and the developer-token/login-customer-id headers are
// already resolved declaratively -- this hook adds only the GAQL body
// construction and the accessible_customers scalar-array walk, never
// touches auth).
package googleads

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	defaultPageSize = 1000
	maxPageSize     = 10000
)

func init() {
	engine.RegisterHooks("google-ads", func() engine.Hooks { return Hooks{} })
}

// Hooks is the google-ads bundle's Tier-2 hook set: StreamHook only (auth is
// fully declarative via streams.json's bearer mode + base.headers -- see
// docs.md "Auth setup").
type Hooks struct{}

func (Hooks) ConnectorName() string { return "google-ads" }

// gaqlStream describes one GAQL search resource (ported verbatim from
// google_ads.go's googleAdsStreamEndpoints): the fixed allow-listed query
// text and the record-mapping function for the results.<resource> nested
// object each row carries.
type gaqlStream struct {
	resource  string
	query     string
	mapRecord func(map[string]any) connectors.Record
}

// gaqlStreams is the routing table for the two GAQL search streams.
var gaqlStreams = map[string]gaqlStream{
	"campaigns": {
		resource:  "campaign",
		query:     "SELECT campaign.id, campaign.name, campaign.status, campaign.resource_name FROM campaign",
		mapRecord: campaignResultRecord,
	},
	"ad_groups": {
		resource:  "ad_group",
		query:     "SELECT ad_group.id, ad_group.name, ad_group.status, ad_group.resource_name FROM ad_group",
		mapRecord: adGroupResultRecord,
	},
}

// ReadStream implements engine.StreamHook, handling all 3 declared streams
// (accessible_customers, campaigns, ad_groups) with handled=true always --
// the declarative streams.json fallback is a structural "shadow" path
// exercised only by conformance's dynamic checks (Hooks=nil there, and every
// stream carries a skip_dynamic marker -- see docs.md "Known limits").
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if err := ctx.Err(); err != nil {
		return true, err
	}

	name := stream.Name
	if name == "" {
		name = "accessible_customers"
	}

	if name == "accessible_customers" {
		return true, h.readAccessibleCustomers(ctx, rt.Requester, emit)
	}

	gs, ok := gaqlStreams[name]
	if !ok {
		return false, nil
	}

	customerID := strings.TrimSpace(req.Config.Config["customer_id"])
	if customerID == "" {
		return true, fmt.Errorf("google-ads connector requires config customer_id for GAQL streams")
	}
	pageSize, err := resolvePageSize(req.Config)
	if err != nil {
		return true, err
	}
	maxPages, err := resolveMaxPages(req.Config)
	if err != nil {
		return true, err
	}
	return true, h.search(ctx, rt.Requester, customerID, gs, pageSize, maxPages, emit)
}

// readAccessibleCustomers ports google_ads.go's readAccessibleCustomers
// verbatim: GET customers:listAccessibleCustomers, then derive customer_id
// from the trailing "/"-delimited segment of each raw resource name (a bare
// scalar string array element -- see package doc reason 1).
func (h Hooks) readAccessibleCustomers(ctx context.Context, r *connsdk.Requester, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, "customers:listAccessibleCustomers", nil, nil)
	if err != nil {
		return fmt.Errorf("read google-ads accessible customers: %w", err)
	}
	var out struct {
		ResourceNames []string `json:"resourceNames"`
	}
	if err := json.NewDecoder(bytes.NewReader(resp.Body)).Decode(&out); err != nil {
		return fmt.Errorf("decode google-ads accessible customers: %w", err)
	}
	for _, rn := range out.ResourceNames {
		if err := ctx.Err(); err != nil {
			return err
		}
		parts := strings.Split(strings.TrimSpace(rn), "/")
		customerID := rn
		if len(parts) > 0 {
			customerID = parts[len(parts)-1]
		}
		if err := emit(connectors.Record{"customer_id": customerID, "resource_name": rn}); err != nil {
			return err
		}
	}
	return nil
}

// search ports google_ads.go's search verbatim: POST
// customers/{customerID}/googleAds:search with a JSON body carrying
// query/pageSize/pageToken, following nextPageToken until it is empty or
// maxPages (0 = unbounded) is reached.
func (h Hooks) search(ctx context.Context, r *connsdk.Requester, customerID string, gs gaqlStream, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := "customers/" + customerID + "/googleAds:search"
	pageToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body := map[string]any{"query": gs.query}
		if pageSize > 0 {
			body["pageSize"] = pageSize
		}
		if pageToken != "" {
			body["pageToken"] = pageToken
		}
		resp, err := r.Do(ctx, http.MethodPost, path, nil, body)
		if err != nil {
			return fmt.Errorf("read google-ads %s: %w", gs.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode google-ads %s: %w", gs.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(gs.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "nextPageToken")
		if err != nil {
			return fmt.Errorf("decode google-ads nextPageToken: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		pageToken = next
	}
	return nil
}

// campaignResultRecord/adGroupResultRecord port google_ads.go's
// googleAdsCampaignResultRecord/googleAdsAdGroupResultRecord verbatim: each
// GAQL search result row nests the requested resource under a
// camelCase-or-lowercase key ("campaign"/"adGroup"), from which
// id/name/status/resourceName are copied.
func campaignResultRecord(item map[string]any) connectors.Record {
	return nestedRecord(item, "campaign")
}

func adGroupResultRecord(item map[string]any) connectors.Record {
	return nestedRecord(item, "adGroup")
}

func nestedRecord(item map[string]any, key string) connectors.Record {
	nested, _ := item[key].(map[string]any)
	if nested == nil {
		nested, _ = item[strings.ToLower(key)].(map[string]any)
	}
	return connectors.Record{
		"id":            nested["id"],
		"name":          nested["name"],
		"status":        nested["status"],
		"resource_name": first(nested["resourceName"], nested["resource_name"]),
	}
}

// first returns the first non-blank/non-nil value, ported verbatim from
// google_ads.go's first helper (a resourceName/resource_name naming
// tolerance).
func first(values ...any) any {
	for _, value := range values {
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		if value != nil {
			return value
		}
	}
	return nil
}

// resolvePageSize/resolveMaxPages port google_ads.go's
// googleAdsPageSize/googleAdsMaxPages+intConfig/maxPagesConfig verbatim:
// page_size is bounded [1,10000] (default 1000); max_pages accepts
// ""/"all"/"unlimited" (unbounded, 0) or a non-negative integer.
func resolvePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-ads config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("google-ads config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func resolveMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-ads config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, fmt.Errorf("google-ads config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}
