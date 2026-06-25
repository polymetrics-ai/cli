package greenhouse

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Greenhouse Harvest API resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects.
type streamEndpoint struct {
	// resource is the Harvest list endpoint path segment (e.g. "candidates").
	resource string
	// mapRecord flattens a raw Greenhouse object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// greenhouseStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in greenhouseStreams; the
// read path is fully data-driven from this table.
var greenhouseStreamEndpoints = map[string]streamEndpoint{
	"candidates":   {resource: "candidates", mapRecord: candidateRecord},
	"applications": {resource: "applications", mapRecord: applicationRecord},
	"jobs":         {resource: "jobs", mapRecord: jobRecord},
	"offers":       {resource: "offers", mapRecord: offerRecord},
	"users":        {resource: "users", mapRecord: userRecord},
}

// greenhouseStreams returns the connector's published stream catalog. Greenhouse
// Harvest objects expose an integer id and (mostly) an updated_at timestamp, so
// the primary key is ["id"] and the incremental cursor field is ["updated_at"]
// where available.
func greenhouseStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "candidates",
			Description:  "Greenhouse candidates.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       candidateFields(),
		},
		{
			Name:         "applications",
			Description:  "Greenhouse applications.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_activity_at"},
			Fields:       applicationFields(),
		},
		{
			Name:         "jobs",
			Description:  "Greenhouse jobs.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       jobFields(),
		},
		{
			Name:         "offers",
			Description:  "Greenhouse offers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       offerFields(),
		},
		{
			Name:         "users",
			Description:  "Greenhouse users.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       userFields(),
		},
	}
}

func candidateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "is_private", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "last_activity", Type: "timestamp"},
	}
}

func applicationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "candidate_id", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "applied_at", Type: "timestamp"},
		{Name: "rejected_at", Type: "timestamp"},
		{Name: "last_activity_at", Type: "timestamp"},
		{Name: "source_id", Type: "integer"},
	}
}

func jobFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "confidential", Type: "boolean"},
		{Name: "requisition_id", Type: "string"},
		{Name: "opened_at", Type: "timestamp"},
		{Name: "closed_at", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func offerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "application_id", Type: "integer"},
		{Name: "candidate_id", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "version", Type: "integer"},
		{Name: "starts_at", Type: "timestamp"},
		{Name: "sent_at", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "primary_email_address", Type: "string"},
		{Name: "disabled", Type: "boolean"},
		{Name: "site_admin", Type: "boolean"},
		{Name: "employee_id", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func candidateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"first_name":    item["first_name"],
		"last_name":     item["last_name"],
		"company":       item["company"],
		"title":         item["title"],
		"is_private":    item["is_private"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
		"last_activity": item["last_activity"],
	}
}

func applicationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"candidate_id":     item["candidate_id"],
		"status":           item["status"],
		"applied_at":       item["applied_at"],
		"rejected_at":      item["rejected_at"],
		"last_activity_at": item["last_activity_at"],
		"source_id":        item["source_id"],
	}
}

func jobRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"status":         item["status"],
		"confidential":   item["confidential"],
		"requisition_id": item["requisition_id"],
		"opened_at":      item["opened_at"],
		"closed_at":      item["closed_at"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
	}
}

func offerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"application_id": item["application_id"],
		"candidate_id":   item["candidate_id"],
		"status":         item["status"],
		"version":        item["version"],
		"starts_at":      item["starts_at"],
		"sent_at":        item["sent_at"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                    item["id"],
		"name":                  item["name"],
		"first_name":            item["first_name"],
		"last_name":             item["last_name"],
		"primary_email_address": item["primary_email_address"],
		"disabled":              item["disabled"],
		"site_admin":            item["site_admin"],
		"employee_id":           item["employee_id"],
		"created_at":            item["created_at"],
		"updated_at":            item["updated_at"],
	}
}
