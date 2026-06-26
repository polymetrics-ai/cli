package mailerlite

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the MailerLite API resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects. Every v2 list response wraps records in a top-level "data" array and
// carries meta.next_cursor for pagination, so the read path is fully data-driven
// from this table.
type streamEndpoint struct {
	// resource is the MailerLite list endpoint path segment (e.g. "subscribers").
	resource string
	// mapRecord flattens a raw MailerLite object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// mailerliteStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in mailerliteStreams.
var mailerliteStreamEndpoints = map[string]streamEndpoint{
	"subscribers": {resource: "subscribers", mapRecord: subscriberRecord},
	"campaigns":   {resource: "campaigns", mapRecord: campaignRecord},
	"groups":      {resource: "groups", mapRecord: groupRecord},
	"segments":    {resource: "segments", mapRecord: segmentRecord},
	"automations": {resource: "automations", mapRecord: automationRecord},
}

// mailerliteStreams returns the connector's published stream catalog. MailerLite
// objects expose a string id and (mostly) an updated_at/created_at timestamp; the
// primary key is ["id"] and the incremental cursor field is the object's
// timestamp where one exists.
func mailerliteStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "subscribers",
			Description:  "MailerLite subscribers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       subscriberFields(),
		},
		{
			Name:         "campaigns",
			Description:  "MailerLite email campaigns.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       campaignFields(),
		},
		{
			Name:         "groups",
			Description:  "MailerLite subscriber groups.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       groupFields(),
		},
		{
			Name:         "segments",
			Description:  "MailerLite subscriber segments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       segmentFields(),
		},
		{
			Name:         "automations",
			Description:  "MailerLite automation workflows.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       automationFields(),
		},
	}
}

func subscriberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "sent", Type: "integer"},
		{Name: "opens_count", Type: "integer"},
		{Name: "clicks_count", Type: "integer"},
		{Name: "open_rate", Type: "number"},
		{Name: "click_rate", Type: "number"},
		{Name: "ip_address", Type: "string"},
		{Name: "subscribed_at", Type: "timestamp"},
		{Name: "unsubscribed_at", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "fields", Type: "object"},
	}
}

func campaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "account_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "scheduled_for", Type: "timestamp"},
		{Name: "started_at", Type: "timestamp"},
		{Name: "finished_at", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "is_stopped", Type: "boolean"},
		{Name: "stats", Type: "object"},
	}
}

func groupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "active_count", Type: "integer"},
		{Name: "sent_count", Type: "integer"},
		{Name: "opens_count", Type: "integer"},
		{Name: "clicks_count", Type: "integer"},
		{Name: "unsubscribed_count", Type: "integer"},
		{Name: "open_rate", Type: "object"},
		{Name: "click_rate", Type: "object"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func segmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "total", Type: "integer"},
		{Name: "open_rate", Type: "object"},
		{Name: "click_rate", Type: "object"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func automationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "enabled", Type: "boolean"},
		{Name: "status", Type: "string"},
		{Name: "trigger_data", Type: "object"},
		{Name: "steps", Type: "object"},
		{Name: "stats", Type: "object"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func subscriberRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"email":           item["email"],
		"status":          item["status"],
		"source":          item["source"],
		"sent":            item["sent"],
		"opens_count":     item["opens_count"],
		"clicks_count":    item["clicks_count"],
		"open_rate":       item["open_rate"],
		"click_rate":      item["click_rate"],
		"ip_address":      item["ip_address"],
		"subscribed_at":   item["subscribed_at"],
		"unsubscribed_at": item["unsubscribed_at"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
		"fields":          item["fields"],
	}
}

func campaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"account_id":    item["account_id"],
		"name":          item["name"],
		"type":          item["type"],
		"status":        item["status"],
		"scheduled_for": item["scheduled_for"],
		"started_at":    item["started_at"],
		"finished_at":   item["finished_at"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
		"is_stopped":    item["is_stopped"],
		"stats":         item["stats"],
	}
}

func groupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"name":               item["name"],
		"active_count":       item["active_count"],
		"sent_count":         item["sent_count"],
		"opens_count":        item["opens_count"],
		"clicks_count":       item["clicks_count"],
		"unsubscribed_count": item["unsubscribed_count"],
		"open_rate":          item["open_rate"],
		"click_rate":         item["click_rate"],
		"created_at":         item["created_at"],
	}
}

func segmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"total":      item["total"],
		"open_rate":  item["open_rate"],
		"click_rate": item["click_rate"],
		"created_at": item["created_at"],
	}
}

func automationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"enabled":      item["enabled"],
		"status":       item["status"],
		"trigger_data": item["trigger_data"],
		"steps":        item["steps"],
		"stats":        item["stats"],
		"created_at":   item["created_at"],
	}
}
