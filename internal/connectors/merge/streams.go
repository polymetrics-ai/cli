package merge

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Merge ATS API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Merge list endpoint path segment (e.g. "candidates").
	resource string
	// mapRecord flattens a raw Merge object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// mergeStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in mergeStreams; the read path
// is fully data-driven from this table. These are core Merge ATS (Applicant
// Tracking System) Common Model objects.
var mergeStreamEndpoints = map[string]streamEndpoint{
	"candidates":   {resource: "candidates", mapRecord: mergeCandidateRecord},
	"applications": {resource: "applications", mapRecord: mergeApplicationRecord},
	"jobs":         {resource: "jobs", mapRecord: mergeJobRecord},
	"offers":       {resource: "offers", mapRecord: mergeOfferRecord},
	"departments":  {resource: "departments", mapRecord: mergeDepartmentRecord},
	"users":        {resource: "users", mapRecord: mergeUserRecord},
}

// mergeStreams returns the connector's published stream catalog. Every Merge
// Common Model object exposes a string id and a `modified_at` ISO-8601
// timestamp, so the primary key is ["id"] and the incremental cursor field is
// ["modified_at"] across the board.
func mergeStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "candidates",
			Description:  "Merge ATS candidates.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified_at"},
			Fields:       mergeCandidateFields(),
		},
		{
			Name:         "applications",
			Description:  "Merge ATS applications.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified_at"},
			Fields:       mergeApplicationFields(),
		},
		{
			Name:         "jobs",
			Description:  "Merge ATS jobs.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified_at"},
			Fields:       mergeJobFields(),
		},
		{
			Name:         "offers",
			Description:  "Merge ATS offers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified_at"},
			Fields:       mergeOfferFields(),
		},
		{
			Name:         "departments",
			Description:  "Merge ATS departments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified_at"},
			Fields:       mergeDepartmentFields(),
		},
		{
			Name:         "users",
			Description:  "Merge ATS users.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified_at"},
			Fields:       mergeUserFields(),
		},
	}
}

func mergeCandidateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "remote_id", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "is_private", Type: "boolean"},
		{Name: "can_email", Type: "boolean"},
		{Name: "remote_created_at", Type: "string"},
		{Name: "remote_updated_at", Type: "string"},
		{Name: "last_interaction_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
		{Name: "remote_was_deleted", Type: "boolean"},
	}
}

func mergeApplicationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "remote_id", Type: "string"},
		{Name: "candidate", Type: "string"},
		{Name: "job", Type: "string"},
		{Name: "applied_at", Type: "string"},
		{Name: "rejected_at", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "credited_to", Type: "string"},
		{Name: "current_stage", Type: "string"},
		{Name: "reject_reason", Type: "string"},
		{Name: "modified_at", Type: "string"},
		{Name: "remote_was_deleted", Type: "boolean"},
	}
}

func mergeJobFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "remote_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "confidential", Type: "boolean"},
		{Name: "remote_created_at", Type: "string"},
		{Name: "remote_updated_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
		{Name: "remote_was_deleted", Type: "boolean"},
	}
}

func mergeOfferFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "remote_id", Type: "string"},
		{Name: "application", Type: "string"},
		{Name: "creator", Type: "string"},
		{Name: "remote_created_at", Type: "string"},
		{Name: "closed_at", Type: "string"},
		{Name: "sent_at", Type: "string"},
		{Name: "start_date", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "modified_at", Type: "string"},
		{Name: "remote_was_deleted", Type: "boolean"},
	}
}

func mergeDepartmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "remote_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "modified_at", Type: "string"},
		{Name: "remote_was_deleted", Type: "boolean"},
	}
}

func mergeUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "remote_id", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "disabled", Type: "boolean"},
		{Name: "access_role", Type: "string"},
		{Name: "remote_created_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
		{Name: "remote_was_deleted", Type: "boolean"},
	}
}

func mergeCandidateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"remote_id":           item["remote_id"],
		"first_name":          item["first_name"],
		"last_name":           item["last_name"],
		"company":             item["company"],
		"title":               item["title"],
		"is_private":          item["is_private"],
		"can_email":           item["can_email"],
		"remote_created_at":   item["remote_created_at"],
		"remote_updated_at":   item["remote_updated_at"],
		"last_interaction_at": item["last_interaction_at"],
		"modified_at":         item["modified_at"],
		"remote_was_deleted":  item["remote_was_deleted"],
	}
}

func mergeApplicationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"remote_id":          item["remote_id"],
		"candidate":          item["candidate"],
		"job":                item["job"],
		"applied_at":         item["applied_at"],
		"rejected_at":        item["rejected_at"],
		"source":             item["source"],
		"credited_to":        item["credited_to"],
		"current_stage":      item["current_stage"],
		"reject_reason":      item["reject_reason"],
		"modified_at":        item["modified_at"],
		"remote_was_deleted": item["remote_was_deleted"],
	}
}

func mergeJobRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"remote_id":          item["remote_id"],
		"name":               item["name"],
		"description":        item["description"],
		"code":               item["code"],
		"status":             item["status"],
		"type":               item["type"],
		"confidential":       item["confidential"],
		"remote_created_at":  item["remote_created_at"],
		"remote_updated_at":  item["remote_updated_at"],
		"modified_at":        item["modified_at"],
		"remote_was_deleted": item["remote_was_deleted"],
	}
}

func mergeOfferRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"remote_id":          item["remote_id"],
		"application":        item["application"],
		"creator":            item["creator"],
		"remote_created_at":  item["remote_created_at"],
		"closed_at":          item["closed_at"],
		"sent_at":            item["sent_at"],
		"start_date":         item["start_date"],
		"status":             item["status"],
		"modified_at":        item["modified_at"],
		"remote_was_deleted": item["remote_was_deleted"],
	}
}

func mergeDepartmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"remote_id":          item["remote_id"],
		"name":               item["name"],
		"modified_at":        item["modified_at"],
		"remote_was_deleted": item["remote_was_deleted"],
	}
}

func mergeUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"remote_id":          item["remote_id"],
		"first_name":         item["first_name"],
		"last_name":          item["last_name"],
		"email":              item["email"],
		"disabled":           item["disabled"],
		"access_role":        item["access_role"],
		"remote_created_at":  item["remote_created_at"],
		"modified_at":        item["modified_at"],
		"remote_was_deleted": item["remote_was_deleted"],
	}
}
