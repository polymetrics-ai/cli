package klausapi

import "polymetrics.ai/internal/connectors"

// scope describes how a stream's request path is constructed. Klaus paths are
// scoped by account (and, for most streams, workspace) rather than by query
// params, so the path is templated from config at read time.
type scope int

const (
	// scopeAccount paths look like /account/{account}/<resource>.
	scopeAccount scope = iota
	// scopeWorkspace paths look like /account/{account}/workspace/{workspace}/<resource>.
	scopeWorkspace
)

// streamDef holds the per-stream routing: the path scope, the trailing resource
// segment, the JSON path to the records array (DpathExtractor field_path), the
// record mapper, and whether the stream is incremental via a date window.
type streamDef struct {
	scope       scope
	resource    string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
	// windowed marks streams (reviews) that page by fromDate/toDate date windows.
	windowed bool
}

// klausStreamDefs is the routing table. Adding a stream means adding one entry
// here plus a Stream definition in klausStreams.
var klausStreamDefs = map[string]streamDef{
	"users": {
		scope:       scopeAccount,
		resource:    "users",
		recordsPath: "users",
		mapRecord:   klausUserRecord,
	},
	"categories": {
		scope:       scopeWorkspace,
		resource:    "categories",
		recordsPath: "categories",
		mapRecord:   klausCategoryRecord,
	},
	"reviews": {
		scope:       scopeWorkspace,
		resource:    "reviews",
		recordsPath: "conversations",
		mapRecord:   klausReviewRecord,
		windowed:    true,
	},
}

// klausStreams returns the published catalog.
func klausStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "users",
			Description: "Klaus / Zendesk QA users in the account.",
			PrimaryKey:  []string{"id"},
			Fields:      klausUserFields(),
		},
		{
			Name:        "categories",
			Description: "Rating categories configured in the workspace.",
			PrimaryKey:  []string{"id"},
			Fields:      klausCategoryFields(),
		},
		{
			Name:         "reviews",
			Description:  "Review conversations in the workspace.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"lastUpdatedISO"},
			Fields:       klausReviewFields(),
		},
	}
}

func klausUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
	}
}

func klausCategoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "groupId", Type: "string"},
		{Name: "groupName", Type: "string"},
		{Name: "groupPosition", Type: "integer"},
		{Name: "position", Type: "integer"},
		{Name: "maxRating", Type: "integer"},
		{Name: "weight", Type: "number"},
		{Name: "critical", Type: "boolean"},
		{Name: "archived", Type: "boolean"},
	}
}

func klausReviewFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "externalId", Type: "string"},
		{Name: "externalUrl", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "sourceType", Type: "string"},
		{Name: "workspaceId", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "createdAtISO", Type: "string"},
		{Name: "lastUpdatedISO", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "reviews", Type: "array"},
		{Name: "comments", Type: "array"},
	}
}

func klausUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":    item["id"],
		"name":  item["name"],
		"email": item["email"],
	}
}

func klausCategoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"description":   item["description"],
		"groupId":       item["groupId"],
		"groupName":     item["groupName"],
		"groupPosition": item["groupPosition"],
		"position":      item["position"],
		"maxRating":     item["maxRating"],
		"weight":        item["weight"],
		"critical":      item["critical"],
		"archived":      item["archived"],
		"rootCauses":    item["rootCauses"],
		"scorecards":    item["scorecards"],
	}
}

func klausReviewRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"externalId":     item["externalId"],
		"externalUrl":    item["externalUrl"],
		"url":            item["url"],
		"sourceType":     item["sourceType"],
		"workspaceId":    item["workspaceId"],
		"createdAt":      item["createdAt"],
		"createdAtISO":   item["createdAtISO"],
		"lastUpdatedISO": item["lastUpdatedISO"],
		"updated_at":     item["updated_at"],
		"reviews":        item["reviews"],
		"comments":       item["comments"],
	}
}
