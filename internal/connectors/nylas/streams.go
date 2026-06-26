package nylas

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Nylas grant-scoped resource path
// segment (e.g. "calendars") and the record mapper that flattens its objects.
// All Nylas v3 list endpoints live under /v3/grants/{grant_id}/<resource>.
type streamEndpoint struct {
	// resource is the path segment under the grant (e.g. "calendars").
	resource string
	// requiresCalendarID is true for the events endpoint, which mandates a
	// calendar_id query parameter.
	requiresCalendarID bool
	// mapRecord flattens a raw Nylas object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// nylasStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in nylasStreams; the read path
// is fully data-driven from this table.
var nylasStreamEndpoints = map[string]streamEndpoint{
	"calendars": {resource: "calendars", mapRecord: nylasCalendarRecord},
	"contacts":  {resource: "contacts", mapRecord: nylasContactRecord},
	"messages":  {resource: "messages", mapRecord: nylasMessageRecord},
	"events":    {resource: "events", requiresCalendarID: true, mapRecord: nylasEventRecord},
}

// nylasStreams returns the connector's published stream catalog. Every Nylas v3
// object exposes a string id, so the primary key is ["id"] across the board.
// Messages and events carry a timestamp cursor (date / updated_at).
func nylasStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "calendars",
			Description: "Nylas calendars for the connected grant.",
			PrimaryKey:  []string{"id"},
			Fields:      nylasCalendarFields(),
		},
		{
			Name:        "contacts",
			Description: "Nylas contacts for the connected grant.",
			PrimaryKey:  []string{"id"},
			Fields:      nylasContactFields(),
		},
		{
			Name:         "messages",
			Description:  "Nylas email messages for the connected grant.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date"},
			Fields:       nylasMessageFields(),
		},
		{
			Name:         "events",
			Description:  "Nylas calendar events for the connected grant (requires calendar_id).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       nylasEventFields(),
		},
	}
}

func nylasCalendarFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "grant_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "is_primary", Type: "boolean"},
		{Name: "read_only", Type: "boolean"},
		{Name: "hex_color", Type: "string"},
	}
}

func nylasContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "grant_id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "given_name", Type: "string"},
		{Name: "surname", Type: "string"},
		{Name: "company_name", Type: "string"},
		{Name: "job_title", Type: "string"},
		{Name: "emails", Type: "array"},
		{Name: "phone_numbers", Type: "array"},
		{Name: "source", Type: "string"},
	}
}

func nylasMessageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "grant_id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "thread_id", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "snippet", Type: "string"},
		{Name: "from", Type: "array"},
		{Name: "to", Type: "array"},
		{Name: "date", Type: "integer"},
		{Name: "unread", Type: "boolean"},
		{Name: "starred", Type: "boolean"},
		{Name: "folders", Type: "array"},
	}
}

func nylasEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "grant_id", Type: "string"},
		{Name: "calendar_id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "location", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "busy", Type: "boolean"},
		{Name: "read_only", Type: "boolean"},
		{Name: "when", Type: "object"},
		{Name: "updated_at", Type: "integer"},
	}
}

func nylasCalendarRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"grant_id":    item["grant_id"],
		"name":        item["name"],
		"description": item["description"],
		"timezone":    item["timezone"],
		"object":      item["object"],
		"is_primary":  item["is_primary"],
		"read_only":   item["read_only"],
		"hex_color":   item["hex_color"],
	}
}

func nylasContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"grant_id":      item["grant_id"],
		"object":        item["object"],
		"given_name":    item["given_name"],
		"surname":       item["surname"],
		"company_name":  item["company_name"],
		"job_title":     item["job_title"],
		"emails":        item["emails"],
		"phone_numbers": item["phone_numbers"],
		"source":        item["source"],
	}
}

func nylasMessageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"grant_id":  item["grant_id"],
		"object":    item["object"],
		"thread_id": item["thread_id"],
		"subject":   item["subject"],
		"snippet":   item["snippet"],
		"from":      item["from"],
		"to":        item["to"],
		"date":      item["date"],
		"unread":    item["unread"],
		"starred":   item["starred"],
		"folders":   item["folders"],
	}
}

func nylasEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"grant_id":    item["grant_id"],
		"calendar_id": item["calendar_id"],
		"object":      item["object"],
		"title":       item["title"],
		"description": item["description"],
		"location":    item["location"],
		"status":      item["status"],
		"busy":        item["busy"],
		"read_only":   item["read_only"],
		"when":        item["when"],
		"updated_at":  item["updated_at"],
	}
}
