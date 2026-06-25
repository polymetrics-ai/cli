package notion

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Notion API resource and the request
// shape it uses. Notion exposes two list shapes: POST /search (databases, pages)
// and GET /users; both wrap their rows under "results" and paginate with a body
// or query start_cursor against next_cursor/has_more.
type streamEndpoint struct {
	// resource is the API path segment relative to base_url (e.g. "search").
	resource string
	// method is the HTTP verb (POST for /search, GET for /users).
	method string
	// searchObject, when set, is the object filter sent to POST /search
	// ("database" or "page"). Empty for GET endpoints.
	searchObject string
	// mapRecord flattens a raw Notion object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// notionStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in notionStreams; the read path
// is fully data-driven from this table.
var notionStreamEndpoints = map[string]streamEndpoint{
	"databases": {resource: "search", method: "POST", searchObject: "database", mapRecord: notionDatabaseRecord},
	"pages":     {resource: "search", method: "POST", searchObject: "page", mapRecord: notionPageRecord},
	"users":     {resource: "users", method: "GET", mapRecord: notionUserRecord},
}

// notionStreams returns the connector's published stream catalog. Databases and
// pages are versioned objects keyed by id with a last_edited_time incremental
// cursor; users have no edit timestamp so they full-sync on id.
func notionStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "databases",
			Description:  "Notion databases the integration can access (POST /search, object=database).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_edited_time"},
			Fields:       notionObjectFields(),
		},
		{
			Name:         "pages",
			Description:  "Notion pages the integration can access (POST /search, object=page).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_edited_time"},
			Fields:       notionObjectFields(),
		},
		{
			Name:        "users",
			Description: "Notion workspace users and bots (GET /users).",
			PrimaryKey:  []string{"id"},
			Fields:      notionUserFields(),
		},
	}
}

// notionObjectFields describes the shared shape of search results (databases and
// pages). The parent and url fields are object/string respectively.
func notionObjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "created_time", Type: "timestamp"},
		{Name: "last_edited_time", Type: "timestamp"},
		{Name: "archived", Type: "boolean"},
		{Name: "in_trash", Type: "boolean"},
		{Name: "url", Type: "string"},
		{Name: "parent", Type: "object"},
	}
}

func notionUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "avatar_url", Type: "string"},
		{Name: "person", Type: "object"},
		{Name: "bot", Type: "object"},
	}
}

func notionDatabaseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"object":           item["object"],
		"created_time":     item["created_time"],
		"last_edited_time": item["last_edited_time"],
		"archived":         item["archived"],
		"in_trash":         item["in_trash"],
		"url":              item["url"],
		"parent":           item["parent"],
		"title":            item["title"],
	}
}

func notionPageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"object":           item["object"],
		"created_time":     item["created_time"],
		"last_edited_time": item["last_edited_time"],
		"archived":         item["archived"],
		"in_trash":         item["in_trash"],
		"url":              item["url"],
		"parent":           item["parent"],
		"properties":       item["properties"],
	}
}

func notionUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"object":     item["object"],
		"name":       item["name"],
		"type":       item["type"],
		"avatar_url": item["avatar_url"],
		"person":     item["person"],
		"bot":        item["bot"],
	}
}
