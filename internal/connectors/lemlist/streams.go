package lemlist

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the lemlist API resource path (relative
// to base_url) it reads from, the record mapper that flattens its objects, and
// whether the endpoint is paginated. The lemlist API returns records at the JSON
// root (an array for list endpoints, a single object for team), so every stream
// extracts records from the root path "".
type streamEndpoint struct {
	// resource is the lemlist endpoint path segment (e.g. "campaigns").
	resource string
	// paginated is true for offset/limit list endpoints (campaigns, activities,
	// unsubscribes). The team endpoint returns a single object and is not paged.
	paginated bool
	// mapRecord flattens a raw lemlist object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// lemlistStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in lemlistStreams; the read path
// is fully data-driven from this table.
var lemlistStreamEndpoints = map[string]streamEndpoint{
	"team":         {resource: "team", paginated: false, mapRecord: lemlistTeamRecord},
	"campaigns":    {resource: "campaigns", paginated: true, mapRecord: lemlistCampaignRecord},
	"activities":   {resource: "activities", paginated: true, mapRecord: lemlistActivityRecord},
	"unsubscribes": {resource: "unsubscribes", paginated: true, mapRecord: lemlistUnsubscribeRecord},
}

// lemlistStreams returns the connector's published stream catalog. Every lemlist
// object exposes a string `_id`, so the primary key is ["_id"] across the board.
// The lemlist API supports full-refresh only (no incremental cursor), so
// CursorFields is empty.
func lemlistStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "team",
			Description: "The lemlist team (workspace) the API key belongs to.",
			PrimaryKey:  []string{"_id"},
			Fields:      lemlistTeamFields(),
		},
		{
			Name:        "campaigns",
			Description: "lemlist outreach campaigns.",
			PrimaryKey:  []string{"_id"},
			Fields:      lemlistCampaignFields(),
		},
		{
			Name:        "activities",
			Description: "lemlist campaign activities (emails sent, opens, clicks, replies).",
			PrimaryKey:  []string{"_id"},
			Fields:      lemlistActivityFields(),
		},
		{
			Name:        "unsubscribes",
			Description: "Leads who unsubscribed from the team's campaigns.",
			PrimaryKey:  []string{"_id"},
			Fields:      lemlistUnsubscribeFields(),
		},
	}
}

func lemlistTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "createdAt", Type: "timestamp"},
		{Name: "createdBy", Type: "string"},
		{Name: "_updatedAt", Type: "timestamp"},
		{Name: "userIds", Type: "array"},
		{Name: "beta", Type: "array"},
		{Name: "billing", Type: "object"},
		{Name: "revenueVisualization", Type: "object"},
	}
}

func lemlistCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "labels", Type: "array"},
	}
}

func lemlistActivityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "campaignId", Type: "string"},
		{Name: "campaignName", Type: "string"},
		{Name: "companyName", Type: "string"},
		{Name: "createdAt", Type: "timestamp"},
		{Name: "createdBy", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "emailTemplateId", Type: "string"},
		{Name: "emailTemplateName", Type: "string"},
		{Name: "extra", Type: "object"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "icebreaker", Type: "string"},
		{Name: "isFirst", Type: "boolean"},
		{Name: "leadId", Type: "string"},
		{Name: "leadEmail", Type: "string"},
		{Name: "leadFirstName", Type: "string"},
		{Name: "leadLastName", Type: "string"},
		{Name: "linkedinUrl", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "sequenceId", Type: "string"},
		{Name: "sequenceStep", Type: "integer"},
		{Name: "teamId", Type: "string"},
	}
}

func lemlistUnsubscribeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "email", Type: "string"},
	}
}

func lemlistTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":                  item["_id"],
		"name":                 item["name"],
		"createdAt":            item["createdAt"],
		"createdBy":            item["createdBy"],
		"_updatedAt":           item["_updatedAt"],
		"userIds":              item["userIds"],
		"beta":                 item["beta"],
		"billing":              item["billing"],
		"revenueVisualization": item["revenueVisualization"],
	}
}

func lemlistCampaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":    item["_id"],
		"name":   item["name"],
		"labels": item["labels"],
	}
}

func lemlistActivityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":               item["_id"],
		"type":              item["type"],
		"campaignId":        item["campaignId"],
		"campaignName":      item["campaignName"],
		"companyName":       item["companyName"],
		"createdAt":         item["createdAt"],
		"createdBy":         item["createdBy"],
		"email":             item["email"],
		"emailTemplateId":   item["emailTemplateId"],
		"emailTemplateName": item["emailTemplateName"],
		"extra":             item["extra"],
		"firstName":         item["firstName"],
		"lastName":          item["lastName"],
		"icebreaker":        item["icebreaker"],
		"isFirst":           item["isFirst"],
		"leadId":            item["leadId"],
		"leadEmail":         item["leadEmail"],
		"leadFirstName":     item["leadFirstName"],
		"leadLastName":      item["leadLastName"],
		"linkedinUrl":       item["linkedinUrl"],
		"phone":             item["phone"],
		"sequenceId":        item["sequenceId"],
		"sequenceStep":      item["sequenceStep"],
		"teamId":            item["teamId"],
	}
}

func lemlistUnsubscribeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":   item["_id"],
		"email": item["email"],
	}
}
