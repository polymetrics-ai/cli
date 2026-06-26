package dixa

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Dixa export resource it reads from and
// the record mapper that flattens its objects. Dixa's public export surface is a
// single endpoint (conversation_export) that returns a top-level JSON array of
// rich conversation objects; the connector projects that one payload into a few
// focused, well-keyed streams.
type streamEndpoint struct {
	// resource is the Dixa export endpoint path segment (always
	// "conversation_export" today).
	resource string
	// mapRecord flattens a raw Dixa conversation object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// dixaStreamEndpoints is the per-stream routing table. Every stream is derived
// from the same conversation_export payload; adding a projection means adding one
// entry here plus a Stream definition in dixaStreams.
var dixaStreamEndpoints = map[string]streamEndpoint{
	"conversations":           {resource: "conversation_export", mapRecord: dixaConversationRecord},
	"conversation_queue":      {resource: "conversation_export", mapRecord: dixaQueueRecord},
	"conversation_rating":     {resource: "conversation_export", mapRecord: dixaRatingRecord},
	"conversation_assignment": {resource: "conversation_export", mapRecord: dixaAssignmentRecord},
}

// dixaStreams returns the connector's published stream catalog. Every projection
// keeps the conversation id as primary key and the millisecond updated_at as the
// incremental cursor field, matching the upstream DatetimeBasedCursor.
func dixaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "conversations",
			Description:  "Dixa conversations exported via the conversation_export API.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       dixaConversationFields(),
		},
		{
			Name:         "conversation_queue",
			Description:  "Queue routing details for each Dixa conversation.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       dixaQueueFields(),
		},
		{
			Name:         "conversation_rating",
			Description:  "CSAT/rating details for each Dixa conversation.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       dixaRatingFields(),
		},
		{
			Name:         "conversation_assignment",
			Description:  "Assignee/agent routing details for each Dixa conversation.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       dixaAssignmentFields(),
		},
	}
}

func dixaConversationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "closed_at", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "direction", Type: "string"},
		{Name: "initial_channel", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "requester_id", Type: "string"},
		{Name: "requester_name", Type: "string"},
		{Name: "requester_email", Type: "string"},
		{Name: "total_duration", Type: "integer"},
		{Name: "handling_duration", Type: "integer"},
		{Name: "last_message_created_at", Type: "integer"},
		{Name: "originating_country", Type: "string"},
	}
}

func dixaQueueFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "queue_id", Type: "string"},
		{Name: "queue_name", Type: "string"},
		{Name: "queued_at", Type: "integer"},
		{Name: "initial_channel", Type: "string"},
		{Name: "direction", Type: "string"},
	}
}

func dixaRatingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "rating_score", Type: "integer"},
		{Name: "rating_message", Type: "string"},
		{Name: "status", Type: "string"},
	}
}

func dixaAssignmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "assigned_at", Type: "integer"},
		{Name: "assignee_id", Type: "string"},
		{Name: "assignee_name", Type: "string"},
		{Name: "assignee_email", Type: "string"},
		{Name: "transferee_name", Type: "string"},
		{Name: "transfer_time", Type: "integer"},
	}
}

func dixaConversationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                      item["id"],
		"created_at":              item["created_at"],
		"updated_at":              item["updated_at"],
		"closed_at":               item["closed_at"],
		"status":                  item["status"],
		"direction":               item["direction"],
		"initial_channel":         item["initial_channel"],
		"subject":                 item["subject"],
		"requester_id":            item["requester_id"],
		"requester_name":          item["requester_name"],
		"requester_email":         item["requester_email"],
		"total_duration":          item["total_duration"],
		"handling_duration":       item["handling_duration"],
		"last_message_created_at": item["last_message_created_at"],
		"originating_country":     item["originating_country"],
	}
}

func dixaQueueRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"updated_at":      item["updated_at"],
		"queue_id":        item["queue_id"],
		"queue_name":      item["queue_name"],
		"queued_at":       item["queued_at"],
		"initial_channel": item["initial_channel"],
		"direction":       item["direction"],
	}
}

func dixaRatingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"updated_at":     item["updated_at"],
		"rating_score":   item["rating_score"],
		"rating_message": item["rating_message"],
		"status":         item["status"],
	}
}

func dixaAssignmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"updated_at":      item["updated_at"],
		"assigned_at":     item["assigned_at"],
		"assignee_id":     item["assignee_id"],
		"assignee_name":   item["assignee_name"],
		"assignee_email":  item["assignee_email"],
		"transferee_name": item["transferee_name"],
		"transfer_time":   item["transfer_time"],
	}
}
