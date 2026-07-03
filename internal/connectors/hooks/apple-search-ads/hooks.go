// Package applesearchads is the Tier-2 StreamHook for the apple-search-ads
// defs bundle (docs/migration/conventions.md §1's hook table: "sub-resource
// fan-out reads" / body-carried pagination, the same class of trigger
// monday's and plaid's hooks document): Apple Search Ads exposes two access
// patterns for its core object streams — campaigns are listed with
// GET /campaigns and offset/limit QUERY params, while ad groups, keywords,
// and ads are read org-wide via the POST .../find endpoints, whose selector
// BODY carries {pagination:{offset,limit}}. engine/bundle.go's
// StreamSpec.Body field exists but engine/read.go's declarative read path
// never sends it (the declarative path always issues a nil body), so the
// 3 .../find streams cannot be expressed in streams.json alone. This hook
// handles ALL 4 streams (including campaigns, which COULD be fully
// declarative on its own) for consistency with the bundle's other 3 streams
// and with monday's documented precedent (docs.md's "Streams notes":
// "hooks/monday/hooks.go's ReadStream ports every one of these shapes
// verbatim" even though not every monday stream strictly needed a hook).
//
// This ports internal/connectors/apple-search-ads/apple_search_ads.go's
// harvest loop and internal/connectors/apple-search-ads/streams.go's record
// mappers verbatim, reusing rt.Requester (the engine's already-built
// *connsdk.Requester: base URL/auth/headers/retry/rate-limit plumbing
// already resolved — auth is the declarative oauth2_client_credentials mode,
// no AuthHook needed here).
package applesearchads

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	defaultPageSize = 1000
	maxPageSize     = 1000
)

func init() {
	engine.RegisterHooks("apple-search-ads", func() engine.Hooks { return Hooks{} })
}

// Hooks is apple-search-ads's Tier-2 hook set: StreamHook only (auth stays
// fully declarative oauth2_client_credentials; Check stays the declarative
// base.check GET /campaigns request).
type Hooks struct{}

func (Hooks) ConnectorName() string { return "apple-search-ads" }

var _ engine.StreamHook = Hooks{}

// streamEndpoint mirrors legacy streams.go's streamEndpoint table: the API
// resource, whether it uses the POST .../find selector-body access pattern
// (vs. GET-with-query), and the record mapper that flattens raw camelCase
// fields into the schema's snake_case shape.
type streamEndpoint struct {
	resource  string
	usesFind  bool
	mapRecord func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"campaigns": {resource: "campaigns", usesFind: false, mapRecord: campaignRecord},
	"adgroups":  {resource: "adgroups/find", usesFind: true, mapRecord: adGroupRecord},
	"keywords":  {resource: "targetingkeywords/find", usesFind: true, mapRecord: keywordRecord},
	"ads":       {resource: "ads/find", usesFind: true, mapRecord: adRecord},
}

// ReadStream implements engine.StreamHook, handling all 4 declared streams
// with handled=true always — the declarative streams.json fallback is a
// structural "shadow" path exercised only by conformance's dynamic checks
// (Hooks=nil there, and every stream in this bundle carries its own
// skip_dynamic marker anyway), never here, matching monday's/plaid's
// documented precedent.
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	name := stream.Name
	if name == "" {
		name = "campaigns"
	}
	endpoint, ok := streamEndpoints[name]
	if !ok {
		return false, nil
	}

	pageSize, err := pageSizeFor(req.Config)
	if err != nil {
		return true, err
	}
	maxPages, err := maxPagesFor(req.Config)
	if err != nil {
		return true, err
	}

	return true, harvest(ctx, rt.Requester, endpoint, pageSize, maxPages, emit)
}

// harvest drives Apple Search Ads offset pagination over the {data,
// pagination} envelope, porting legacy apple_search_ads.go's harvest
// verbatim: campaigns paginate with offset/limit query params; the
// .../find streams paginate with a {pagination:{offset,limit}} selector
// body. The loop stops when a short page is returned, when the running
// total reaches pagination.totalResults, or when maxPages is reached.
func harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	seen := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		var (
			resp *connsdk.Response
			err  error
		)
		if endpoint.usesFind {
			body := map[string]any{
				"pagination": map[string]any{
					"offset": offset,
					"limit":  pageSize,
				},
			}
			resp, err = r.Do(ctx, http.MethodPost, endpoint.resource, nil, body)
		} else {
			query := url.Values{}
			query.Set("offset", strconv.Itoa(offset))
			query.Set("limit", strconv.Itoa(pageSize))
			resp, err = r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		}
		if err != nil {
			return fmt.Errorf("read apple-search-ads %s: %w", endpoint.resource, err)
		}

		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode apple-search-ads %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		seen += len(records)

		// Stop on a short page (fewer than requested) — there is no further data.
		if len(records) < pageSize || len(records) == 0 {
			return nil
		}
		// Stop once we have collected the reported total, when present.
		if total, ok := totalResults(resp.Body); ok && seen >= total {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// totalResults reads pagination.totalResults from a response body, if present.
func totalResults(body []byte) (int, bool) {
	raw, err := connsdk.StringAt(body, "pagination.totalResults")
	if err != nil || strings.TrimSpace(raw) == "" {
		return 0, false
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, false
	}
	return v, true
}

func pageSizeFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("apple-search-ads config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("apple-search-ads config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPagesFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("apple-search-ads config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, fmt.Errorf("apple-search-ads config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func campaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                   item["id"],
		"org_id":               item["orgId"],
		"name":                 item["name"],
		"status":               item["status"],
		"serving_status":       item["servingStatus"],
		"display_status":       item["displayStatus"],
		"ad_channel_type":      item["adChannelType"],
		"supply_sources":       item["supplySources"],
		"billing_event":        item["billingEvent"],
		"daily_budget_amount":  item["dailyBudgetAmount"],
		"budget_amount":        item["budgetAmount"],
		"countries_or_regions": item["countriesOrRegions"],
		"creation_time":        item["creationTime"],
		"modification_time":    item["modificationTime"],
		"deleted":              item["deleted"],
	}
}

func adGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"campaign_id":        item["campaignId"],
		"name":               item["name"],
		"status":             item["status"],
		"serving_status":     item["servingStatus"],
		"display_status":     item["displayStatus"],
		"pricing_model":      item["pricingModel"],
		"default_bid_amount": item["defaultBidAmount"],
		"cpa_goal":           item["cpaGoal"],
		"start_time":         item["startTime"],
		"end_time":           item["endTime"],
		"creation_time":      item["creationTime"],
		"modification_time":  item["modificationTime"],
		"deleted":            item["deleted"],
	}
}

func keywordRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"ad_group_id":       item["adGroupId"],
		"campaign_id":       item["campaignId"],
		"text":              item["text"],
		"match_type":        item["matchType"],
		"status":            item["status"],
		"bid_amount":        item["bidAmount"],
		"modification_time": item["modificationTime"],
		"deleted":           item["deleted"],
	}
}

func adRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"ad_group_id":       item["adGroupId"],
		"campaign_id":       item["campaignId"],
		"name":              item["name"],
		"creative_id":       item["creativeId"],
		"creative_type":     item["creativeType"],
		"status":            item["status"],
		"serving_status":    item["servingStatus"],
		"creation_time":     item["creationTime"],
		"modification_time": item["modificationTime"],
		"deleted":           item["deleted"],
	}
}
