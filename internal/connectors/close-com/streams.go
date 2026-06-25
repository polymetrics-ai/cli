package closecom

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Close API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Close list endpoint path segment (e.g. "lead").
	resource string
	// mapRecord flattens a raw Close object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// closeStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in closeStreams; the read path
// is fully data-driven from this table. Close list paths are singular and
// trailing-slashed (e.g. /lead/, /contact/).
var closeStreamEndpoints = map[string]streamEndpoint{
	"leads":         {resource: "lead/", mapRecord: closeLeadRecord},
	"contacts":      {resource: "contact/", mapRecord: closeContactRecord},
	"opportunities": {resource: "opportunity/", mapRecord: closeOpportunityRecord},
	"activities":    {resource: "activity/", mapRecord: closeActivityRecord},
	"users":         {resource: "user/", mapRecord: closeUserRecord},
}

// closeStreams returns the connector's published stream catalog. Every Close
// object exposes a string id and date_created/date_updated timestamps, so the
// primary key is ["id"] and the incremental cursor field is ["date_updated"].
func closeStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "leads",
			Description:  "Close leads (companies/accounts tracked in the CRM).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_updated"},
			Fields:       closeLeadFields(),
		},
		{
			Name:         "contacts",
			Description:  "Close contacts (people) associated with leads.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_updated"},
			Fields:       closeContactFields(),
		},
		{
			Name:         "opportunities",
			Description:  "Close opportunities (deals) on leads.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_updated"},
			Fields:       closeOpportunityFields(),
		},
		{
			Name:         "activities",
			Description:  "Close activities (calls, emails, notes, meetings, SMS).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_updated"},
			Fields:       closeActivityFields(),
		},
		{
			Name:         "users",
			Description:  "Close users (members of the organization).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_updated"},
			Fields:       closeUserFields(),
		},
	}
}

func closeLeadFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "status_id", Type: "string"},
		{Name: "status_label", Type: "string"},
		{Name: "organization_id", Type: "string"},
		{Name: "created_by", Type: "string"},
		{Name: "date_created", Type: "string"},
		{Name: "date_updated", Type: "string"},
	}
}

func closeContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "lead_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "organization_id", Type: "string"},
		{Name: "created_by", Type: "string"},
		{Name: "date_created", Type: "string"},
		{Name: "date_updated", Type: "string"},
	}
}

func closeOpportunityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "lead_id", Type: "string"},
		{Name: "lead_name", Type: "string"},
		{Name: "status_id", Type: "string"},
		{Name: "status_label", Type: "string"},
		{Name: "status_type", Type: "string"},
		{Name: "pipeline_id", Type: "string"},
		{Name: "value", Type: "integer"},
		{Name: "value_currency", Type: "string"},
		{Name: "value_formatted", Type: "string"},
		{Name: "confidence", Type: "integer"},
		{Name: "organization_id", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "date_won", Type: "string"},
		{Name: "date_created", Type: "string"},
		{Name: "date_updated", Type: "string"},
	}
}

func closeActivityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "_type", Type: "string"},
		{Name: "lead_id", Type: "string"},
		{Name: "contact_id", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "user_name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "direction", Type: "string"},
		{Name: "organization_id", Type: "string"},
		{Name: "created_by", Type: "string"},
		{Name: "date_created", Type: "string"},
		{Name: "date_updated", Type: "string"},
	}
}

func closeUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "image", Type: "string"},
		{Name: "date_created", Type: "string"},
		{Name: "date_updated", Type: "string"},
	}
}

func closeLeadRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"display_name":    item["display_name"],
		"name":            item["name"],
		"description":     item["description"],
		"url":             item["url"],
		"status_id":       item["status_id"],
		"status_label":    item["status_label"],
		"organization_id": item["organization_id"],
		"created_by":      item["created_by"],
		"date_created":    item["date_created"],
		"date_updated":    item["date_updated"],
	}
}

func closeContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"lead_id":         item["lead_id"],
		"name":            item["name"],
		"title":           item["title"],
		"organization_id": item["organization_id"],
		"created_by":      item["created_by"],
		"date_created":    item["date_created"],
		"date_updated":    item["date_updated"],
	}
}

func closeOpportunityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"lead_id":         item["lead_id"],
		"lead_name":       item["lead_name"],
		"status_id":       item["status_id"],
		"status_label":    item["status_label"],
		"status_type":     item["status_type"],
		"pipeline_id":     item["pipeline_id"],
		"value":           item["value"],
		"value_currency":  item["value_currency"],
		"value_formatted": item["value_formatted"],
		"confidence":      item["confidence"],
		"organization_id": item["organization_id"],
		"user_id":         item["user_id"],
		"date_won":        item["date_won"],
		"date_created":    item["date_created"],
		"date_updated":    item["date_updated"],
	}
}

func closeActivityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"_type":           item["_type"],
		"lead_id":         item["lead_id"],
		"contact_id":      item["contact_id"],
		"user_id":         item["user_id"],
		"user_name":       item["user_name"],
		"status":          item["status"],
		"direction":       item["direction"],
		"organization_id": item["organization_id"],
		"created_by":      item["created_by"],
		"date_created":    item["date_created"],
		"date_updated":    item["date_updated"],
	}
}

func closeUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"first_name":   item["first_name"],
		"last_name":    item["last_name"],
		"email":        item["email"],
		"image":        item["image"],
		"date_created": item["date_created"],
		"date_updated": item["date_updated"],
	}
}
