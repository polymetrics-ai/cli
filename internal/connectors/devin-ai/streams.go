package devinai

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Devin v3 API resource path (relative
// to base_url and the org segment) and the record mapper that flattens its
// objects. resource is the path suffix appended after
// "v3/organizations/{org_id}/"; e.g. "sessions" -> v3/organizations/<org>/sessions.
type streamEndpoint struct {
	resource string
	// idField is the primary key field name for the stream's objects, used to
	// build deterministic fixture ids.
	idField string
	// mapRecord flattens a raw Devin object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// devinStreamEndpoints is the per-stream routing table. Every Devin v3 list
// endpoint returns {items:[...], has_next_page, end_cursor} and paginates with
// the `after` cursor, so the read path is fully data-driven from this table.
var devinStreamEndpoints = map[string]streamEndpoint{
	"sessions":          {resource: "sessions", idField: "session_id", mapRecord: devinSessionRecord},
	"sessions_insights": {resource: "sessions/insights", idField: "session_id", mapRecord: devinSessionInsightRecord},
	"session_messages":  {resource: "sessions/messages", idField: "message_id", mapRecord: devinSessionMessageRecord},
	"playbooks":         {resource: "playbooks", idField: "playbook_id", mapRecord: devinPlaybookRecord},
	"secrets":           {resource: "secrets", idField: "secret_id", mapRecord: devinSecretRecord},
}

// devinStreams returns the connector's published stream catalog. Session-derived
// streams carry a created_at timestamp usable as an incremental cursor; the
// metadata streams (playbooks, secrets) are full-refresh.
func devinStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "sessions",
			Description:  "Devin sessions in the organization.",
			PrimaryKey:   []string{"session_id"},
			CursorFields: []string{"created_at"},
			Fields:       devinSessionFields(),
		},
		{
			Name:         "sessions_insights",
			Description:  "Devin sessions enriched with message counts and AI analysis.",
			PrimaryKey:   []string{"session_id"},
			CursorFields: []string{"created_at"},
			Fields:       devinSessionInsightFields(),
		},
		{
			Name:         "session_messages",
			Description:  "Conversation messages within Devin sessions.",
			PrimaryKey:   []string{"message_id"},
			CursorFields: []string{"created_at"},
			Fields:       devinSessionMessageFields(),
		},
		{
			Name:        "playbooks",
			Description: "Reusable Devin playbook definitions.",
			PrimaryKey:  []string{"playbook_id"},
			Fields:      devinPlaybookFields(),
		},
		{
			Name:        "secrets",
			Description: "Devin secret metadata (values are never exposed).",
			PrimaryKey:  []string{"secret_id"},
			Fields:      devinSecretFields(),
		},
	}
}

func devinSessionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "session_id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "origin", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "playbook_id", Type: "string"},
		{Name: "is_archived", Type: "boolean"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func devinSessionInsightFields() []connectors.Field {
	return []connectors.Field{
		{Name: "session_id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "message_count", Type: "integer"},
		{Name: "user_id", Type: "string"},
		{Name: "summary", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func devinSessionMessageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "message_id", Type: "string"},
		{Name: "session_id", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "content", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func devinPlaybookFields() []connectors.Field {
	return []connectors.Field{
		{Name: "playbook_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func devinSecretFields() []connectors.Field {
	return []connectors.Field{
		{Name: "secret_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func devinSessionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"session_id":  item["session_id"],
		"title":       item["title"],
		"status":      item["status"],
		"category":    item["category"],
		"origin":      item["origin"],
		"user_id":     item["user_id"],
		"playbook_id": item["playbook_id"],
		"is_archived": item["is_archived"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func devinSessionInsightRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"session_id":    item["session_id"],
		"title":         item["title"],
		"status":        item["status"],
		"message_count": item["message_count"],
		"user_id":       item["user_id"],
		"summary":       item["summary"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func devinSessionMessageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"message_id": item["message_id"],
		"session_id": item["session_id"],
		"role":       item["role"],
		"type":       item["type"],
		"content":    item["content"],
		"created_at": item["created_at"],
	}
}

func devinPlaybookRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"playbook_id": item["playbook_id"],
		"name":        item["name"],
		"description": item["description"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func devinSecretRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"secret_id":  item["secret_id"],
		"name":       item["name"],
		"type":       item["type"],
		"created_at": item["created_at"],
	}
}
