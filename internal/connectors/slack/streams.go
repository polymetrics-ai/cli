package slack

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Slack Web API method path (relative
// to base_url), the JSON list key its records live under (members/channels/
// messages, NOT "data"), and the record mapper that flattens each object.
type streamEndpoint struct {
	// resource is the Slack Web API method path (e.g. "users.list").
	resource string
	// listKey is the top-level JSON key holding the record array
	// (e.g. "members" for users.list, "channels" for conversations.list).
	listKey string
	// mapRecord flattens a raw Slack object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// slackStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in slackStreams; the read path
// is fully data-driven from this table.
var slackStreamEndpoints = map[string]streamEndpoint{
	"users":            {resource: "users.list", listKey: "members", mapRecord: slackUserRecord},
	"channels":         {resource: "conversations.list", listKey: "channels", mapRecord: slackChannelRecord},
	"channel_messages": {resource: "conversations.history", listKey: "messages", mapRecord: slackMessageRecord},
}

// slackStreams returns the connector's published stream catalog. Slack only
// supports full_refresh sync (no reliable incremental cursor across the Web API
// list methods), so CursorFields is left empty.
func slackStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "users",
			Description: "Slack workspace members (users.list -> members[]).",
			PrimaryKey:  []string{"id"},
			Fields:      slackUserFields(),
		},
		{
			Name:        "channels",
			Description: "Slack conversations/channels (conversations.list -> channels[]).",
			PrimaryKey:  []string{"id"},
			Fields:      slackChannelFields(),
		},
		{
			Name:        "channel_messages",
			Description: "Messages within a channel (conversations.history -> messages[]). Requires config channel_id.",
			PrimaryKey:  []string{"ts"},
			Fields:      slackMessageFields(),
		},
	}
}

func slackUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "team_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "real_name", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "deleted", Type: "boolean"},
		{Name: "is_admin", Type: "boolean"},
		{Name: "is_bot", Type: "boolean"},
		{Name: "tz", Type: "string"},
		{Name: "updated", Type: "integer"},
	}
}

func slackChannelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "is_channel", Type: "boolean"},
		{Name: "is_group", Type: "boolean"},
		{Name: "is_private", Type: "boolean"},
		{Name: "is_archived", Type: "boolean"},
		{Name: "is_general", Type: "boolean"},
		{Name: "created", Type: "integer"},
		{Name: "creator", Type: "string"},
		{Name: "num_members", Type: "integer"},
		{Name: "topic", Type: "string"},
		{Name: "purpose", Type: "string"},
	}
}

func slackMessageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ts", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "subtype", Type: "string"},
		{Name: "user", Type: "string"},
		{Name: "text", Type: "string"},
		{Name: "thread_ts", Type: "string"},
		{Name: "reply_count", Type: "integer"},
		{Name: "team", Type: "string"},
	}
}

func slackUserRecord(item map[string]any) connectors.Record {
	profile, _ := item["profile"].(map[string]any)
	return connectors.Record{
		"id":           item["id"],
		"team_id":      item["team_id"],
		"name":         item["name"],
		"real_name":    item["real_name"],
		"display_name": profileField(profile, "display_name"),
		"email":        profileField(profile, "email"),
		"deleted":      item["deleted"],
		"is_admin":     item["is_admin"],
		"is_bot":       item["is_bot"],
		"tz":           item["tz"],
		"updated":      item["updated"],
	}
}

func slackChannelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"is_channel":  item["is_channel"],
		"is_group":    item["is_group"],
		"is_private":  item["is_private"],
		"is_archived": item["is_archived"],
		"is_general":  item["is_general"],
		"created":     item["created"],
		"creator":     item["creator"],
		"num_members": item["num_members"],
		"topic":       nestedValue(item, "topic", "value"),
		"purpose":     nestedValue(item, "purpose", "value"),
	}
}

func slackMessageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ts":          item["ts"],
		"type":        item["type"],
		"subtype":     item["subtype"],
		"user":        item["user"],
		"text":        item["text"],
		"thread_ts":   item["thread_ts"],
		"reply_count": item["reply_count"],
		"team":        item["team"],
	}
}

// profileField safely reads a string-ish field from a user's profile object.
func profileField(profile map[string]any, key string) any {
	if profile == nil {
		return nil
	}
	return profile[key]
}

// nestedValue reads object[outer][inner], used for Slack's topic/purpose objects
// which carry their text under a "value" sub-key.
func nestedValue(item map[string]any, outer, inner string) any {
	obj, ok := item[outer].(map[string]any)
	if !ok {
		return nil
	}
	return obj[inner]
}
