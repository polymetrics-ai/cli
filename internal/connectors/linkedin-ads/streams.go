package linkedinads

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the LinkedIn Marketing API resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects. LinkedIn entity list endpoints return {"elements":[...]} and are paged
// with start/count offset parameters.
type streamEndpoint struct {
	// resource is the LinkedIn REST resource path segment (e.g. "adAccounts").
	resource string
	// mapRecord flattens a raw LinkedIn entity into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// linkedinStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in linkedinStreams; the read
// path is fully data-driven from this table.
var linkedinStreamEndpoints = map[string]streamEndpoint{
	"accounts":        {resource: "adAccounts", mapRecord: linkedinAccountRecord},
	"campaign_groups": {resource: "adCampaignGroups", mapRecord: linkedinCampaignGroupRecord},
	"campaigns":       {resource: "adCampaigns", mapRecord: linkedinCampaignRecord},
	"creatives":       {resource: "creatives", mapRecord: linkedinCreativeRecord},
}

// linkedinStreams returns the connector's published stream catalog. Every
// LinkedIn entity exposes a numeric id, so the primary key is ["id"] across the
// board. lastModified is the LinkedIn change-tracking timestamp used as the
// incremental cursor where available.
func linkedinStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "accounts",
			Description:  "LinkedIn ad accounts the authenticated user can access.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_modified"},
			Fields:       linkedinAccountFields(),
		},
		{
			Name:         "campaign_groups",
			Description:  "LinkedIn ad campaign groups.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_modified"},
			Fields:       linkedinCampaignGroupFields(),
		},
		{
			Name:         "campaigns",
			Description:  "LinkedIn ad campaigns.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_modified"},
			Fields:       linkedinCampaignFields(),
		},
		{
			Name:         "creatives",
			Description:  "LinkedIn ad creatives.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_modified"},
			Fields:       linkedinCreativeFields(),
		},
	}
}

func linkedinAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "reference", Type: "string"},
		{Name: "test", Type: "boolean"},
		{Name: "version", Type: "object"},
		{Name: "created_at", Type: "integer"},
		{Name: "last_modified", Type: "integer"},
	}
}

func linkedinCampaignGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "account", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "total_budget", Type: "object"},
		{Name: "run_schedule", Type: "object"},
		{Name: "created_at", Type: "integer"},
		{Name: "last_modified", Type: "integer"},
	}
}

func linkedinCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "account", Type: "string"},
		{Name: "campaign_group", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "cost_type", Type: "string"},
		{Name: "objective_type", Type: "string"},
		{Name: "format", Type: "string"},
		{Name: "daily_budget", Type: "object"},
		{Name: "unit_cost", Type: "object"},
		{Name: "run_schedule", Type: "object"},
		{Name: "created_at", Type: "integer"},
		{Name: "last_modified", Type: "integer"},
	}
}

func linkedinCreativeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "campaign", Type: "string"},
		{Name: "account", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "is_serving", Type: "boolean"},
		{Name: "intended_status", Type: "string"},
		{Name: "review_status", Type: "object"},
		{Name: "content", Type: "object"},
		{Name: "created_at", Type: "integer"},
		{Name: "last_modified", Type: "integer"},
	}
}

// linkedinAccountRecord flattens an adAccounts element. LinkedIn nests the
// audit timestamps under a "changeAuditStamps" object; we surface the
// created/lastModified epoch millis as flat fields for downstream cursoring.
func linkedinAccountRecord(item map[string]any) connectors.Record {
	created, modified := changeAuditStamps(item)
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"type":          item["type"],
		"status":        item["status"],
		"currency":      item["currency"],
		"reference":     item["reference"],
		"test":          item["test"],
		"version":       item["version"],
		"created_at":    created,
		"last_modified": modified,
	}
}

func linkedinCampaignGroupRecord(item map[string]any) connectors.Record {
	created, modified := changeAuditStamps(item)
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"account":       item["account"],
		"status":        item["status"],
		"total_budget":  item["totalBudget"],
		"run_schedule":  item["runSchedule"],
		"created_at":    created,
		"last_modified": modified,
	}
}

func linkedinCampaignRecord(item map[string]any) connectors.Record {
	created, modified := changeAuditStamps(item)
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"account":        item["account"],
		"campaign_group": item["campaignGroup"],
		"status":         item["status"],
		"type":           item["type"],
		"cost_type":      item["costType"],
		"objective_type": item["objectiveType"],
		"format":         item["format"],
		"daily_budget":   item["dailyBudget"],
		"unit_cost":      item["unitCost"],
		"run_schedule":   item["runSchedule"],
		"created_at":     created,
		"last_modified":  modified,
	}
}

// linkedinCreativeRecord flattens a creatives element. The newer creatives
// endpoint exposes createdAt/lastModifiedAt epoch millis directly rather than
// nesting them under changeAuditStamps.
func linkedinCreativeRecord(item map[string]any) connectors.Record {
	created := item["createdAt"]
	modified := item["lastModifiedAt"]
	if created == nil && modified == nil {
		created, modified = changeAuditStamps(item)
	}
	return connectors.Record{
		"id":              item["id"],
		"campaign":        item["campaign"],
		"account":         item["account"],
		"status":          item["status"],
		"is_serving":      item["isServing"],
		"intended_status": item["intendedStatus"],
		"review_status":   item["reviewStatus"],
		"content":         item["content"],
		"created_at":      created,
		"last_modified":   modified,
	}
}

// changeAuditStamps pulls the created.time and lastModified.time epoch-millis out
// of LinkedIn's nested changeAuditStamps object, returning nil for either when
// absent.
func changeAuditStamps(item map[string]any) (created any, modified any) {
	stamps, ok := item["changeAuditStamps"].(map[string]any)
	if !ok {
		return nil, nil
	}
	if c, ok := stamps["created"].(map[string]any); ok {
		created = c["time"]
	}
	if m, ok := stamps["lastModified"].(map[string]any); ok {
		modified = m["time"]
	}
	return created, modified
}
