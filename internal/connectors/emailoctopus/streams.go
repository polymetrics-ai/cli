package emailoctopus

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the EmailOctopus API resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects. parentResource, when set, marks a stream whose resource path is
// templated per parent record (e.g. lists/{id}/contacts).
type streamEndpoint struct {
	// resource is the API list endpoint path segment (e.g. "lists").
	resource string
	// mapRecord flattens a raw API object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// emailOctopusStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in emailOctopusStreams;
// the read path is data-driven from this table.
var emailOctopusStreamEndpoints = map[string]streamEndpoint{
	"lists":         {resource: "lists", mapRecord: listRecord},
	"campaigns":     {resource: "campaigns", mapRecord: campaignRecord},
	"list_contacts": {resource: "lists/{list_id}/contacts", mapRecord: contactRecord},
}

// emailOctopusStreams returns the connector's published stream catalog. The API
// supports full_refresh only (no incremental cursor), so CursorFields are empty;
// created_at is surfaced as a field for downstream use.
func emailOctopusStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "lists",
			Description: "EmailOctopus lists with subscriber counts.",
			PrimaryKey:  []string{"id"},
			Fields:      listFields(),
		},
		{
			Name:        "campaigns",
			Description: "EmailOctopus campaigns.",
			PrimaryKey:  []string{"id"},
			Fields:      campaignFields(),
		},
		{
			Name:        "list_contacts",
			Description: "Contacts belonging to a list (requires config list_id).",
			PrimaryKey:  []string{"id"},
			Fields:      contactFields(),
		},
	}
}

func listFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "double_opt_in", Type: "boolean"},
		{Name: "created_at", Type: "string"},
		{Name: "pending_count", Type: "integer"},
		{Name: "subscribed_count", Type: "integer"},
		{Name: "unsubscribed_count", Type: "integer"},
	}
}

func campaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "from_name", Type: "string"},
		{Name: "from_email_address", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "sent_at", Type: "string"},
	}
}

func contactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email_address", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "tags", Type: "array"},
		{Name: "fields", Type: "object"},
		{Name: "created_at", Type: "string"},
		{Name: "last_updated_at", Type: "string"},
	}
}

func listRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"double_opt_in": item["double_opt_in"],
		"created_at":    item["created_at"],
	}
	if counts, ok := item["counts"].(map[string]any); ok {
		rec["pending_count"] = counts["pending"]
		rec["subscribed_count"] = counts["subscribed"]
		rec["unsubscribed_count"] = counts["unsubscribed"]
	}
	return rec
}

func campaignRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":         item["id"],
		"status":     item["status"],
		"name":       item["name"],
		"subject":    item["subject"],
		"created_at": item["created_at"],
		"sent_at":    item["sent_at"],
	}
	if from, ok := item["from"].(map[string]any); ok {
		rec["from_name"] = from["name"]
		rec["from_email_address"] = from["email_address"]
	}
	return rec
}

func contactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"email_address":   item["email_address"],
		"status":          item["status"],
		"tags":            item["tags"],
		"fields":          item["fields"],
		"created_at":      item["created_at"],
		"last_updated_at": item["last_updated_at"],
	}
}
