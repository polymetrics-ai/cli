package kisi

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Kisi API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
// Kisi list endpoints return a top-level JSON array (field_path: []).
type streamEndpoint struct {
	// resource is the Kisi list endpoint path segment (e.g. "members").
	resource string
	// mapRecord flattens a raw Kisi object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// kisiStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in kisiStreams; the read path
// is fully data-driven from this table.
var kisiStreamEndpoints = map[string]streamEndpoint{
	"members": {resource: "members", mapRecord: kisiMemberRecord},
	"locks":   {resource: "locks", mapRecord: kisiLockRecord},
	"groups":  {resource: "groups", mapRecord: kisiGroupRecord},
	"users":   {resource: "users", mapRecord: kisiUserRecord},
	"logins":  {resource: "logins", mapRecord: kisiLoginRecord},
}

// kisiStreams returns the connector's published stream catalog. Kisi objects
// expose a numeric id primary key. The API is full-refresh only (no incremental
// cursor), so CursorFields is empty.
func kisiStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "members",
			Description: "Kisi organization members and their access.",
			PrimaryKey:  []string{"id"},
			Fields:      kisiMemberFields(),
		},
		{
			Name:        "locks",
			Description: "Kisi locks (doors / access points).",
			PrimaryKey:  []string{"id"},
			Fields:      kisiLockFields(),
		},
		{
			Name:        "groups",
			Description: "Kisi access groups.",
			PrimaryKey:  []string{"id"},
			Fields:      kisiGroupFields(),
		},
		{
			Name:        "users",
			Description: "Kisi users.",
			PrimaryKey:  []string{"id"},
			Fields:      kisiUserFields(),
		},
		{
			Name:        "logins",
			Description: "Kisi login sessions / API logins.",
			PrimaryKey:  []string{"id"},
			Fields:      kisiLoginFields(),
		},
	}
}

func kisiMemberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "role_id", Type: "integer"},
		{Name: "confirmed", Type: "boolean"},
		{Name: "access_enabled", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func kisiLockFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "place_id", Type: "integer"},
		{Name: "geofence_restriction_enabled", Type: "boolean"},
		{Name: "online", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func kisiGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "place_id", Type: "integer"},
		{Name: "login_count", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func kisiUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "confirmed", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func kisiLoginFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "user_id", Type: "integer"},
		{Name: "last_used_at", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func kisiMemberRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"email":          item["email"],
		"role_id":        item["role_id"],
		"confirmed":      item["confirmed"],
		"access_enabled": item["access_enabled"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
	}
}

func kisiLockRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                           item["id"],
		"name":                         item["name"],
		"description":                  item["description"],
		"place_id":                     item["place_id"],
		"geofence_restriction_enabled": item["geofence_restriction_enabled"],
		"online":                       item["online"],
		"created_at":                   item["created_at"],
		"updated_at":                   item["updated_at"],
	}
}

func kisiGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"place_id":    item["place_id"],
		"login_count": item["login_count"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func kisiUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"email":      item["email"],
		"confirmed":  item["confirmed"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func kisiLoginRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"type":         item["type"],
		"user_id":      item["user_id"],
		"last_used_at": item["last_used_at"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}
