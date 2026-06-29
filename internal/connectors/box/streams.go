package box

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Box API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its entries. Box
// list endpoints all share the {entries:[...], offset, limit, total_count}
// envelope, so the offset paginator in box.go is shared across every entry.
type streamEndpoint struct {
	// resource is the Box list endpoint path segment (e.g. "users"). It may
	// contain a placeholder for folder_items, resolved at read time.
	resource string
	// mapRecord flattens a raw Box entry into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// folderScoped is true for endpoints addressed under a folder id
	// (folders/{id}/items); the resource is computed from the folder_id config.
	folderScoped bool
}

// boxStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in boxStreams; the read path is
// fully data-driven from this table.
var boxStreamEndpoints = map[string]streamEndpoint{
	"users":        {resource: "users", mapRecord: boxUserRecord},
	"groups":       {resource: "groups", mapRecord: boxGroupRecord},
	"collections":  {resource: "collections", mapRecord: boxCollectionRecord},
	"folder_items": {resource: "folders/0/items", mapRecord: boxItemRecord, folderScoped: true},
}

// boxStreams returns the connector's published stream catalog. Box objects expose
// a string id, so the primary key is ["id"] across the board. Box's manifest-only
// upstream source supports full_refresh only; the modified_at/created_at fields are
// still surfaced as cursor candidates for incremental users of this connector.
func boxStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "users",
			Description:  "Box enterprise users.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified_at"},
			Fields:       boxUserFields(),
		},
		{
			Name:         "groups",
			Description:  "Box groups.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified_at"},
			Fields:       boxGroupFields(),
		},
		{
			Name:         "collections",
			Description:  "Box collections (e.g. Favorites).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{},
			Fields:       boxCollectionFields(),
		},
		{
			Name:         "folder_items",
			Description:  "Items (files, folders, web links) inside a Box folder (folder_id config, defaults to root folder 0).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified_at"},
			Fields:       boxItemFields(),
		},
	}
}

func boxUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "login", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
	}
}

func boxGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "group_type", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
	}
}

func boxCollectionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "collection_type", Type: "string"},
	}
}

func boxItemFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "etag", Type: "string"},
		{Name: "sequence_id", Type: "string"},
		{Name: "sha1", Type: "string"},
		{Name: "size", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
	}
}

func boxUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"type":        item["type"],
		"name":        item["name"],
		"login":       item["login"],
		"status":      item["status"],
		"language":    item["language"],
		"timezone":    item["timezone"],
		"created_at":  item["created_at"],
		"modified_at": item["modified_at"],
	}
}

func boxGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"type":        item["type"],
		"name":        item["name"],
		"group_type":  item["group_type"],
		"created_at":  item["created_at"],
		"modified_at": item["modified_at"],
	}
}

func boxCollectionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"type":            item["type"],
		"name":            item["name"],
		"collection_type": item["collection_type"],
	}
}

func boxItemRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"type":        item["type"],
		"name":        item["name"],
		"etag":        item["etag"],
		"sequence_id": item["sequence_id"],
		"sha1":        item["sha1"],
		"size":        item["size"],
		"created_at":  item["created_at"],
		"modified_at": item["modified_at"],
	}
}
