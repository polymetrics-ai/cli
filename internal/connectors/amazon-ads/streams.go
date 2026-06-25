package amazonads

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Amazon Ads API resource path
// (relative to base_url), whether the request must carry the
// Amazon-Advertising-API-Scope (profile) header, and the record mapper that
// flattens its objects. The read path is fully data-driven from this table.
type streamEndpoint struct {
	// resource is the API path segment, e.g. "v2/sp/campaigns".
	resource string
	// scoped is true when the endpoint requires the profile scope header.
	// The profiles endpoint enumerates the scopes themselves, so it is unscoped.
	scoped bool
	// mapRecord flattens a raw Amazon Ads object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// amazonAdsStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in amazonAdsStreams.
var amazonAdsStreamEndpoints = map[string]streamEndpoint{
	"profiles":   {resource: "v2/profiles", scoped: false, mapRecord: profileRecord},
	"campaigns":  {resource: "v2/sp/campaigns", scoped: true, mapRecord: campaignRecord},
	"ad_groups":  {resource: "v2/sp/adGroups", scoped: true, mapRecord: adGroupRecord},
	"portfolios": {resource: "v2/portfolios", scoped: true, mapRecord: portfolioRecord},
	"keywords":   {resource: "v2/sp/keywords", scoped: true, mapRecord: keywordRecord},
}

// amazonAdsStreams returns the connector's published stream catalog. Amazon Ads
// v2 entity endpoints return top-level JSON arrays of objects; each object
// carries an integer id. These are full-refresh (no reliable updated-at cursor
// on the v2 entity endpoints), so CursorFields is empty.
func amazonAdsStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "profiles",
			Description: "Amazon Ads advertising profiles (seller/vendor accounts) the credentials can access.",
			PrimaryKey:  []string{"profile_id"},
			Fields:      profileFields(),
		},
		{
			Name:        "campaigns",
			Description: "Sponsored Products campaigns for the configured profile.",
			PrimaryKey:  []string{"campaign_id"},
			Fields:      campaignFields(),
		},
		{
			Name:        "ad_groups",
			Description: "Sponsored Products ad groups for the configured profile.",
			PrimaryKey:  []string{"ad_group_id"},
			Fields:      adGroupFields(),
		},
		{
			Name:        "portfolios",
			Description: "Portfolios that group campaigns for the configured profile.",
			PrimaryKey:  []string{"portfolio_id"},
			Fields:      portfolioFields(),
		},
		{
			Name:        "keywords",
			Description: "Sponsored Products targeting keywords for the configured profile.",
			PrimaryKey:  []string{"keyword_id"},
			Fields:      keywordFields(),
		},
	}
}

func profileFields() []connectors.Field {
	return []connectors.Field{
		{Name: "profile_id", Type: "integer"},
		{Name: "country_code", Type: "string"},
		{Name: "currency_code", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "daily_budget", Type: "number"},
		{Name: "marketplace_string_id", Type: "string"},
		{Name: "account_type", Type: "string"},
		{Name: "account_name", Type: "string"},
		{Name: "account_id", Type: "string"},
	}
}

func campaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "campaign_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "campaign_type", Type: "string"},
		{Name: "targeting_type", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "daily_budget", Type: "number"},
		{Name: "start_date", Type: "string"},
		{Name: "end_date", Type: "string"},
		{Name: "premium_bid_adjustment", Type: "boolean"},
		{Name: "portfolio_id", Type: "integer"},
	}
}

func adGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ad_group_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "campaign_id", Type: "integer"},
		{Name: "default_bid", Type: "number"},
		{Name: "state", Type: "string"},
	}
}

func portfolioFields() []connectors.Field {
	return []connectors.Field{
		{Name: "portfolio_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "in_budget", Type: "boolean"},
	}
}

func keywordFields() []connectors.Field {
	return []connectors.Field{
		{Name: "keyword_id", Type: "integer"},
		{Name: "campaign_id", Type: "integer"},
		{Name: "ad_group_id", Type: "integer"},
		{Name: "keyword_text", Type: "string"},
		{Name: "match_type", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "bid", Type: "number"},
	}
}

func profileRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"profile_id":    item["profileId"],
		"country_code":  item["countryCode"],
		"currency_code": item["currencyCode"],
		"timezone":      item["timezone"],
		"daily_budget":  item["dailyBudget"],
	}
	if info, ok := item["accountInfo"].(map[string]any); ok {
		rec["marketplace_string_id"] = info["marketplaceStringId"]
		rec["account_type"] = info["type"]
		rec["account_name"] = info["name"]
		rec["account_id"] = info["id"]
	}
	return rec
}

func campaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"campaign_id":            item["campaignId"],
		"name":                   item["name"],
		"campaign_type":          item["campaignType"],
		"targeting_type":         item["targetingType"],
		"state":                  item["state"],
		"daily_budget":           item["dailyBudget"],
		"start_date":             item["startDate"],
		"end_date":               item["endDate"],
		"premium_bid_adjustment": item["premiumBidAdjustment"],
		"portfolio_id":           item["portfolioId"],
	}
}

func adGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ad_group_id": item["adGroupId"],
		"name":        item["name"],
		"campaign_id": item["campaignId"],
		"default_bid": item["defaultBid"],
		"state":       item["state"],
	}
}

func portfolioRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"portfolio_id": item["portfolioId"],
		"name":         item["name"],
		"state":        item["state"],
		"in_budget":    item["inBudget"],
	}
}

func keywordRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"keyword_id":   item["keywordId"],
		"campaign_id":  item["campaignId"],
		"ad_group_id":  item["adGroupId"],
		"keyword_text": item["keywordText"],
		"match_type":   item["matchType"],
		"state":        item["state"],
		"bid":          item["bid"],
	}
}
