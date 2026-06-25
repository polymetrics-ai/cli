package confluence

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Confluence v2 API resource path
// (relative to base_url, which already ends in /wiki/api/v2) and the record
// mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the v2 list endpoint path segment (e.g. "spaces").
	resource string
	// mapRecord flattens a raw Confluence object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// confluenceStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in confluenceStreams; the
// read path is fully data-driven from this table.
var confluenceStreamEndpoints = map[string]streamEndpoint{
	"spaces":      {resource: "spaces", mapRecord: confluenceSpaceRecord},
	"pages":       {resource: "pages", mapRecord: confluencePageRecord},
	"blogposts":   {resource: "blogposts", mapRecord: confluenceBlogPostRecord},
	"labels":      {resource: "labels", mapRecord: confluenceLabelRecord},
	"attachments": {resource: "attachments", mapRecord: confluenceAttachmentRecord},
}

// confluenceStreams returns the connector's published stream catalog. Confluence
// v2 objects expose a string id and (for content) a createdAt timestamp, so the
// primary key is ["id"] and content streams use ["createdAt"] as the cursor.
func confluenceStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "spaces",
			Description: "Confluence spaces.",
			PrimaryKey:  []string{"id"},
			Fields:      confluenceSpaceFields(),
		},
		{
			Name:         "pages",
			Description:  "Confluence pages.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       confluencePageFields(),
		},
		{
			Name:         "blogposts",
			Description:  "Confluence blog posts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       confluenceBlogPostFields(),
		},
		{
			Name:        "labels",
			Description: "Confluence labels.",
			PrimaryKey:  []string{"id"},
			Fields:      confluenceLabelFields(),
		},
		{
			Name:         "attachments",
			Description:  "Confluence attachments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       confluenceAttachmentFields(),
		},
	}
}

func confluenceSpaceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "key", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "authorId", Type: "string"},
		{Name: "homepageId", Type: "string"},
		{Name: "createdAt", Type: "timestamp"},
	}
}

func confluencePageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "spaceId", Type: "string"},
		{Name: "parentId", Type: "string"},
		{Name: "authorId", Type: "string"},
		{Name: "createdAt", Type: "timestamp"},
		{Name: "version", Type: "integer"},
	}
}

func confluenceBlogPostFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "spaceId", Type: "string"},
		{Name: "authorId", Type: "string"},
		{Name: "createdAt", Type: "timestamp"},
		{Name: "version", Type: "integer"},
	}
}

func confluenceLabelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "prefix", Type: "string"},
	}
}

func confluenceAttachmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "pageId", Type: "string"},
		{Name: "mediaType", Type: "string"},
		{Name: "fileSize", Type: "integer"},
		{Name: "createdAt", Type: "timestamp"},
	}
}

func confluenceSpaceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"key":        item["key"],
		"name":       item["name"],
		"type":       item["type"],
		"status":     item["status"],
		"authorId":   item["authorId"],
		"homepageId": item["homepageId"],
		"createdAt":  item["createdAt"],
	}
}

func confluencePageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"status":    item["status"],
		"title":     item["title"],
		"spaceId":   item["spaceId"],
		"parentId":  item["parentId"],
		"authorId":  item["authorId"],
		"createdAt": item["createdAt"],
		"version":   versionNumber(item["version"]),
	}
}

func confluenceBlogPostRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"status":    item["status"],
		"title":     item["title"],
		"spaceId":   item["spaceId"],
		"authorId":  item["authorId"],
		"createdAt": item["createdAt"],
		"version":   versionNumber(item["version"]),
	}
}

func confluenceLabelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":     item["id"],
		"name":   item["name"],
		"prefix": item["prefix"],
	}
}

func confluenceAttachmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"status":    item["status"],
		"title":     item["title"],
		"pageId":    item["pageId"],
		"mediaType": item["mediaType"],
		"fileSize":  item["fileSize"],
		"createdAt": item["createdAt"],
	}
}

// versionNumber pulls the numeric version out of a Confluence content object's
// nested {"version": {"number": N}} shape, falling back to the raw value.
func versionNumber(v any) any {
	if obj, ok := v.(map[string]any); ok {
		return obj["number"]
	}
	return v
}
