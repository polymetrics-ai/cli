package gologin

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the GoLogin API resource path (relative to
// base_url), the JSON path to the records array within the response, whether the
// endpoint is page-paginated, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the GoLogin API path segment (e.g. "browser/v2").
	resource string
	// recordsPath is the dotted JSON path to the records array. An empty string
	// selects the root (used for endpoints that return a bare array or a single
	// object).
	recordsPath string
	// paginated is true when the endpoint supports ?page=N pagination over a
	// fixed page size. Single-object / small-list endpoints set this false.
	paginated bool
	// mapRecord flattens a raw GoLogin object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// gologinStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in gologinStreams; the read path
// is fully data-driven from this table.
//
// The core set covers the standalone (non-substream) GoLogin resources:
//   - profiles: paginated list under {"profiles":[...]}.
//   - folders:  root array of folders.
//   - user:     single account object (root).
//   - tags:     list under {"tags":[...]}.
var gologinStreamEndpoints = map[string]streamEndpoint{
	"profiles": {resource: "browser/v2", recordsPath: "profiles", paginated: true, mapRecord: gologinProfileRecord},
	"folders":  {resource: "folders", recordsPath: "", paginated: false, mapRecord: gologinFolderRecord},
	"user":     {resource: "user", recordsPath: "", paginated: false, mapRecord: gologinUserRecord},
	"tags":     {resource: "tags/all", recordsPath: "tags", paginated: false, mapRecord: gologinTagRecord},
}

// gologinStreams returns the connector's published stream catalog.
func gologinStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "profiles",
			Description:  "GoLogin browser profiles.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       gologinProfileFields(),
		},
		{
			Name:        "folders",
			Description: "GoLogin profile folders.",
			PrimaryKey:  []string{"id"},
			Fields:      gologinFolderFields(),
		},
		{
			Name:         "user",
			Description:  "GoLogin account / user information.",
			PrimaryKey:   []string{"_id"},
			CursorFields: []string{"createdAt"},
			Fields:       gologinUserFields(),
		},
		{
			Name:        "tags",
			Description: "GoLogin profile tags.",
			PrimaryKey:  []string{"_id"},
			Fields:      gologinTagFields(),
		},
	}
}

func gologinProfileFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "notes", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "os", Type: "string"},
		{Name: "browserType", Type: "string"},
		{Name: "folderName", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func gologinFolderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "profilesCount", Type: "integer"},
	}
}

func gologinUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "plan", Type: "string"},
		{Name: "profilesCount", Type: "integer"},
		{Name: "createdAt", Type: "string"},
	}
}

func gologinTagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "field", Type: "string"},
	}
}

func gologinProfileRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"notes":       item["notes"],
		"role":        item["role"],
		"os":          item["os"],
		"browserType": item["browserType"],
		"folderName":  item["folderName"],
		"createdAt":   item["createdAt"],
		"updatedAt":   item["updatedAt"],
	}
}

func gologinFolderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"profilesCount": item["profilesCount"],
	}
}

func gologinUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":           item["_id"],
		"email":         item["email"],
		"firstName":     item["firstName"],
		"lastName":      item["lastName"],
		"plan":          item["plan"],
		"profilesCount": item["profilesCount"],
		"createdAt":     item["createdAt"],
	}
}

func gologinTagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":   item["_id"],
		"title": item["title"],
		"color": item["color"],
		"field": item["field"],
	}
}
