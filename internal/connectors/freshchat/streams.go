package freshchat

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Freshchat API resource path (relative
// to base_url) it reads from, the JSON wrapper key that holds the list array,
// and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Freshchat list endpoint path segment (e.g. "agents").
	resource string
	// wrapper is the response key holding the array of records (e.g. "agents").
	wrapper string
	// mapRecord flattens a raw Freshchat object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// freshchatStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in freshchatStreams; the read
// path is fully data-driven from this table.
var freshchatStreamEndpoints = map[string]streamEndpoint{
	"agents":   {resource: "agents", wrapper: "agents", mapRecord: freshchatAgentRecord},
	"users":    {resource: "users", wrapper: "users", mapRecord: freshchatUserRecord},
	"groups":   {resource: "groups", wrapper: "groups", mapRecord: freshchatGroupRecord},
	"channels": {resource: "channels", wrapper: "channels", mapRecord: freshchatChannelRecord},
	"roles":    {resource: "roles", wrapper: "roles", mapRecord: freshchatRoleRecord},
}

// freshchatStreams returns the connector's published stream catalog. Freshchat
// resources are keyed by a string id; agents/users/channels expose mutation
// timestamps usable as incremental cursors.
func freshchatStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "agents",
			Description:  "Freshchat agents (support staff).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_time"},
			Fields:       freshchatAgentFields(),
		},
		{
			Name:         "users",
			Description:  "Freshchat end users (contacts).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_time"},
			Fields:       freshchatUserFields(),
		},
		{
			Name:        "groups",
			Description: "Freshchat agent groups.",
			PrimaryKey:  []string{"id"},
			Fields:      freshchatGroupFields(),
		},
		{
			Name:         "channels",
			Description:  "Freshchat conversation channels (topics).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_time"},
			Fields:       freshchatChannelFields(),
		},
		{
			Name:        "roles",
			Description: "Freshchat agent roles.",
			PrimaryKey:  []string{"id"},
			Fields:      freshchatRoleFields(),
		},
	}
}

func freshchatAgentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "avatar", Type: "object"},
		{Name: "biography", Type: "string"},
		{Name: "is_deactivated", Type: "boolean"},
		{Name: "is_deleted", Type: "boolean"},
		{Name: "social_profiles", Type: "array"},
		{Name: "groups", Type: "array"},
		{Name: "role_id", Type: "string"},
		{Name: "created_time", Type: "string"},
		{Name: "updated_time", Type: "string"},
	}
}

func freshchatUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "avatar", Type: "object"},
		{Name: "properties", Type: "array"},
		{Name: "reference_id", Type: "string"},
		{Name: "restore_id", Type: "string"},
		{Name: "created_time", Type: "string"},
		{Name: "updated_time", Type: "string"},
	}
}

func freshchatGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "routing_type", Type: "string"},
		{Name: "created_time", Type: "string"},
		{Name: "updated_time", Type: "string"},
	}
}

func freshchatChannelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "icon", Type: "object"},
		{Name: "tags", Type: "array"},
		{Name: "enabled", Type: "boolean"},
		{Name: "public", Type: "boolean"},
		{Name: "locale", Type: "string"},
		{Name: "welcome_message", Type: "object"},
		{Name: "created_time", Type: "string"},
		{Name: "updated_time", Type: "string"},
	}
}

func freshchatRoleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "role", Type: "string"},
	}
}

func freshchatAgentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"email":           item["email"],
		"first_name":      item["first_name"],
		"last_name":       item["last_name"],
		"avatar":          item["avatar"],
		"biography":       item["biography"],
		"is_deactivated":  item["is_deactivated"],
		"is_deleted":      item["is_deleted"],
		"social_profiles": item["social_profiles"],
		"groups":          item["groups"],
		"role_id":         item["role_id"],
		"created_time":    item["created_time"],
		"updated_time":    item["updated_time"],
	}
}

func freshchatUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"email":        item["email"],
		"first_name":   item["first_name"],
		"last_name":    item["last_name"],
		"phone":        item["phone"],
		"avatar":       item["avatar"],
		"properties":   item["properties"],
		"reference_id": item["reference_id"],
		"restore_id":   item["restore_id"],
		"created_time": item["created_time"],
		"updated_time": item["updated_time"],
	}
}

func freshchatGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"description":  item["description"],
		"routing_type": item["routing_type"],
		"created_time": item["created_time"],
		"updated_time": item["updated_time"],
	}
}

func freshchatChannelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            item["name"],
		"icon":            item["icon"],
		"tags":            item["tags"],
		"enabled":         item["enabled"],
		"public":          item["public"],
		"locale":          item["locale"],
		"welcome_message": item["welcome_message"],
		"created_time":    item["created_time"],
		"updated_time":    item["updated_time"],
	}
}

func freshchatRoleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"role":        item["role"],
	}
}
