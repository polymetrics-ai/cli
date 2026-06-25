package zendeskchat

import "polymetrics/internal/connectors"

// pagination distinguishes the two read shapes Zendesk Chat exposes.
type pagination int

const (
	// paginateArray streams return a single top-level JSON array (the full list).
	paginateArray pagination = iota
	// paginateChats streams use the incremental-export shape
	// {chats:[...], next_url:..., count:N} and follow next_url until it is empty.
	paginateChats
)

// streamEndpoint maps a stream name to the Zendesk Chat API resource path
// (relative to base_url), its pagination shape, the JSON path the records live
// at, and the record mapper that flattens its objects.
type streamEndpoint struct {
	resource    string
	pagination  pagination
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

// zendeskChatStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in zendeskChatStreams; the
// read path is fully data-driven from this table.
var zendeskChatStreamEndpoints = map[string]streamEndpoint{
	"agents":      {resource: "agents", pagination: paginateArray, recordsPath: ".", mapRecord: zendeskChatAgentRecord},
	"chats":       {resource: "chats", pagination: paginateChats, recordsPath: "chats", mapRecord: zendeskChatChatRecord},
	"departments": {resource: "departments", pagination: paginateArray, recordsPath: ".", mapRecord: zendeskChatDepartmentRecord},
	"shortcuts":   {resource: "shortcuts", pagination: paginateArray, recordsPath: ".", mapRecord: zendeskChatShortcutRecord},
	"triggers":    {resource: "triggers", pagination: paginateArray, recordsPath: ".", mapRecord: zendeskChatTriggerRecord},
}

// zendeskChatStreams returns the connector's published stream catalog.
func zendeskChatStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "agents",
			Description:  "Zendesk Chat agents.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       zendeskChatAgentFields(),
		},
		{
			Name:         "chats",
			Description:  "Zendesk Chat conversations (incremental export).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"timestamp"},
			Fields:       zendeskChatChatFields(),
		},
		{
			Name:         "departments",
			Description:  "Zendesk Chat departments.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       zendeskChatDepartmentFields(),
		},
		{
			Name:         "shortcuts",
			Description:  "Zendesk Chat canned-message shortcuts.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       zendeskChatShortcutFields(),
		},
		{
			Name:         "triggers",
			Description:  "Zendesk Chat automation triggers.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       zendeskChatTriggerFields(),
		},
	}
}

func zendeskChatAgentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "display_name", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "role_id", Type: "integer"},
		{Name: "enabled", Type: "boolean"},
		{Name: "create_date", Type: "string"},
		{Name: "last_login", Type: "string"},
	}
}

func zendeskChatChatFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "timestamp", Type: "string"},
		{Name: "session", Type: "object"},
		{Name: "visitor", Type: "object"},
		{Name: "department_id", Type: "integer"},
		{Name: "rating", Type: "string"},
		{Name: "comment", Type: "string"},
		{Name: "duration", Type: "integer"},
	}
}

func zendeskChatDepartmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "enabled", Type: "boolean"},
		{Name: "members", Type: "array"},
		{Name: "settings", Type: "object"},
	}
}

func zendeskChatShortcutFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "options", Type: "string"},
		{Name: "message", Type: "string"},
		{Name: "tags", Type: "array"},
		{Name: "scope", Type: "string"},
	}
}

func zendeskChatTriggerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "enabled", Type: "boolean"},
		{Name: "definition", Type: "object"},
	}
}

func zendeskChatAgentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"display_name": item["display_name"],
		"first_name":   item["first_name"],
		"last_name":    item["last_name"],
		"email":        item["email"],
		"role_id":      item["role_id"],
		"enabled":      item["enabled"],
		"create_date":  item["create_date"],
		"last_login":   item["last_login"],
	}
}

func zendeskChatChatRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"type":          item["type"],
		"timestamp":     item["timestamp"],
		"session":       item["session"],
		"visitor":       item["visitor"],
		"department_id": item["department_id"],
		"rating":        item["rating"],
		"comment":       item["comment"],
		"duration":      item["duration"],
	}
}

func zendeskChatDepartmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"enabled":     item["enabled"],
		"members":     item["members"],
		"settings":    item["settings"],
	}
}

func zendeskChatShortcutRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      item["id"],
		"name":    item["name"],
		"options": item["options"],
		"message": item["message"],
		"tags":    item["tags"],
		"scope":   item["scope"],
	}
}

func zendeskChatTriggerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"enabled":     item["enabled"],
		"definition":  item["definition"],
	}
}
