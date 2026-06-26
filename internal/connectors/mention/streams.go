package mention

import "polymetrics.ai/internal/connectors"

// scope describes how a stream's API path is built relative to the resolved
// account (and, for mention/alert_tag streams, an alert).
type scope int

const (
	// scopeAccountMe reads /accounts/me (the authenticated account).
	scopeAccountMe scope = iota
	// scopeAccount reads /accounts/{account_id}.
	scopeAccount
	// scopeAccountAlerts reads /accounts/{account_id}/alerts.
	scopeAccountAlerts
	// scopeAlertMentions reads /accounts/{account_id}/alerts/{alert_id}/mentions.
	scopeAlertMentions
	// scopeAlertTags reads /accounts/{account_id}/alerts/{alert_id}/tags.
	scopeAlertTags
)

// streamEndpoint maps a stream name to its scope, the JSON field path where the
// records array lives, whether it is paginated via _links.more.params.cursor,
// and the record mapper that flattens its objects.
type streamEndpoint struct {
	scope     scope
	fieldPath string
	paginated bool
	mapRecord func(map[string]any) connectors.Record
}

// mentionStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in mentionStreams.
var mentionStreamEndpoints = map[string]streamEndpoint{
	"account_me": {scope: scopeAccountMe, fieldPath: "account", paginated: false, mapRecord: mentionAccountRecord},
	"account":    {scope: scopeAccount, fieldPath: "account", paginated: false, mapRecord: mentionAccountRecord},
	"alert":      {scope: scopeAccountAlerts, fieldPath: "alerts", paginated: true, mapRecord: mentionAlertRecord},
	"mention":    {scope: scopeAlertMentions, fieldPath: "mentions", paginated: true, mapRecord: mentionMentionRecord},
	"alert_tag":  {scope: scopeAlertTags, fieldPath: "tags", paginated: false, mapRecord: mentionTagRecord},
}

// mentionStreams returns the connector's published stream catalog.
func mentionStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "account_me",
			Description: "The authenticated Mention account (/accounts/me).",
			PrimaryKey:  []string{"id"},
			Fields:      mentionAccountFields(),
		},
		{
			Name:        "account",
			Description: "A Mention account by id.",
			PrimaryKey:  []string{"id"},
			Fields:      mentionAccountFields(),
		},
		{
			Name:        "alert",
			Description: "Mention alerts (monitored queries) for the account.",
			PrimaryKey:  []string{"id"},
			Fields:      mentionAlertFields(),
		},
		{
			Name:        "mention",
			Description: "Individual mentions matched by an alert.",
			PrimaryKey:  []string{"id"},
			Fields:      mentionMentionFields(),
		},
		{
			Name:        "alert_tag",
			Description: "Tags configured on an alert.",
			PrimaryKey:  []string{"id"},
			Fields:      mentionTagFields(),
		},
	}
}

func mentionAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "permission", Type: "string"},
	}
}

func mentionAlertFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "query", Type: "object"},
		{Name: "languages", Type: "array"},
		{Name: "countries", Type: "array"},
		{Name: "sources", Type: "array"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func mentionMentionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "source_name", Type: "string"},
		{Name: "source_type", Type: "string"},
		{Name: "tone", Type: "number"},
		{Name: "favorite", Type: "boolean"},
		{Name: "published_at", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func mentionTagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "color", Type: "string"},
	}
}

func mentionAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"language":   item["language"],
		"timezone":   item["timezone"],
		"created_at": item["created_at"],
		"permission": item["permission"],
	}
}

func mentionAlertRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"query":       item["query"],
		"languages":   item["languages"],
		"countries":   item["countries"],
		"sources":     item["sources"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func mentionMentionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"title":        item["title"],
		"description":  item["description"],
		"url":          item["url"],
		"language":     item["language"],
		"source_name":  item["source_name"],
		"source_type":  item["source_type"],
		"tone":         item["tone"],
		"favorite":     item["favorite"],
		"published_at": item["published_at"],
		"created_at":   item["created_at"],
	}
}

func mentionTagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":    item["id"],
		"name":  item["name"],
		"color": item["color"],
	}
}
