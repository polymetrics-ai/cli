package pinterest

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Pinterest API v5 resource path it
// reads from, whether that path is scoped to an ad account, and the record
// mapper that flattens its objects.
//
// Pinterest v5 list endpoints share one shape: {"items":[...],"bookmark":"..."}.
// The next page is requested with ?bookmark=<token>; an empty/absent bookmark
// ends the walk. Adding a stream means adding one entry here plus a Stream
// definition in pinterestStreams.
type streamEndpoint struct {
	// resource is the path template. Account-scoped streams use "%s" where the
	// configured ad account id is substituted (e.g. "ad_accounts/%s/campaigns").
	resource string
	// accountScoped marks streams that live under /ad_accounts/{account_id}/.
	accountScoped bool
	// mapRecord flattens a raw Pinterest object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// pinterestStreamEndpoints is the per-stream routing table. The read path is
// fully data-driven from this table.
var pinterestStreamEndpoints = map[string]streamEndpoint{
	"ad_accounts": {resource: "ad_accounts", mapRecord: adAccountRecord},
	"boards":      {resource: "boards", mapRecord: boardRecord},
	"campaigns":   {resource: "ad_accounts/%s/campaigns", accountScoped: true, mapRecord: campaignRecord},
	"ad_groups":   {resource: "ad_accounts/%s/ad_groups", accountScoped: true, mapRecord: adGroupRecord},
	"audiences":   {resource: "ad_accounts/%s/audiences", accountScoped: true, mapRecord: audienceRecord},
}

// pinterestStreams returns the connector's published stream catalog. Every
// Pinterest object exposes a string id, so the primary key is ["id"]. These
// streams are full-refresh in the upstream connector, so no incremental cursor
// field is published.
func pinterestStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "ad_accounts",
			Description: "Pinterest ad accounts accessible to the authenticated user.",
			PrimaryKey:  []string{"id"},
			Fields:      adAccountFields(),
		},
		{
			Name:        "boards",
			Description: "Pinterest boards owned by the authenticated user.",
			PrimaryKey:  []string{"id"},
			Fields:      boardFields(),
		},
		{
			Name:        "campaigns",
			Description: "Advertising campaigns under the configured ad account.",
			PrimaryKey:  []string{"id"},
			Fields:      campaignFields(),
		},
		{
			Name:        "ad_groups",
			Description: "Ad groups under the configured ad account.",
			PrimaryKey:  []string{"id"},
			Fields:      adGroupFields(),
		},
		{
			Name:        "audiences",
			Description: "Audiences under the configured ad account.",
			PrimaryKey:  []string{"id"},
			Fields:      audienceFields(),
		},
	}
}

func adAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "owner", Type: "object"},
		{Name: "country", Type: "string"},
		{Name: "currency", Type: "string"},
	}
}

func boardFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "owner", Type: "object"},
		{Name: "privacy", Type: "string"},
		{Name: "pin_count", Type: "integer"},
		{Name: "follower_count", Type: "integer"},
		{Name: "created_at", Type: "string"},
	}
}

func campaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "ad_account_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "objective_type", Type: "string"},
		{Name: "created_time", Type: "integer"},
		{Name: "updated_time", Type: "integer"},
	}
}

func adGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "ad_account_id", Type: "string"},
		{Name: "campaign_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_time", Type: "integer"},
		{Name: "updated_time", Type: "integer"},
	}
}

func audienceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "ad_account_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "audience_type", Type: "string"},
		{Name: "size", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "created_timestamp", Type: "integer"},
	}
}

func adAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"name":     item["name"],
		"owner":    item["owner"],
		"country":  item["country"],
		"currency": item["currency"],
	}
}

func boardRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"description":    item["description"],
		"owner":          item["owner"],
		"privacy":        item["privacy"],
		"pin_count":      item["pin_count"],
		"follower_count": item["follower_count"],
		"created_at":     item["created_at"],
	}
}

func campaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"ad_account_id":  item["ad_account_id"],
		"name":           item["name"],
		"status":         item["status"],
		"objective_type": item["objective_type"],
		"created_time":   item["created_time"],
		"updated_time":   item["updated_time"],
	}
}

func adGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"ad_account_id": item["ad_account_id"],
		"campaign_id":   item["campaign_id"],
		"name":          item["name"],
		"status":        item["status"],
		"created_time":  item["created_time"],
		"updated_time":  item["updated_time"],
	}
}

func audienceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"ad_account_id":     item["ad_account_id"],
		"name":              item["name"],
		"audience_type":     item["audience_type"],
		"size":              item["size"],
		"status":            item["status"],
		"created_timestamp": item["created_timestamp"],
	}
}
