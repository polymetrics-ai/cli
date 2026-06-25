package calcom

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to its Cal.com v2 API resource path and the
// behaviour needed to extract and map its records. The read path is fully
// data-driven from this table: adding a stream means adding one entry here plus a
// Stream definition in calcomStreams.
type streamEndpoint struct {
	// resource is the v2 path segment (relative to base_url), e.g. "bookings".
	resource string
	// recordsPath is the dotted JSON path to the records array/object in the
	// response envelope (Cal.com wraps payloads under "data").
	recordsPath string
	// paginated is true for streams that support offset (skip/take) pagination.
	paginated bool
	// nested marks event-types, whose records live two levels deep under
	// data.eventTypeGroups[].eventTypes[] and need custom flattening.
	nested bool
	// mapRecord flattens a raw Cal.com object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// calcomStreamEndpoints is the per-stream routing table for the CORE stream set.
var calcomStreamEndpoints = map[string]streamEndpoint{
	"bookings":    {resource: "bookings", recordsPath: "data", paginated: true, mapRecord: calcomBookingRecord},
	"event_types": {resource: "event-types", recordsPath: "data.eventTypeGroups", nested: true, mapRecord: calcomEventTypeRecord},
	"schedules":   {resource: "schedules", recordsPath: "data", paginated: true, mapRecord: calcomScheduleRecord},
	"my_profile":  {resource: "me", recordsPath: "data", paginated: false, mapRecord: calcomProfileRecord},
}

// calcomStreams returns the connector's published stream catalog.
func calcomStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "bookings",
			Description: "Cal.com bookings.",
			PrimaryKey:  []string{"id"},
			Fields:      calcomBookingFields(),
		},
		{
			Name:        "event_types",
			Description: "Cal.com event types, flattened from event type groups.",
			PrimaryKey:  []string{"id"},
			Fields:      calcomEventTypeFields(),
		},
		{
			Name:        "schedules",
			Description: "Cal.com availability schedules.",
			PrimaryKey:  []string{"id"},
			Fields:      calcomScheduleFields(),
		},
		{
			Name:        "my_profile",
			Description: "The authenticated Cal.com user profile.",
			PrimaryKey:  []string{"id"},
			Fields:      calcomProfileFields(),
		},
	}
}

func calcomBookingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "uid", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "start", Type: "string"},
		{Name: "end", Type: "string"},
		{Name: "eventTypeId", Type: "integer"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func calcomEventTypeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "slug", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "length", Type: "integer"},
		{Name: "hidden", Type: "boolean"},
		{Name: "position", Type: "integer"},
	}
}

func calcomScheduleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "timeZone", Type: "string"},
		{Name: "isDefault", Type: "boolean"},
		{Name: "ownerId", Type: "integer"},
	}
}

func calcomProfileFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "username", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "timeZone", Type: "string"},
		{Name: "timeFormat", Type: "integer"},
		{Name: "weekStart", Type: "string"},
	}
}

func calcomBookingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"uid":         item["uid"],
		"title":       item["title"],
		"description": item["description"],
		"status":      item["status"],
		"start":       item["start"],
		"end":         item["end"],
		"eventTypeId": item["eventTypeId"],
		"createdAt":   item["createdAt"],
		"updatedAt":   item["updatedAt"],
	}
}

func calcomEventTypeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"slug":        item["slug"],
		"title":       item["title"],
		"description": item["description"],
		"length":      item["length"],
		"hidden":      item["hidden"],
		"position":    item["position"],
	}
}

func calcomScheduleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"timeZone":  item["timeZone"],
		"isDefault": item["isDefault"],
		"ownerId":   item["ownerId"],
	}
}

func calcomProfileRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"username":   item["username"],
		"email":      item["email"],
		"name":       item["name"],
		"timeZone":   item["timeZone"],
		"timeFormat": item["timeFormat"],
		"weekStart":  item["weekStart"],
	}
}
