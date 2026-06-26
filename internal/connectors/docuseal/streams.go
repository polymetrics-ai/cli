package docuseal

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the DocuSeal API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the DocuSeal list endpoint path segment (e.g. "templates").
	resource string
	// mapRecord flattens a raw DocuSeal object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// docusealStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in docusealStreams; the read
// path is fully data-driven from this table.
var docusealStreamEndpoints = map[string]streamEndpoint{
	"templates":   {resource: "templates", mapRecord: docusealTemplateRecord},
	"submissions": {resource: "submissions", mapRecord: docusealSubmissionRecord},
	"submitters":  {resource: "submitters", mapRecord: docusealSubmitterRecord},
}

// docusealStreams returns the connector's published stream catalog. Every
// DocuSeal object exposes an integer id and an updated_at/created_at timestamp,
// so the primary key is ["id"] and the incremental cursor field is
// ["updated_at"] (falling back to created_at where the resource lacks updates).
func docusealStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "templates",
			Description:  "DocuSeal document templates.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       docusealTemplateFields(),
		},
		{
			Name:         "submissions",
			Description:  "DocuSeal submissions (signing requests).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       docusealSubmissionFields(),
		},
		{
			Name:         "submitters",
			Description:  "DocuSeal submitters (signers on a submission).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       docusealSubmitterFields(),
		},
	}
}

func docusealTemplateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "slug", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "folder_name", Type: "string"},
		{Name: "external_id", Type: "string"},
		{Name: "author_id", Type: "integer"},
		{Name: "archived_at", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func docusealSubmissionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "slug", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "template_id", Type: "integer"},
		{Name: "template_name", Type: "string"},
		{Name: "audit_log_url", Type: "string"},
		{Name: "combined_document_url", Type: "string"},
		{Name: "expire_at", Type: "string"},
		{Name: "completed_at", Type: "string"},
		{Name: "archived_at", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func docusealSubmitterFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "submission_id", Type: "integer"},
		{Name: "uuid", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "external_id", Type: "string"},
		{Name: "sent_at", Type: "string"},
		{Name: "opened_at", Type: "string"},
		{Name: "completed_at", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func docusealTemplateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"slug":        item["slug"],
		"name":        item["name"],
		"folder_name": item["folder_name"],
		"external_id": item["external_id"],
		"author_id":   item["author_id"],
		"archived_at": item["archived_at"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func docusealSubmissionRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":                    item["id"],
		"slug":                  item["slug"],
		"name":                  item["name"],
		"source":                item["source"],
		"status":                item["status"],
		"audit_log_url":         item["audit_log_url"],
		"combined_document_url": item["combined_document_url"],
		"expire_at":             item["expire_at"],
		"completed_at":          item["completed_at"],
		"archived_at":           item["archived_at"],
		"created_at":            item["created_at"],
		"updated_at":            item["updated_at"],
	}
	// Flatten the nested template object to its id and name.
	if tmpl, ok := item["template"].(map[string]any); ok {
		rec["template_id"] = tmpl["id"]
		rec["template_name"] = tmpl["name"]
	}
	return rec
}

func docusealSubmitterRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"submission_id": item["submission_id"],
		"uuid":          item["uuid"],
		"slug":          item["slug"],
		"email":         item["email"],
		"name":          item["name"],
		"phone":         item["phone"],
		"status":        item["status"],
		"role":          item["role"],
		"external_id":   item["external_id"],
		"sent_at":       item["sent_at"],
		"opened_at":     item["opened_at"],
		"completed_at":  item["completed_at"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}
