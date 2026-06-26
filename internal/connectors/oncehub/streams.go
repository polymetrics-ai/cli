package oncehub

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the OnceHub API resource path (relative to
// base_url) it reads from, the record mapper that flattens its objects, and
// whether the stream supports the last_updated_time.gt incremental filter.
type streamEndpoint struct {
	// resource is the OnceHub list endpoint path (e.g. "/v2/bookings").
	resource string
	// incremental is true when the stream accepts the last_updated_time.gt query
	// filter (only bookings does, per the OnceHub API).
	incremental bool
	// mapRecord flattens a raw OnceHub object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// oncehubStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in oncehubStreams; the read path
// is fully data-driven from this table.
var oncehubStreamEndpoints = map[string]streamEndpoint{
	"bookings":      {resource: "/v2/bookings", incremental: true, mapRecord: oncehubBookingRecord},
	"contacts":      {resource: "/v2/contacts", mapRecord: oncehubContactRecord},
	"booking_pages": {resource: "/v2/booking-pages", mapRecord: oncehubBookingPageRecord},
	"users":         {resource: "/v2/users", mapRecord: oncehubUserRecord},
	"event_types":   {resource: "/v2/event-types", mapRecord: oncehubEventTypeRecord},
}

// oncehubStreams returns the connector's published stream catalog. Every OnceHub
// object exposes a string id, so the primary key is ["id"] across the board.
// bookings and contacts additionally carry a last_updated_time cursor.
func oncehubStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "bookings",
			Description:  "OnceHub scheduled bookings/meetings.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_updated_time"},
			Fields:       oncehubBookingFields(),
		},
		{
			Name:         "contacts",
			Description:  "OnceHub contacts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_updated_time"},
			Fields:       oncehubContactFields(),
		},
		{
			Name:        "booking_pages",
			Description: "OnceHub booking pages.",
			PrimaryKey:  []string{"id"},
			Fields:      oncehubBookingPageFields(),
		},
		{
			Name:        "users",
			Description: "OnceHub account users.",
			PrimaryKey:  []string{"id"},
			Fields:      oncehubUserFields(),
		},
		{
			Name:        "event_types",
			Description: "OnceHub event types.",
			PrimaryKey:  []string{"id"},
			Fields:      oncehubEventTypeFields(),
		},
	}
}

func oncehubBookingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "booking_page", Type: "string"},
		{Name: "event_type", Type: "string"},
		{Name: "contact", Type: "string"},
		{Name: "owner", Type: "string"},
		{Name: "starting_time", Type: "timestamp"},
		{Name: "duration_minutes", Type: "number"},
		{Name: "customer_timezone", Type: "string"},
		{Name: "location_description", Type: "string"},
		{Name: "tracking_id", Type: "string"},
		{Name: "in_trash", Type: "boolean"},
		{Name: "creation_time", Type: "timestamp"},
		{Name: "last_updated_time", Type: "timestamp"},
	}
}

func oncehubContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "mobile_phone", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "owner", Type: "string"},
		{Name: "creation_time", Type: "timestamp"},
		{Name: "last_updated_time", Type: "timestamp"},
	}
}

func oncehubBookingPageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "label", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "active", Type: "boolean"},
	}
}

func oncehubUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "role_name", Type: "string"},
		{Name: "status", Type: "string"},
	}
}

func oncehubEventTypeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "label", Type: "string"},
	}
}

func oncehubBookingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                   item["id"],
		"object":               item["object"],
		"subject":              item["subject"],
		"status":               item["status"],
		"booking_page":         item["booking_page"],
		"event_type":           item["event_type"],
		"contact":              item["contact"],
		"owner":                item["owner"],
		"starting_time":        item["starting_time"],
		"duration_minutes":     item["duration_minutes"],
		"customer_timezone":    item["customer_timezone"],
		"location_description": item["location_description"],
		"tracking_id":          item["tracking_id"],
		"in_trash":             item["in_trash"],
		"creation_time":        item["creation_time"],
		"last_updated_time":    item["last_updated_time"],
	}
}

func oncehubContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"object":            item["object"],
		"first_name":        item["first_name"],
		"email":             item["email"],
		"mobile_phone":      item["mobile_phone"],
		"timezone":          item["timezone"],
		"owner":             item["owner"],
		"creation_time":     item["creation_time"],
		"last_updated_time": item["last_updated_time"],
	}
}

func oncehubBookingPageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"object":   item["object"],
		"name":     item["name"],
		"label":    item["label"],
		"url":      item["url"],
		"timezone": item["timezone"],
		"active":   item["active"],
	}
}

func oncehubUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"object":     item["object"],
		"first_name": item["first_name"],
		"last_name":  item["last_name"],
		"email":      item["email"],
		"role_name":  item["role_name"],
		"status":     item["status"],
	}
}

func oncehubEventTypeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":     item["id"],
		"object": item["object"],
		"name":   item["name"],
		"label":  item["label"],
	}
}
