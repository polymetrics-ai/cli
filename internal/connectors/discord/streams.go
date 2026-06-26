package discord

import "polymetrics.ai/internal/connectors"

// pagination describes how a Discord stream is paginated.
type pagination int

const (
	// pageNone reads a single response (top-level array or object) with no
	// cursor, e.g. /guilds/{id}/channels and /guilds/{id}/roles.
	pageNone pagination = iota
	// pageAfterID drives Discord's snowflake `after`=highest-id cursor used by
	// /guilds/{id}/members, looping until a short page is returned.
	pageAfterID
	// pageSingle reads a single object resource (e.g. /guilds/{id}) and emits it
	// as one record.
	pageSingle
)

// streamEndpoint maps a stream name to the Discord API resource path (relative to
// base_url, with {guild_id} substituted) it reads from, the pagination style, and
// the record mapper that flattens its objects.
type streamEndpoint struct {
	// pathTemplate is the resource path with a literal "{guild_id}" placeholder.
	pathTemplate string
	pagination   pagination
	mapRecord    func(map[string]any) connectors.Record
}

// discordStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in discordStreams; the read path
// is fully data-driven from this table.
var discordStreamEndpoints = map[string]streamEndpoint{
	"guilds":   {pathTemplate: "guilds/{guild_id}", pagination: pageSingle, mapRecord: discordGuildRecord},
	"channels": {pathTemplate: "guilds/{guild_id}/channels", pagination: pageNone, mapRecord: discordChannelRecord},
	"roles":    {pathTemplate: "guilds/{guild_id}/roles", pagination: pageNone, mapRecord: discordRoleRecord},
	"members":  {pathTemplate: "guilds/{guild_id}/members", pagination: pageAfterID, mapRecord: discordMemberRecord},
}

// discordStreams returns the connector's published stream catalog. Discord is
// full-refresh only (the upstream Airbyte source supports full_refresh), so no
// cursor fields are advertised; each object exposes a snowflake string id.
func discordStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "guilds",
			Description: "The configured Discord guild (server) detail.",
			PrimaryKey:  []string{"id"},
			Fields:      discordGuildFields(),
		},
		{
			Name:        "channels",
			Description: "Channels in the configured guild.",
			PrimaryKey:  []string{"id"},
			Fields:      discordChannelFields(),
		},
		{
			Name:        "roles",
			Description: "Roles defined in the configured guild.",
			PrimaryKey:  []string{"id"},
			Fields:      discordRoleFields(),
		},
		{
			Name:        "members",
			Description: "Members of the configured guild.",
			PrimaryKey:  []string{"id"},
			Fields:      discordMemberFields(),
		},
	}
}

func discordGuildFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "owner_id", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "icon", Type: "string"},
		{Name: "approximate_member_count", Type: "integer"},
		{Name: "approximate_presence_count", Type: "integer"},
		{Name: "premium_tier", Type: "integer"},
		{Name: "preferred_locale", Type: "string"},
	}
}

func discordChannelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "guild_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "integer"},
		{Name: "position", Type: "integer"},
		{Name: "parent_id", Type: "string"},
		{Name: "topic", Type: "string"},
		{Name: "nsfw", Type: "boolean"},
	}
}

func discordRoleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "color", Type: "integer"},
		{Name: "position", Type: "integer"},
		{Name: "permissions", Type: "string"},
		{Name: "hoist", Type: "boolean"},
		{Name: "managed", Type: "boolean"},
		{Name: "mentionable", Type: "boolean"},
	}
}

func discordMemberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "username", Type: "string"},
		{Name: "nick", Type: "string"},
		{Name: "joined_at", Type: "string"},
		{Name: "premium_since", Type: "string"},
		{Name: "roles", Type: "array"},
		{Name: "deaf", Type: "boolean"},
		{Name: "mute", Type: "boolean"},
		{Name: "pending", Type: "boolean"},
	}
}

func discordGuildRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                         item["id"],
		"name":                       item["name"],
		"owner_id":                   item["owner_id"],
		"description":                item["description"],
		"icon":                       item["icon"],
		"approximate_member_count":   item["approximate_member_count"],
		"approximate_presence_count": item["approximate_presence_count"],
		"premium_tier":               item["premium_tier"],
		"preferred_locale":           item["preferred_locale"],
	}
}

func discordChannelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"guild_id":  item["guild_id"],
		"name":      item["name"],
		"type":      item["type"],
		"position":  item["position"],
		"parent_id": item["parent_id"],
		"topic":     item["topic"],
		"nsfw":      item["nsfw"],
	}
}

func discordRoleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"color":       item["color"],
		"position":    item["position"],
		"permissions": item["permissions"],
		"hoist":       item["hoist"],
		"managed":     item["managed"],
		"mentionable": item["mentionable"],
	}
}

// discordMemberRecord flattens a guild member object. The member resource nests
// the account under "user"; the connector's primary key uses that user id (a
// member is unique per user within a guild), surfaced as both id and user_id.
func discordMemberRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"nick":          item["nick"],
		"joined_at":     item["joined_at"],
		"premium_since": item["premium_since"],
		"roles":         item["roles"],
		"deaf":          item["deaf"],
		"mute":          item["mute"],
		"pending":       item["pending"],
	}
	var userID any
	var username any
	if user, ok := item["user"].(map[string]any); ok {
		userID = user["id"]
		username = user["username"]
		rec["global_name"] = user["global_name"]
		rec["bot"] = user["bot"]
	}
	rec["id"] = userID
	rec["user_id"] = userID
	rec["username"] = username
	return rec
}

// memberAfterID returns the snowflake id used as the `after` cursor for the next
// members page: the user id of the just-emitted member.
func memberAfterID(item map[string]any) string {
	if user, ok := item["user"].(map[string]any); ok {
		return stringField(user, "id")
	}
	return ""
}
