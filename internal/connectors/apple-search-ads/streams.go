package applesearchads

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Apple Search Ads API resource it reads
// from, the HTTP method, and the record mapper that flattens its objects.
//
// Apple Search Ads exposes two access patterns for the core object streams:
//   - campaigns are listed with GET /campaigns and offset/limit query params.
//   - ad groups, keywords, and ads are read org-wide via the POST .../find
//     endpoints, which take a selector body carrying {pagination:{offset,limit}}.
//
// Both shapes return the same {data:[...], pagination:{totalResults,...}}
// envelope, so a single offset-driven harvest loop handles both; the only
// difference is GET-with-query vs POST-with-body, captured by usesFind.
type streamEndpoint struct {
	// resource is the API path segment relative to base_url.
	resource string
	// usesFind selects the POST .../find access pattern (pagination in the body)
	// instead of the GET-with-query-params pattern.
	usesFind bool
	// mapRecord flattens a raw Apple object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// appleStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in appleStreams; the read path
// is fully data-driven from this table.
var appleStreamEndpoints = map[string]streamEndpoint{
	"campaigns": {resource: "campaigns", usesFind: false, mapRecord: campaignRecord},
	"adgroups":  {resource: "adgroups/find", usesFind: true, mapRecord: adGroupRecord},
	"keywords":  {resource: "targetingkeywords/find", usesFind: true, mapRecord: keywordRecord},
	"ads":       {resource: "ads/find", usesFind: true, mapRecord: adRecord},
}

// appleStreams returns the connector's published stream catalog. Every Apple
// Search Ads object exposes a numeric id and an RFC3339 modificationTime, so the
// primary key is ["id"] and the incremental cursor field is
// ["modification_time"] across the board.
func appleStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "campaigns",
			Description:  "Apple Search Ads campaigns for the organization.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modification_time"},
			Fields:       campaignFields(),
		},
		{
			Name:         "adgroups",
			Description:  "Apple Search Ads ad groups across all campaigns in the organization.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modification_time"},
			Fields:       adGroupFields(),
		},
		{
			Name:         "keywords",
			Description:  "Apple Search Ads targeting keywords across all ad groups in the organization.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modification_time"},
			Fields:       keywordFields(),
		},
		{
			Name:         "ads",
			Description:  "Apple Search Ads ads across all ad groups in the organization.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modification_time"},
			Fields:       adFields(),
		},
	}
}

func campaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "org_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "serving_status", Type: "string"},
		{Name: "display_status", Type: "string"},
		{Name: "ad_channel_type", Type: "string"},
		{Name: "supply_sources", Type: "object"},
		{Name: "billing_event", Type: "string"},
		{Name: "daily_budget_amount", Type: "object"},
		{Name: "budget_amount", Type: "object"},
		{Name: "countries_or_regions", Type: "object"},
		{Name: "creation_time", Type: "string"},
		{Name: "modification_time", Type: "string"},
		{Name: "deleted", Type: "boolean"},
	}
}

func adGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "campaign_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "serving_status", Type: "string"},
		{Name: "display_status", Type: "string"},
		{Name: "pricing_model", Type: "string"},
		{Name: "default_bid_amount", Type: "object"},
		{Name: "cpa_goal", Type: "object"},
		{Name: "start_time", Type: "string"},
		{Name: "end_time", Type: "string"},
		{Name: "creation_time", Type: "string"},
		{Name: "modification_time", Type: "string"},
		{Name: "deleted", Type: "boolean"},
	}
}

func keywordFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "ad_group_id", Type: "integer"},
		{Name: "campaign_id", Type: "integer"},
		{Name: "text", Type: "string"},
		{Name: "match_type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "bid_amount", Type: "object"},
		{Name: "modification_time", Type: "string"},
		{Name: "deleted", Type: "boolean"},
	}
}

func adFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "ad_group_id", Type: "integer"},
		{Name: "campaign_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "creative_id", Type: "integer"},
		{Name: "creative_type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "serving_status", Type: "string"},
		{Name: "creation_time", Type: "string"},
		{Name: "modification_time", Type: "string"},
		{Name: "deleted", Type: "boolean"},
	}
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
