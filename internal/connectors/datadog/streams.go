package datadog

import "polymetrics.ai/internal/connectors"

// pageStyle describes how a Datadog endpoint paginates. Datadog is not uniform:
// v1 list endpoints use a 0-based page + page_size and a top-level array; v2
// list endpoints use page[number]/page[size] and a {data:[...]} envelope; a few
// endpoints (dashboards, downtimes) return everything in one unpaginated call.
type pageStyle int

const (
	// pageNone returns all records in a single response (no pagination params).
	pageNone pageStyle = iota
	// pageV1 uses page=<0-based> + page_size and a top-level JSON array.
	pageV1
	// pageV2 uses page[number]=<0-based> + page[size] and a {data:[...]} envelope.
	pageV2
)

// streamEndpoint maps a stream to its Datadog API resource path, the JSON path
// to the records array, its pagination style, and the record mapper.
type streamEndpoint struct {
	// resource is the API path relative to base_url, e.g. "api/v1/monitor".
	resource string
	// recordsPath is the dotted JSON path to the records array ("" = root array,
	// "data" = body["data"], "dashboards" = body["dashboards"]).
	recordsPath string
	page        pageStyle
	mapRecord   func(map[string]any) connectors.Record
}

// datadogStreamEndpoints is the per-stream routing table. The read path is fully
// data-driven from this table: add a stream by adding one entry here plus a
// Stream definition in datadogStreams.
var datadogStreamEndpoints = map[string]streamEndpoint{
	"monitors":   {resource: "api/v1/monitor", recordsPath: "", page: pageV1, mapRecord: datadogMonitorRecord},
	"dashboards": {resource: "api/v1/dashboard", recordsPath: "dashboards", page: pageNone, mapRecord: datadogDashboardRecord},
	"users":      {resource: "api/v2/users", recordsPath: "data", page: pageV2, mapRecord: datadogUserRecord},
	"slo":        {resource: "api/v1/slo", recordsPath: "data", page: pageV1, mapRecord: datadogSLORecord},
	"downtimes":  {resource: "api/v1/downtime", recordsPath: "", page: pageNone, mapRecord: datadogDowntimeRecord},
}

// datadogStreams returns the connector's published stream catalog.
func datadogStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "monitors",
			Description:  "Datadog monitors (alert definitions and their current state).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modified"},
			Fields:       datadogMonitorFields(),
		},
		{
			Name:        "dashboards",
			Description: "Datadog dashboards.",
			PrimaryKey:  []string{"id"},
			Fields:      datadogDashboardFields(),
		},
		{
			Name:        "users",
			Description: "Datadog organization users.",
			PrimaryKey:  []string{"id"},
			Fields:      datadogUserFields(),
		},
		{
			Name:        "slo",
			Description: "Datadog service level objectives.",
			PrimaryKey:  []string{"id"},
			Fields:      datadogSLOFields(),
		},
		{
			Name:        "downtimes",
			Description: "Datadog scheduled downtimes.",
			PrimaryKey:  []string{"id"},
			Fields:      datadogDowntimeFields(),
		},
	}
}

func datadogMonitorFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "query", Type: "string"},
		{Name: "message", Type: "string"},
		{Name: "overall_state", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
		{Name: "priority", Type: "integer"},
	}
}

func datadogDashboardFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "layout_type", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "author_handle", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
		{Name: "is_read_only", Type: "boolean"},
	}
}

func datadogUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "handle", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "disabled", Type: "boolean"},
		{Name: "verified", Type: "boolean"},
		{Name: "created_at", Type: "string"},
	}
}

func datadogSLOFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "modified_at", Type: "integer"},
	}
}

func datadogDowntimeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "scope", Type: "string"},
		{Name: "message", Type: "string"},
		{Name: "monitor_id", Type: "integer"},
		{Name: "active", Type: "boolean"},
		{Name: "disabled", Type: "boolean"},
		{Name: "start", Type: "integer"},
		{Name: "end", Type: "integer"},
	}
}

func datadogMonitorRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"type":          item["type"],
		"query":         item["query"],
		"message":       item["message"],
		"overall_state": item["overall_state"],
		"created":       item["created"],
		"modified":      item["modified"],
		"priority":      item["priority"],
	}
}

func datadogDashboardRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"title":         item["title"],
		"description":   item["description"],
		"layout_type":   item["layout_type"],
		"url":           item["url"],
		"author_handle": item["author_handle"],
		"created_at":    item["created_at"],
		"modified_at":   item["modified_at"],
		"is_read_only":  item["is_read_only"],
	}
}

// datadogUserRecord flattens the v2 JSON:API {id, type, attributes:{...}} shape
// into a flat record so cursor/PK fields live at the top level.
func datadogUserRecord(item map[string]any) connectors.Record {
	attrs := mapField(item, "attributes")
	return connectors.Record{
		"id":         item["id"],
		"type":       item["type"],
		"name":       attrs["name"],
		"email":      attrs["email"],
		"handle":     attrs["handle"],
		"status":     attrs["status"],
		"disabled":   attrs["disabled"],
		"verified":   attrs["verified"],
		"created_at": attrs["created_at"],
	}
}

func datadogSLORecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"type":        item["type"],
		"description": item["description"],
		"created_at":  item["created_at"],
		"modified_at": item["modified_at"],
	}
}

func datadogDowntimeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"scope":      item["scope"],
		"message":    item["message"],
		"monitor_id": item["monitor_id"],
		"active":     item["active"],
		"disabled":   item["disabled"],
		"start":      item["start"],
		"end":        item["end"],
	}
}

// mapField returns item[key] as a map, or an empty map if absent/not an object.
func mapField(item map[string]any, key string) map[string]any {
	if m, ok := item[key].(map[string]any); ok {
		return m
	}
	return map[string]any{}
}
