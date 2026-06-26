package fullstory

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the FullStory API resource path (relative
// to base_url), the JSON path to the records array in the response, and the
// record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "segments/v2").
	resource string
	// recordsPath is the dotted JSON path to the array of records in the
	// response body (FullStory list endpoints return {"results":[...]}).
	recordsPath string
	// mapRecord flattens a raw FullStory object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// fullstoryStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in fullstoryStreams; the read
// path is fully data-driven from this table.
var fullstoryStreamEndpoints = map[string]streamEndpoint{
	"segments": {resource: "segments/v2", recordsPath: "results", mapRecord: fullstorySegmentRecord},
	"users":    {resource: "v2/users", recordsPath: "results", mapRecord: fullstoryUserRecord},
	"events":   {resource: "v2/events", recordsPath: "results", mapRecord: fullstoryEventRecord},
}

// fullstoryStreams returns the connector's published stream catalog. FullStory
// only supports full_refresh sync, so cursor fields are advisory (the API has no
// universal incremental filter), but each object carries a stable string id.
func fullstoryStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "segments",
			Description:  "FullStory segments (saved session filters).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       fullstorySegmentFields(),
		},
		{
			Name:         "users",
			Description:  "FullStory users (identified end users).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       fullstoryUserFields(),
		},
		{
			Name:         "events",
			Description:  "FullStory captured events.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"event_time"},
			Fields:       fullstoryEventFields(),
		},
	}
}

func fullstorySegmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "creator", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "is_public", Type: "boolean"},
		{Name: "type", Type: "string"},
	}
}

func fullstoryUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "uid", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "updated", Type: "string"},
		{Name: "is_being_processed", Type: "boolean"},
	}
}

func fullstoryEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "session_id", Type: "string"},
		{Name: "event_time", Type: "string"},
		{Name: "device_id", Type: "string"},
		{Name: "type", Type: "string"},
	}
}

func fullstorySegmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"creator":     item["creator"],
		"created":     item["created"],
		"is_public":   item["is_public"],
		"type":        item["type"],
	}
}

func fullstoryUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"uid":                item["uid"],
		"display_name":       item["display_name"],
		"email":              item["email"],
		"created":            item["created"],
		"updated":            item["updated"],
		"is_being_processed": item["is_being_processed"],
	}
}

func fullstoryEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"user_id":    item["user_id"],
		"session_id": item["session_id"],
		"event_time": item["event_time"],
		"device_id":  item["device_id"],
		"type":       item["type"],
	}
}
