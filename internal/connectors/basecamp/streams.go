package basecamp

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Basecamp API resource path (relative
// to base_url, which already includes the account id) it reads from, and the
// record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Basecamp list endpoint path segment (e.g. "projects.json").
	resource string
	// mapRecord flattens a raw Basecamp object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// basecampStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in basecampStreams; the read
// path is fully data-driven from this table. All chosen endpoints are top-level
// account collections that paginate via the RFC5988 Link header, so they need no
// project context to enumerate.
var basecampStreamEndpoints = map[string]streamEndpoint{
	"projects": {resource: "projects.json", mapRecord: basecampProjectRecord},
	"people":   {resource: "people.json", mapRecord: basecampPersonRecord},
	"events":   {resource: "events.json", mapRecord: basecampEventRecord},
}

// basecampStreams returns the connector's published stream catalog. Every
// Basecamp resource exposes an integer id and an updated_at timestamp, so the
// primary key is ["id"] and the incremental cursor field is ["updated_at"].
func basecampStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "projects",
			Description:  "Basecamp projects (active and archived) on the account.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       basecampProjectFields(),
		},
		{
			Name:         "people",
			Description:  "People (members) visible on the Basecamp account.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       basecampPersonFields(),
		},
		{
			Name:         "events",
			Description:  "Account-wide activity events (recording changes) on Basecamp.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       basecampEventFields(),
		},
	}
}

func basecampProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "purpose", Type: "string"},
		{Name: "bookmark_url", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "app_url", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func basecampPersonFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email_address", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "admin", Type: "boolean"},
		{Name: "owner", Type: "boolean"},
		{Name: "client", Type: "boolean"},
		{Name: "personable_type", Type: "string"},
		{Name: "time_zone", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func basecampEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "recording_id", Type: "integer"},
		{Name: "action", Type: "string"},
		{Name: "kind", Type: "string"},
		{Name: "summary", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func basecampProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"status":       item["status"],
		"name":         item["name"],
		"description":  item["description"],
		"purpose":      item["purpose"],
		"bookmark_url": item["bookmark_url"],
		"url":          item["url"],
		"app_url":      item["app_url"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func basecampPersonRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            item["name"],
		"email_address":   item["email_address"],
		"title":           item["title"],
		"admin":           item["admin"],
		"owner":           item["owner"],
		"client":          item["client"],
		"personable_type": item["personable_type"],
		"time_zone":       item["time_zone"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
	}
}

func basecampEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"recording_id": item["recording_id"],
		"action":       item["action"],
		"kind":         item["kind"],
		"summary":      item["summary"],
		"created_at":   item["created_at"],
	}
}
