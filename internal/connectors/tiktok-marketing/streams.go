package tiktokmarketing

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the TikTok Business API resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects. The read path is fully data-driven from this table.
type streamEndpoint struct {
	// resource is the TikTok endpoint path segment, e.g. "campaign/get/".
	resource string
	// listPath is the dotted JSON path to the records array in the envelope.
	// TikTok wraps lists under data.list for the get endpoints.
	listPath string
	// mapRecord flattens a raw TikTok object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// tiktokStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in tiktokStreams.
var tiktokStreamEndpoints = map[string]streamEndpoint{
	"advertisers": {resource: "advertiser/info/", listPath: "data.list", mapRecord: tiktokAdvertiserRecord},
	"campaigns":   {resource: "campaign/get/", listPath: "data.list", mapRecord: tiktokCampaignRecord},
	"adgroups":    {resource: "adgroup/get/", listPath: "data.list", mapRecord: tiktokAdGroupRecord},
	"ads":         {resource: "ad/get/", listPath: "data.list", mapRecord: tiktokAdRecord},
}

// tiktokStreams returns the connector's published stream catalog. Each TikTok
// management object exposes a string id and a modify_time timestamp, so the
// primary key is the object id and the incremental cursor is modify_time.
func tiktokStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "advertisers",
			Description:  "TikTok advertiser accounts authorized for the access token.",
			PrimaryKey:   []string{"advertiser_id"},
			CursorFields: nil,
			Fields:       tiktokAdvertiserFields(),
		},
		{
			Name:         "campaigns",
			Description:  "TikTok ad campaigns for an advertiser.",
			PrimaryKey:   []string{"campaign_id"},
			CursorFields: []string{"modify_time"},
			Fields:       tiktokCampaignFields(),
		},
		{
			Name:         "adgroups",
			Description:  "TikTok ad groups for an advertiser.",
			PrimaryKey:   []string{"adgroup_id"},
			CursorFields: []string{"modify_time"},
			Fields:       tiktokAdGroupFields(),
		},
		{
			Name:         "ads",
			Description:  "TikTok ads for an advertiser.",
			PrimaryKey:   []string{"ad_id"},
			CursorFields: []string{"modify_time"},
			Fields:       tiktokAdFields(),
		},
	}
}

func tiktokAdvertiserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "advertiser_id", Type: "string"},
		{Name: "advertiser_name", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "role", Type: "string"},
	}
}

func tiktokCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "campaign_id", Type: "string"},
		{Name: "campaign_name", Type: "string"},
		{Name: "advertiser_id", Type: "string"},
		{Name: "objective_type", Type: "string"},
		{Name: "budget", Type: "number"},
		{Name: "budget_mode", Type: "string"},
		{Name: "operation_status", Type: "string"},
		{Name: "create_time", Type: "string"},
		{Name: "modify_time", Type: "string"},
	}
}

func tiktokAdGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "adgroup_id", Type: "string"},
		{Name: "adgroup_name", Type: "string"},
		{Name: "campaign_id", Type: "string"},
		{Name: "advertiser_id", Type: "string"},
		{Name: "placement_type", Type: "string"},
		{Name: "budget", Type: "number"},
		{Name: "budget_mode", Type: "string"},
		{Name: "operation_status", Type: "string"},
		{Name: "create_time", Type: "string"},
		{Name: "modify_time", Type: "string"},
	}
}

func tiktokAdFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ad_id", Type: "string"},
		{Name: "ad_name", Type: "string"},
		{Name: "adgroup_id", Type: "string"},
		{Name: "campaign_id", Type: "string"},
		{Name: "advertiser_id", Type: "string"},
		{Name: "operation_status", Type: "string"},
		{Name: "call_to_action", Type: "string"},
		{Name: "create_time", Type: "string"},
		{Name: "modify_time", Type: "string"},
	}
}

func tiktokAdvertiserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"advertiser_id":   item["advertiser_id"],
		"advertiser_name": item["advertiser_name"],
		"company":         item["company"],
		"status":          item["status"],
		"currency":        item["currency"],
		"timezone":        item["timezone"],
		"country":         item["country"],
		"role":            item["role"],
	}
}

func tiktokCampaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"campaign_id":      item["campaign_id"],
		"campaign_name":    item["campaign_name"],
		"advertiser_id":    item["advertiser_id"],
		"objective_type":   item["objective_type"],
		"budget":           item["budget"],
		"budget_mode":      item["budget_mode"],
		"operation_status": item["operation_status"],
		"create_time":      item["create_time"],
		"modify_time":      item["modify_time"],
	}
}

func tiktokAdGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"adgroup_id":       item["adgroup_id"],
		"adgroup_name":     item["adgroup_name"],
		"campaign_id":      item["campaign_id"],
		"advertiser_id":    item["advertiser_id"],
		"placement_type":   item["placement_type"],
		"budget":           item["budget"],
		"budget_mode":      item["budget_mode"],
		"operation_status": item["operation_status"],
		"create_time":      item["create_time"],
		"modify_time":      item["modify_time"],
	}
}

func tiktokAdRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ad_id":            item["ad_id"],
		"ad_name":          item["ad_name"],
		"adgroup_id":       item["adgroup_id"],
		"campaign_id":      item["campaign_id"],
		"advertiser_id":    item["advertiser_id"],
		"operation_status": item["operation_status"],
		"call_to_action":   item["call_to_action"],
		"create_time":      item["create_time"],
		"modify_time":      item["modify_time"],
	}
}
