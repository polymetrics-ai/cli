package linkrunner

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Linkrunner Data API resource path
// (relative to base_url), the dotted JSON path to its records array, and the
// record mapper that flattens its objects into a connectors.Record.
type streamEndpoint struct {
	// resource is the path segment under base_url (e.g. "campaigns").
	resource string
	// recordsPath is the dotted path to the records array in the response body
	// (e.g. "data.campaigns").
	recordsPath string
	// mapRecord flattens a raw Linkrunner object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// linkrunnerStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in linkrunnerStreams.
var linkrunnerStreamEndpoints = map[string]streamEndpoint{
	"campaigns":        {resource: "campaigns", recordsPath: "data.campaigns", mapRecord: linkrunnerCampaignRecord},
	"attributed_users": {resource: "attributed-users", recordsPath: "data.users", mapRecord: linkrunnerAttributedUserRecord},
}

// linkrunnerStreams returns the connector's published stream catalog.
func linkrunnerStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "campaigns",
			Description:  "Linkrunner attribution campaigns.",
			PrimaryKey:   []string{"display_id"},
			CursorFields: []string{"update_at"},
			Fields:       linkrunnerCampaignFields(),
		},
		{
			Name:         "attributed_users",
			Description:  "Users attributed to a Linkrunner campaign. Requires a display_id config to scope the read.",
			PrimaryKey:   []string{"campaign_display_id", "attributed_at"},
			CursorFields: []string{"attributed_at"},
			Fields:       linkrunnerAttributedUserFields(),
		},
	}
}

func linkrunnerCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "display_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "update_at", Type: "string"},
		{Name: "google", Type: "boolean"},
		{Name: "meta", Type: "boolean"},
		{Name: "meta_campaign_id", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "default_link", Type: "boolean"},
		{Name: "attributed_users", Type: "number"},
		{Name: "link", Type: "string"},
		{Name: "shareable_link", Type: "string"},
		{Name: "website", Type: "string"},
		{Name: "domain", Type: "string"},
	}
}

func linkrunnerAttributedUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "attributed_at", Type: "string"},
		{Name: "campaign_display_id", Type: "string"},
		{Name: "campaign_name", Type: "string"},
		{Name: "link", Type: "string"},
		{Name: "installed_at", Type: "string"},
		{Name: "store_click_at", Type: "string"},
		{Name: "ad_set_id", Type: "string"},
		{Name: "user_data", Type: "object"},
	}
}

func linkrunnerCampaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"display_id":       item["display_id"],
		"name":             item["name"],
		"created_at":       item["created_at"],
		"update_at":        item["update_at"],
		"google":           item["google"],
		"meta":             item["meta"],
		"meta_campaign_id": item["meta_campaign_id"],
		"active":           item["active"],
		"default_link":     item["default_link"],
		"attributed_users": item["attributed_users"],
		"link":             item["link"],
		"shareable_link":   item["shareable_link"],
		"website":          item["website"],
		"domain":           item["domain"],
	}
}

func linkrunnerAttributedUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"attributed_at":       item["attributed_at"],
		"campaign_display_id": item["campaign_display_id"],
		"campaign_name":       item["campaign_name"],
		"link":                item["link"],
		"installed_at":        item["installed_at"],
		"store_click_at":      item["store_click_at"],
		"ad_set_id":           item["ad_set_id"],
		"user_data":           item["user_data"],
	}
}
