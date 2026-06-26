package jotform

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Jotform API resource path (relative
// to base_url) it reads from and the record mapper that flattens its objects.
// paginated marks streams whose endpoints honor resultSet offset/limit
// pagination (forms, submissions); the rest return a bounded single response.
type streamEndpoint struct {
	resource  string
	paginated bool
	mapRecord func(map[string]any) connectors.Record
}

// jotformStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in jotformStreams; the read
// path is fully data-driven from this table.
var jotformStreamEndpoints = map[string]streamEndpoint{
	"forms":       {resource: "user/forms", paginated: true, mapRecord: jotformFormRecord},
	"submissions": {resource: "user/submissions", paginated: true, mapRecord: jotformSubmissionRecord},
	"reports":     {resource: "user/reports", paginated: false, mapRecord: jotformReportRecord},
	"folders":     {resource: "user/folders", paginated: false, mapRecord: jotformFolderRecord},
	"user":        {resource: "user", paginated: false, mapRecord: jotformUserRecord},
}

// jotformStreams returns the connector's published stream catalog. Jotform list
// objects expose a string id and a "created_at" timestamp, so the primary key is
// ["id"] and the incremental cursor field is ["created_at"] where supported.
func jotformStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "forms",
			Description:  "Jotform forms owned by the authenticated user.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       jotformFormFields(),
		},
		{
			Name:         "submissions",
			Description:  "Submissions across the authenticated user's forms.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       jotformSubmissionFields(),
		},
		{
			Name:         "reports",
			Description:  "Reports generated from the user's forms.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       jotformReportFields(),
		},
		{
			Name:        "folders",
			Description: "Folders organizing the user's forms.",
			PrimaryKey:  []string{"id"},
			Fields:      jotformFolderFields(),
		},
		{
			Name:        "user",
			Description: "The authenticated Jotform account profile.",
			PrimaryKey:  []string{"username"},
			Fields:      jotformUserFields(),
		},
	}
}

func jotformFormFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "username", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "last_submission", Type: "string"},
		{Name: "new", Type: "string"},
		{Name: "count", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "url", Type: "string"},
	}
}

func jotformSubmissionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "form_id", Type: "string"},
		{Name: "ip", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "new", Type: "string"},
		{Name: "flag", Type: "string"},
		{Name: "notes", Type: "string"},
		{Name: "answers", Type: "object"},
	}
}

func jotformReportFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "form_id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "fields", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "url", Type: "string"},
	}
}

func jotformFolderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "parent", Type: "string"},
		{Name: "owner", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "forms", Type: "object"},
		{Name: "subfolders", Type: "object"},
	}
}

func jotformUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "username", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "account_type", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "usage", Type: "string"},
		{Name: "time_zone", Type: "string"},
	}
}

func jotformFormRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"username":        item["username"],
		"title":           item["title"],
		"status":          item["status"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
		"last_submission": item["last_submission"],
		"new":             item["new"],
		"count":           item["count"],
		"type":            item["type"],
		"url":             item["url"],
	}
}

func jotformSubmissionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"form_id":    item["form_id"],
		"ip":         item["ip"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
		"status":     item["status"],
		"new":        item["new"],
		"flag":       item["flag"],
		"notes":      item["notes"],
		"answers":    item["answers"],
	}
}

func jotformReportRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"form_id":    item["form_id"],
		"title":      item["title"],
		"type":       item["type"],
		"status":     item["status"],
		"fields":     item["fields"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
		"url":        item["url"],
	}
}

func jotformFolderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"parent":     item["parent"],
		"owner":      item["owner"],
		"color":      item["color"],
		"forms":      item["forms"],
		"subfolders": item["subfolders"],
	}
}

func jotformUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"username":     item["username"],
		"name":         item["name"],
		"email":        item["email"],
		"status":       item["status"],
		"account_type": item["account_type"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
		"usage":        item["usage"],
		"time_zone":    item["time_zone"],
	}
}
