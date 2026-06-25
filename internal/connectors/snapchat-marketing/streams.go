package snapchatmarketing

import "polymetrics.ai/internal/connectors"

// streamSpec describes how a Snapchat Ads API stream is read. The Snapchat
// Marketing API is hierarchical: organizations are top-level, ad accounts hang
// off organizations, and campaigns/ad squads/ads hang off ad accounts. Each list
// response wraps its array under a plural key (e.g. "campaigns") and wraps each
// element under the singular resource name (e.g. {"campaign": {...}}).
type streamSpec struct {
	// arrayKey is the JSON key holding the array of envelope objects, e.g.
	// "organizations", "campaigns".
	arrayKey string
	// itemKey is the singular envelope key wrapping each resource object, e.g.
	// "organization", "campaign". The mapper unwraps this.
	itemKey string
	// scope describes which path the endpoint lives under:
	//   - "organizations": GET /v1/organizations (top-level)
	//   - "adaccounts":     GET /v1/organizations/{org_id}/adaccounts
	//   - "adaccount":      GET /v1/adaccounts/{ad_account_id}/{resource}
	scope string
	// resource is the trailing path segment for ad-account-scoped streams.
	resource string
	// mapRecord projects an already-unwrapped resource object into a Record.
	mapRecord func(map[string]any) connectors.Record
}

// snapchatStreamSpecs is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in snapchatStreams.
var snapchatStreamSpecs = map[string]streamSpec{
	"organizations": {arrayKey: "organizations", itemKey: "organization", scope: "organizations", mapRecord: organizationRecord},
	"adaccounts":    {arrayKey: "adaccounts", itemKey: "adaccount", scope: "adaccounts", mapRecord: adAccountRecord},
	"campaigns":     {arrayKey: "campaigns", itemKey: "campaign", scope: "adaccount", resource: "campaigns", mapRecord: campaignRecord},
	"adsquads":      {arrayKey: "adsquads", itemKey: "adsquad", scope: "adaccount", resource: "adsquads", mapRecord: adSquadRecord},
	"ads":           {arrayKey: "ads", itemKey: "ad", scope: "adaccount", resource: "ads", mapRecord: adRecord},
}

// snapchatStreams returns the connector's published stream catalog. Every
// Snapchat resource exposes a string id and an updated_at timestamp, so the
// primary key is ["id"] and the incremental cursor field is ["updated_at"].
func snapchatStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "organizations",
			Description:  "Snapchat ad organizations the authenticated user can access.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       organizationFields(),
		},
		{
			Name:         "adaccounts",
			Description:  "Snapchat ad accounts under the configured organizations.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       adAccountFields(),
		},
		{
			Name:         "campaigns",
			Description:  "Snapchat advertising campaigns under the configured ad accounts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       campaignFields(),
		},
		{
			Name:         "adsquads",
			Description:  "Snapchat ad squads (ad sets) under the configured ad accounts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       adSquadFields(),
		},
		{
			Name:         "ads",
			Description:  "Snapchat ads under the configured ad accounts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       adFields(),
		},
	}
}

func organizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "address_line_1", Type: "string"},
		{Name: "locality", Type: "string"},
		{Name: "administrative_district_level_1", Type: "string"},
		{Name: "postal_code", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func adAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "organization_id", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "advertiser", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func campaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "ad_account_id", Type: "string"},
		{Name: "objective", Type: "string"},
		{Name: "start_time", Type: "timestamp"},
		{Name: "end_time", Type: "timestamp"},
		{Name: "daily_budget_micro", Type: "integer"},
		{Name: "lifetime_spend_cap_micro", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func adSquadFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "campaign_id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "optimization_goal", Type: "string"},
		{Name: "billing_event", Type: "string"},
		{Name: "bid_micro", Type: "integer"},
		{Name: "daily_budget_micro", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func adFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "ad_squad_id", Type: "string"},
		{Name: "creative_id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "review_status", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func organizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                              item["id"],
		"name":                            item["name"],
		"type":                            item["type"],
		"country":                         item["country"],
		"address_line_1":                  item["address_line_1"],
		"locality":                        item["locality"],
		"administrative_district_level_1": item["administrative_district_level_1"],
		"postal_code":                     item["postal_code"],
		"created_at":                      item["created_at"],
		"updated_at":                      item["updated_at"],
	}
}

func adAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            item["name"],
		"type":            item["type"],
		"status":          item["status"],
		"organization_id": item["organization_id"],
		"currency":        item["currency"],
		"timezone":        item["timezone"],
		"advertiser":      item["advertiser"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
	}
}

func campaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                       item["id"],
		"name":                     item["name"],
		"status":                   item["status"],
		"ad_account_id":            item["ad_account_id"],
		"objective":                item["objective"],
		"start_time":               item["start_time"],
		"end_time":                 item["end_time"],
		"daily_budget_micro":       item["daily_budget_micro"],
		"lifetime_spend_cap_micro": item["lifetime_spend_cap_micro"],
		"created_at":               item["created_at"],
		"updated_at":               item["updated_at"],
	}
}

func adSquadRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"name":               item["name"],
		"status":             item["status"],
		"campaign_id":        item["campaign_id"],
		"type":               item["type"],
		"optimization_goal":  item["optimization_goal"],
		"billing_event":      item["billing_event"],
		"bid_micro":          item["bid_micro"],
		"daily_budget_micro": item["daily_budget_micro"],
		"created_at":         item["created_at"],
		"updated_at":         item["updated_at"],
	}
}

func adRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"status":        item["status"],
		"ad_squad_id":   item["ad_squad_id"],
		"creative_id":   item["creative_id"],
		"type":          item["type"],
		"review_status": item["review_status"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}
