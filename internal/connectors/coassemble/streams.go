package coassemble

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Coassemble API resource path (relative
// to base_url) it reads from, whether that endpoint is paginated, and the record
// mapper that shapes its objects.
type streamEndpoint struct {
	// path is the Coassemble headless API path (e.g. "/api/v1/headless/courses").
	path string
	// paginated is true when the endpoint supports the page/length paginator.
	// screen_types is served as a single un-paginated array.
	paginated bool
	// mapRecord shapes a raw Coassemble object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// coassembleStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in coassembleStreams; the read
// path is fully data-driven from this table.
var coassembleStreamEndpoints = map[string]streamEndpoint{
	"courses":      {path: "/api/v1/headless/courses", paginated: true, mapRecord: coassembleCourseRecord},
	"screen_types": {path: "/api/v1/headless/screen/types", paginated: false, mapRecord: coassembleScreenTypeRecord},
	"trackings":    {path: "/api/v1/headless/trackings", paginated: true, mapRecord: coassembleTrackingRecord},
}

// coassembleStreams returns the connector's published stream catalog. Coassemble's
// headless API exposes no incremental cursor, so every stream is a full refresh
// (CursorFields is empty). Only courses carries a stable primary key.
func coassembleStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "courses",
			Description: "Coassemble courses in the workspace.",
			PrimaryKey:  []string{"id"},
			Fields:      coassembleCourseFields(),
		},
		{
			Name:        "screen_types",
			Description: "Coassemble screen types (available on request only).",
			Fields:      coassembleScreenTypeFields(),
		},
		{
			Name:        "trackings",
			Description: "Coassemble learner tracking records (available on request only).",
			Fields:      coassembleTrackingFields(),
		},
	}
}

func coassembleCourseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "key", Type: "string"},
		{Name: "image", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "private", Type: "boolean"},
		{Name: "paid", Type: "boolean"},
		{Name: "identified", Type: "boolean"},
		{Name: "is_sharable", Type: "boolean"},
	}
}

func coassembleScreenTypeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "icon", Type: "string"},
		{Name: "premium", Type: "boolean"},
	}
}

func coassembleTrackingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "course_id", Type: "integer"},
		{Name: "identifier", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "progress", Type: "number"},
		{Name: "completed", Type: "boolean"},
	}
}

func coassembleCourseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"title":       item["title"],
		"description": item["description"],
		"key":         item["key"],
		"image":       item["image"],
		"active":      item["active"],
		"private":     item["private"],
		"paid":        item["paid"],
		"identified":  item["identified"],
		"is_sharable": item["is_sharable"],
	}
}

func coassembleScreenTypeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      item["id"],
		"name":    item["name"],
		"title":   item["title"],
		"icon":    item["icon"],
		"premium": item["premium"],
	}
}

func coassembleTrackingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"course_id":  item["course_id"],
		"identifier": item["identifier"],
		"status":     item["status"],
		"progress":   item["progress"],
		"completed":  item["completed"],
	}
}
