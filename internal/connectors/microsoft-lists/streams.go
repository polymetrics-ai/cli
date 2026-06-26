package microsoftlists

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Microsoft Graph resource path
// (relative to the site, e.g. "lists" or "lists/{id}/columns") it reads from,
// any extra query params it needs, and the record mapper that flattens its
// OData objects into connectors.Records.
type streamEndpoint struct {
	// resourceFor builds the Graph path segment after /sites/{site_id}/ for the
	// stream. Streams scoped to a single list interpolate the configured
	// list_id; site-scoped streams ignore it.
	resourceFor func(listID string) string
	// needsListID is true when the stream is scoped to a specific list and
	// therefore requires the list_id config to be set.
	needsListID bool
	// query holds extra OData query params merged into every page request
	// (e.g. $expand=fields for list items).
	query map[string]string
	// mapRecord flattens a raw Graph object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streams(); the read path is fully
// data-driven from this table.
var streamEndpoints = map[string]streamEndpoint{
	"lists": {
		resourceFor: func(string) string { return "lists" },
		mapRecord:   listRecord,
	},
	"list_items": {
		resourceFor: func(listID string) string { return "lists/" + listID + "/items" },
		needsListID: true,
		query:       map[string]string{"$expand": "fields"},
		mapRecord:   listItemRecord,
	},
	"columns": {
		resourceFor: func(listID string) string { return "lists/" + listID + "/columns" },
		needsListID: true,
		mapRecord:   columnRecord,
	},
	"content_types": {
		resourceFor: func(listID string) string { return "lists/" + listID + "/contentTypes" },
		needsListID: true,
		mapRecord:   contentTypeRecord,
	},
}

// streams returns the connector's published stream catalog. Every Graph
// baseItem exposes a string id; site lists and list items also carry
// lastModifiedDateTime, which is the incremental cursor where available.
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "lists",
			Description:  "SharePoint/Microsoft Lists in the configured site.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_modified_date_time"},
			Fields:       listFields(),
		},
		{
			Name:         "list_items",
			Description:  "Items within the configured list (requires list_id), with expanded fields.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_modified_date_time"},
			Fields:       listItemFields(),
		},
		{
			Name:        "columns",
			Description: "Column definitions for the configured list (requires list_id).",
			PrimaryKey:  []string{"id"},
			Fields:      columnFields(),
		},
		{
			Name:        "content_types",
			Description: "Content types present in the configured list (requires list_id).",
			PrimaryKey:  []string{"id"},
			Fields:      contentTypeFields(),
		},
	}
}

func listFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "web_url", Type: "string"},
		{Name: "list_template", Type: "string"},
		{Name: "created_date_time", Type: "string"},
		{Name: "last_modified_date_time", Type: "string"},
		{Name: "etag", Type: "string"},
	}
}

func listItemFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "content_type_id", Type: "string"},
		{Name: "web_url", Type: "string"},
		{Name: "created_date_time", Type: "string"},
		{Name: "last_modified_date_time", Type: "string"},
		{Name: "etag", Type: "string"},
		{Name: "fields", Type: "object"},
	}
}

func columnFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "column_group", Type: "string"},
		{Name: "required", Type: "boolean"},
		{Name: "read_only", Type: "boolean"},
		{Name: "hidden", Type: "boolean"},
		{Name: "indexed", Type: "boolean"},
	}
}

func contentTypeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "group", Type: "string"},
		{Name: "hidden", Type: "boolean"},
		{Name: "read_only", Type: "boolean"},
		{Name: "sealed", Type: "boolean"},
	}
}

func listRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":                      item["id"],
		"name":                    item["name"],
		"display_name":            item["displayName"],
		"description":             item["description"],
		"web_url":                 item["webUrl"],
		"created_date_time":       item["createdDateTime"],
		"last_modified_date_time": item["lastModifiedDateTime"],
		"etag":                    item["eTag"],
	}
	// listInfo nests the template under the "list" facet.
	if info, ok := item["list"].(map[string]any); ok {
		rec["list_template"] = info["template"]
	}
	return rec
}

func listItemRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                      item["id"],
		"content_type_id":         contentTypeID(item),
		"web_url":                 item["webUrl"],
		"created_date_time":       item["createdDateTime"],
		"last_modified_date_time": item["lastModifiedDateTime"],
		"etag":                    item["eTag"],
		"fields":                  item["fields"],
	}
}

func columnRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"display_name": item["displayName"],
		"description":  item["description"],
		"column_group": item["columnGroup"],
		"required":     item["required"],
		"read_only":    item["readOnly"],
		"hidden":       item["hidden"],
		"indexed":      item["indexed"],
	}
}

func contentTypeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"group":       item["group"],
		"hidden":      item["hidden"],
		"read_only":   item["readOnly"],
		"sealed":      item["sealed"],
	}
}

// contentTypeID pulls the nested contentType.id off a list item, if present.
func contentTypeID(item map[string]any) any {
	if ct, ok := item["contentType"].(map[string]any); ok {
		return ct["id"]
	}
	return nil
}
