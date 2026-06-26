package drift

import "polymetrics.ai/internal/connectors"

// paginationKind enumerates the three pagination shapes the Drift API uses
// across the connector's streams.
type paginationKind int

const (
	// paginationNone is a single-shot list with no paging (users/list).
	paginationNone paginationKind = iota
	// paginationCursor follows a body cursor at pagination.next, gated by
	// pagination.more, supplied back as the "next" query param (conversations).
	paginationCursor
	// paginationNextURL follows an absolute "next" URL embedded in the body
	// (accounts: data.next).
	paginationNextURL
)

// streamEndpoint maps a stream name to its Drift API resource path (relative to
// base_url), the JSON path where its record array lives, the pagination shape,
// and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Drift list endpoint path (e.g. "users/list").
	resource string
	// recordsPath is the dotted JSON path to the record array (e.g. "data" or
	// "data.accounts").
	recordsPath string
	// nextPath is the dotted JSON path to the next-page token/URL, when the
	// stream paginates.
	nextPath string
	// morePath, when set (cursor pagination), gates whether to follow nextPath.
	morePath string
	// pagination is the pagination shape for the stream.
	pagination paginationKind
	// mapRecord flattens a raw Drift object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// driftStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in driftStreams; the read path
// is fully data-driven from this table.
var driftStreamEndpoints = map[string]streamEndpoint{
	"users": {
		resource:    "users/list",
		recordsPath: "data",
		pagination:  paginationNone,
		mapRecord:   driftUserRecord,
	},
	"accounts": {
		resource:    "accounts",
		recordsPath: "data.accounts",
		nextPath:    "data.next",
		pagination:  paginationNextURL,
		mapRecord:   driftAccountRecord,
	},
	"conversations": {
		resource:    "conversations/list",
		recordsPath: "data",
		nextPath:    "pagination.next",
		morePath:    "pagination.more",
		pagination:  paginationCursor,
		mapRecord:   driftConversationRecord,
	},
	"contacts": {
		resource:    "contacts",
		recordsPath: "data",
		pagination:  paginationNone,
		mapRecord:   driftContactRecord,
	},
}

// driftStreams returns the connector's published stream catalog.
func driftStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "users",
			Description:  "Drift users (agents) in the organization.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       driftUserFields(),
		},
		{
			Name:         "accounts",
			Description:  "Drift accounts (companies).",
			PrimaryKey:   []string{"account_id"},
			CursorFields: []string{"updateDateTime"},
			Fields:       driftAccountFields(),
		},
		{
			Name:         "conversations",
			Description:  "Drift conversations.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       driftConversationFields(),
		},
		{
			Name:         "contacts",
			Description:  "Drift contacts (people) looked up by email.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       driftContactFields(),
		},
	}
}

func driftUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "orgId", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "alias", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "locale", Type: "string"},
		{Name: "availability", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "timeZone", Type: "string"},
		{Name: "avatarUrl", Type: "string"},
		{Name: "verified", Type: "boolean"},
		{Name: "bot", Type: "boolean"},
		{Name: "createdAt", Type: "integer"},
		{Name: "updatedAt", Type: "integer"},
	}
}

func driftAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "account_id", Type: "string"},
		{Name: "ownerId", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "domain", Type: "string"},
		{Name: "deleted", Type: "boolean"},
		{Name: "targeted", Type: "boolean"},
		{Name: "createDateTime", Type: "integer"},
		{Name: "updateDateTime", Type: "integer"},
		{Name: "customProperties", Type: "array"},
	}
}

func driftConversationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "orgId", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "contactId", Type: "integer"},
		{Name: "inboxId", Type: "integer"},
		{Name: "participants", Type: "array"},
		{Name: "conversationTags", Type: "array"},
		{Name: "relatedPlaybookId", Type: "integer"},
		{Name: "createdAt", Type: "integer"},
		{Name: "updatedAt", Type: "integer"},
	}
}

func driftContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "createdAt", Type: "integer"},
		{Name: "updatedAt", Type: "integer"},
		{Name: "attributes", Type: "object"},
	}
}

func driftUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"orgId":        item["orgId"],
		"name":         item["name"],
		"alias":        item["alias"],
		"email":        item["email"],
		"phone":        item["phone"],
		"locale":       item["locale"],
		"availability": item["availability"],
		"role":         item["role"],
		"timeZone":     item["timeZone"],
		"avatarUrl":    item["avatarUrl"],
		"verified":     item["verified"],
		"bot":          item["bot"],
		"createdAt":    item["createdAt"],
		"updatedAt":    item["updatedAt"],
	}
}

func driftAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"account_id":       item["accountId"],
		"ownerId":          item["ownerId"],
		"name":             item["name"],
		"domain":           item["domain"],
		"deleted":          item["deleted"],
		"targeted":         item["targeted"],
		"createDateTime":   item["createDateTime"],
		"updateDateTime":   item["updateDateTime"],
		"customProperties": item["customProperties"],
	}
}

func driftConversationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"orgId":             item["orgId"],
		"status":            item["status"],
		"contactId":         item["contactId"],
		"inboxId":           item["inboxId"],
		"participants":      item["participants"],
		"conversationTags":  item["conversationTags"],
		"relatedPlaybookId": item["relatedPlaybookId"],
		"createdAt":         item["createdAt"],
		"updatedAt":         item["updatedAt"],
	}
}

func driftContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"createdAt":  item["createdAt"],
		"updatedAt":  item["updatedAt"],
		"attributes": item["attributes"],
	}
}
