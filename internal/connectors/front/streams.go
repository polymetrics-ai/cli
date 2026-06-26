package front

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Front API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Front list endpoint path segment (e.g. "contacts").
	resource string
	// mapRecord flattens a raw Front object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// frontStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in frontStreams; the read path
// is fully data-driven from this table. Every Front list response wraps its
// objects in {_results:[...], _pagination:{next:<url>}}.
var frontStreamEndpoints = map[string]streamEndpoint{
	"contacts":      {resource: "contacts", mapRecord: frontContactRecord},
	"conversations": {resource: "conversations", mapRecord: frontConversationRecord},
	"inboxes":       {resource: "inboxes", mapRecord: frontInboxRecord},
	"tags":          {resource: "tags", mapRecord: frontTagRecord},
	"teammates":     {resource: "teammates", mapRecord: frontTeammateRecord},
	"channels":      {resource: "channels", mapRecord: frontChannelRecord},
}

// frontStreams returns the connector's published stream catalog. Front objects
// expose a string id; resources that track recency expose unix timestamps
// (created_at / last_message_at) used as incremental cursor fields.
func frontStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "Front contacts (people and accounts the team communicates with).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       frontContactFields(),
		},
		{
			Name:         "conversations",
			Description:  "Front conversations (threads across all inboxes).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_message_at"},
			Fields:       frontConversationFields(),
		},
		{
			Name:        "inboxes",
			Description: "Front inboxes (shared mailboxes and channels).",
			PrimaryKey:  []string{"id"},
			Fields:      frontInboxFields(),
		},
		{
			Name:        "tags",
			Description: "Front tags used to categorize conversations.",
			PrimaryKey:  []string{"id"},
			Fields:      frontTagFields(),
		},
		{
			Name:        "teammates",
			Description: "Front teammates (members of the workspace).",
			PrimaryKey:  []string{"id"},
			Fields:      frontTeammateFields(),
		},
		{
			Name:        "channels",
			Description: "Front channels (the underlying transport for an inbox).",
			PrimaryKey:  []string{"id"},
			Fields:      frontChannelFields(),
		},
	}
}

func frontContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "is_spammer", Type: "boolean"},
		{Name: "is_private", Type: "boolean"},
		{Name: "created_at", Type: "number"},
		{Name: "updated_at", Type: "number"},
	}
}

func frontConversationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "is_private", Type: "boolean"},
		{Name: "created_at", Type: "number"},
		{Name: "last_message_at", Type: "number"},
		{Name: "waiting_since", Type: "number"},
	}
}

func frontInboxFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "is_private", Type: "boolean"},
		{Name: "is_public", Type: "boolean"},
		{Name: "custom_fields", Type: "object"},
	}
}

func frontTagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "highlight", Type: "string"},
		{Name: "is_private", Type: "boolean"},
		{Name: "is_visible_in_conversation_lists", Type: "boolean"},
		{Name: "created_at", Type: "number"},
		{Name: "updated_at", Type: "number"},
	}
}

func frontTeammateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "username", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "is_admin", Type: "boolean"},
		{Name: "is_available", Type: "boolean"},
		{Name: "is_blocked", Type: "boolean"},
	}
}

func frontChannelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "address", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "send_as", Type: "string"},
		{Name: "is_private", Type: "boolean"},
		{Name: "is_valid", Type: "boolean"},
	}
}

func frontContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"is_spammer":  item["is_spammer"],
		"is_private":  item["is_private"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func frontConversationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"subject":         item["subject"],
		"status":          item["status"],
		"is_private":      item["is_private"],
		"created_at":      item["created_at"],
		"last_message_at": item["last_message_at"],
		"waiting_since":   item["waiting_since"],
	}
}

func frontInboxRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"is_private":    item["is_private"],
		"is_public":     item["is_public"],
		"custom_fields": item["custom_fields"],
	}
}

func frontTagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                               item["id"],
		"name":                             item["name"],
		"highlight":                        item["highlight"],
		"is_private":                       item["is_private"],
		"is_visible_in_conversation_lists": item["is_visible_in_conversation_lists"],
		"created_at":                       item["created_at"],
		"updated_at":                       item["updated_at"],
	}
}

func frontTeammateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"email":        item["email"],
		"username":     item["username"],
		"first_name":   item["first_name"],
		"last_name":    item["last_name"],
		"is_admin":     item["is_admin"],
		"is_available": item["is_available"],
		"is_blocked":   item["is_blocked"],
	}
}

func frontChannelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"address":    item["address"],
		"type":       item["type"],
		"send_as":    item["send_as"],
		"is_private": item["is_private"],
		"is_valid":   item["is_valid"],
	}
}
