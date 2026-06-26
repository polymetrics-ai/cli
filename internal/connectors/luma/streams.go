package luma

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Luma API resource path (relative to
// base_url) it reads from, the JSON path to the wrapper array (always "entries"
// for Luma), the key inside each entry that holds the actual object (Luma wraps
// records as entries[].event / entries[].guest / entries[].host), and the record
// mapper that flattens those objects.
type streamEndpoint struct {
	// resource is the Luma list endpoint path segment (e.g. "calendar/list-events").
	resource string
	// entryKey is the field inside each `entries` element holding the record
	// (e.g. "event", "guest"). Empty means the entry itself is the record.
	entryKey string
	// requiresEventID is true for sub-streams scoped to a single event; the
	// connector threads the configured event_api_id query param.
	requiresEventID bool
	// mapRecord flattens a raw Luma object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// lumaStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in lumaStreams; the read path is
// fully data-driven from this table.
var lumaStreamEndpoints = map[string]streamEndpoint{
	"events": {
		resource:  "calendar/list-events",
		entryKey:  "event",
		mapRecord: lumaEventRecord,
	},
	"event_guests": {
		resource:        "event/get-guests",
		entryKey:        "guest",
		requiresEventID: true,
		mapRecord:       lumaGuestRecord,
	},
	"event_hosts": {
		resource:        "event/get-hosts",
		entryKey:        "host",
		requiresEventID: true,
		mapRecord:       lumaHostRecord,
	},
}

// lumaStreams returns the connector's published stream catalog. Every Luma object
// is keyed by a string api_id, so the primary key is ["api_id"] across the board.
// The Luma public API only supports full-refresh syncs, so no cursor fields are
// declared.
func lumaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "events",
			Description: "Events on the calendars accessible to the API key.",
			PrimaryKey:  []string{"api_id"},
			Fields:      lumaEventFields(),
		},
		{
			Name:        "event_guests",
			Description: "Guests for a single event (set config event_api_id).",
			PrimaryKey:  []string{"api_id"},
			Fields:      lumaGuestFields(),
		},
		{
			Name:        "event_hosts",
			Description: "Hosts for a single event (set config event_api_id).",
			PrimaryKey:  []string{"api_id"},
			Fields:      lumaHostFields(),
		},
	}
}

func lumaEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "api_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "start_at", Type: "timestamp"},
		{Name: "end_at", Type: "timestamp"},
		{Name: "timezone", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "cover_url", Type: "string"},
		{Name: "visibility", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "calendar_api_id", Type: "string"},
	}
}

func lumaGuestFields() []connectors.Field {
	return []connectors.Field{
		{Name: "api_id", Type: "string"},
		{Name: "event_api_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "approval_status", Type: "string"},
		{Name: "registered_at", Type: "timestamp"},
		{Name: "checked_in_at", Type: "timestamp"},
		{Name: "user_api_id", Type: "string"},
		{Name: "user_name", Type: "string"},
	}
}

func lumaHostFields() []connectors.Field {
	return []connectors.Field{
		{Name: "api_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "avatar_url", Type: "string"},
		{Name: "access_level", Type: "string"},
	}
}

func lumaEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"api_id":          item["api_id"],
		"name":            item["name"],
		"description":     item["description"],
		"start_at":        item["start_at"],
		"end_at":          item["end_at"],
		"timezone":        item["timezone"],
		"url":             item["url"],
		"cover_url":       item["cover_url"],
		"visibility":      item["visibility"],
		"created_at":      item["created_at"],
		"calendar_api_id": item["calendar_api_id"],
	}
}

func lumaGuestRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"api_id":          item["api_id"],
		"event_api_id":    item["event_api_id"],
		"name":            item["name"],
		"email":           item["email"],
		"approval_status": item["approval_status"],
		"registered_at":   item["registered_at"],
		"checked_in_at":   item["checked_in_at"],
		"user_api_id":     item["user_api_id"],
		"user_name":       item["user_name"],
	}
}

func lumaHostRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"api_id":       item["api_id"],
		"name":         item["name"],
		"email":        item["email"],
		"avatar_url":   item["avatar_url"],
		"access_level": item["access_level"],
	}
}
