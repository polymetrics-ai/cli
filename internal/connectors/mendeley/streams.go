package mendeley

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Mendeley API resource path (relative
// to base_url), the resource-specific Accept media type the endpoint requires,
// and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Mendeley list endpoint path segment (e.g. "documents").
	resource string
	// accept is the versioned vendor media type required by the endpoint
	// (Mendeley uses application/vnd.mendeley-<resource>.1+json).
	accept string
	// mapRecord flattens a raw Mendeley object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// mendeleyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in mendeleyStreams; the read
// path is fully data-driven from this table.
var mendeleyStreamEndpoints = map[string]streamEndpoint{
	"documents":   {resource: "documents", accept: "application/vnd.mendeley-document.1+json", mapRecord: mendeleyDocumentRecord},
	"folders":     {resource: "folders", accept: "application/vnd.mendeley-folder.1+json", mapRecord: mendeleyFolderRecord},
	"groups":      {resource: "groups", accept: "application/vnd.mendeley-group.1+json", mapRecord: mendeleyGroupRecord},
	"annotations": {resource: "annotations", accept: "application/vnd.mendeley-annotation.1+json", mapRecord: mendeleyAnnotationRecord},
}

// mendeleyStreams returns the connector's published stream catalog. Every
// Mendeley object exposes a string id; documents/folders/annotations also carry
// a last_modified timestamp used as the incremental cursor.
func mendeleyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "documents",
			Description:  "Documents in the authenticated user's Mendeley library.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_modified"},
			Fields:       mendeleyDocumentFields(),
		},
		{
			Name:         "folders",
			Description:  "Folders in the authenticated user's Mendeley library.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       mendeleyFolderFields(),
		},
		{
			Name:         "groups",
			Description:  "Groups the authenticated user belongs to.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       mendeleyGroupFields(),
		},
		{
			Name:         "annotations",
			Description:  "Annotations attached to the authenticated user's documents.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_modified"},
			Fields:       mendeleyAnnotationFields(),
		},
	}
}

func mendeleyDocumentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "year", Type: "integer"},
		{Name: "abstract", Type: "string"},
		{Name: "profile_id", Type: "string"},
		{Name: "group_id", Type: "string"},
		{Name: "created", Type: "timestamp"},
		{Name: "last_modified", Type: "timestamp"},
	}
}

func mendeleyFolderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "parent_id", Type: "string"},
		{Name: "group_id", Type: "string"},
		{Name: "created", Type: "timestamp"},
		{Name: "modified", Type: "timestamp"},
	}
}

func mendeleyGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "access_level", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "owning_profile_id", Type: "string"},
		{Name: "webpage", Type: "string"},
		{Name: "created", Type: "timestamp"},
	}
}

func mendeleyAnnotationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "text", Type: "string"},
		{Name: "document_id", Type: "string"},
		{Name: "profile_id", Type: "string"},
		{Name: "privacy_level", Type: "string"},
		{Name: "filehash", Type: "string"},
		{Name: "created", Type: "timestamp"},
		{Name: "last_modified", Type: "timestamp"},
	}
}

func mendeleyDocumentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"title":         item["title"],
		"type":          item["type"],
		"source":        item["source"],
		"year":          item["year"],
		"abstract":      item["abstract"],
		"profile_id":    item["profile_id"],
		"group_id":      item["group_id"],
		"created":       item["created"],
		"last_modified": item["last_modified"],
	}
}

func mendeleyFolderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"parent_id": item["parent_id"],
		"group_id":  item["group_id"],
		"created":   item["created"],
		"modified":  item["modified"],
	}
}

func mendeleyGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"description":       item["description"],
		"access_level":      item["access_level"],
		"role":              item["role"],
		"owning_profile_id": item["owning_profile_id"],
		"webpage":           item["webpage"],
		"created":           item["created"],
	}
}

func mendeleyAnnotationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"type":          item["type"],
		"text":          item["text"],
		"document_id":   item["document_id"],
		"profile_id":    item["profile_id"],
		"privacy_level": item["privacy_level"],
		"filehash":      item["filehash"],
		"created":       item["created"],
		"last_modified": item["last_modified"],
	}
}
