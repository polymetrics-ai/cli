package onehms

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the 100ms API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the 100ms list endpoint path segment (e.g. "rooms").
	resource string
	// mapRecord flattens a raw 100ms object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streams(); the read path is fully
// data-driven from this table. All 100ms list endpoints share the same
// {data:[...], last:"<cursor>"} envelope, so the harvest loop is stream-agnostic.
var streamEndpoints = map[string]streamEndpoint{
	"rooms":      {resource: "rooms", mapRecord: roomRecord},
	"sessions":   {resource: "sessions", mapRecord: sessionRecord},
	"recordings": {resource: "recordings", mapRecord: recordingRecord},
	"templates":  {resource: "templates", mapRecord: templateRecord},
}

// streams returns the connector's published stream catalog. Every 100ms object
// exposes a string id and an RFC3339 created_at timestamp, so the primary key is
// ["id"] and the incremental cursor field is ["created_at"] across the board.
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "rooms",
			Description:  "100ms rooms (conferencing rooms configured for an account).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       roomFields(),
		},
		{
			Name:         "sessions",
			Description:  "100ms sessions (a single live instance of a room).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       sessionFields(),
		},
		{
			Name:         "recordings",
			Description:  "100ms recordings produced from sessions.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       recordingFields(),
		},
		{
			Name:         "templates",
			Description:  "100ms templates (reusable room configuration presets).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       templateFields(),
		},
	}
}

func roomFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "enabled", Type: "boolean"},
		{Name: "description", Type: "string"},
		{Name: "customer_id", Type: "string"},
		{Name: "template_id", Type: "string"},
		{Name: "region", Type: "string"},
		{Name: "large_room", Type: "boolean"},
		{Name: "max_duration_seconds", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func sessionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "room_id", Type: "string"},
		{Name: "customer_id", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func recordingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "room_id", Type: "string"},
		{Name: "session_id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "size", Type: "integer"},
		{Name: "duration", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func templateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "customer_id", Type: "string"},
		{Name: "default", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func roomRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                   item["id"],
		"name":                 item["name"],
		"enabled":              item["enabled"],
		"description":          item["description"],
		"customer_id":          item["customer_id"],
		"template_id":          item["template_id"],
		"region":               item["region"],
		"large_room":           item["large_room"],
		"max_duration_seconds": item["max_duration_seconds"],
		"created_at":           item["created_at"],
		"updated_at":           item["updated_at"],
	}
}

func sessionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"room_id":     item["room_id"],
		"customer_id": item["customer_id"],
		"active":      item["active"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func recordingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"room_id":    item["room_id"],
		"session_id": item["session_id"],
		"status":     item["status"],
		"size":       item["size"],
		"duration":   item["duration"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func templateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"customer_id": item["customer_id"],
		"default":     item["default"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}
