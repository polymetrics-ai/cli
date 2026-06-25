package monday

import "polymetrics/internal/connectors"

// streamSpec describes how one monday.com stream is read from the GraphQL API.
//
// monday.com exposes data through a single GraphQL endpoint, so each stream is
// defined by the GraphQL root field it queries (root), the comma-separated set
// of sub-fields to select (selection), the JSON path under "data" where the
// records array lives after a successful response, and a mapper that flattens a
// raw object into a connectors.Record.
//
// The boards/users/teams/tags streams use simple page-number pagination
// (limit + page arguments). The items stream is special: it is read through the
// boards { items_page } envelope and continued with the cursor-based
// next_items_page field, so it is routed separately in monday.go.
type streamSpec struct {
	// root is the GraphQL root field (e.g. "boards", "users").
	root string
	// selection is the GraphQL field selection set for each object, without the
	// surrounding braces (e.g. "id name state board_kind").
	selection string
	// recordsPath is the dotted path under the response's "data" object where the
	// records array lives (e.g. "boards").
	recordsPath string
	// mapRecord flattens a raw GraphQL object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// pageStreamSpecs is the routing table for page-number paginated streams. Adding
// a page-based stream means adding one entry here plus a Stream definition in
// mondayStreams. The items stream is handled separately (cursor pagination).
var pageStreamSpecs = map[string]streamSpec{
	"boards": {
		root:        "boards",
		selection:   "id name state board_kind description type updated_at workspace_id",
		recordsPath: "boards",
		mapRecord:   boardRecord,
	},
	"users": {
		root:        "users",
		selection:   "id name email enabled is_admin is_guest is_pending created_at",
		recordsPath: "users",
		mapRecord:   userRecord,
	},
	"teams": {
		root:        "teams",
		selection:   "id name picture_url",
		recordsPath: "teams",
		mapRecord:   teamRecord,
	},
	"tags": {
		root:        "tags",
		selection:   "id name color",
		recordsPath: "tags",
		mapRecord:   tagRecord,
	},
}

// itemSelection is the GraphQL field selection for an item, shared by the
// boards { items_page } and next_items_page queries.
const itemSelection = "id name state created_at updated_at group { id title } board { id name }"

// mondayStreams returns the connector's published stream catalog.
func mondayStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "boards",
			Description:  "monday.com boards in the account.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       boardFields(),
		},
		{
			Name:         "items",
			Description:  "monday.com items (rows) across the account's boards.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       itemFields(),
		},
		{
			Name:        "users",
			Description: "monday.com users in the account.",
			PrimaryKey:  []string{"id"},
			Fields:      userFields(),
		},
		{
			Name:        "teams",
			Description: "monday.com teams in the account.",
			PrimaryKey:  []string{"id"},
			Fields:      teamFields(),
		},
		{
			Name:        "tags",
			Description: "monday.com public tags in the account.",
			PrimaryKey:  []string{"id"},
			Fields:      tagFields(),
		},
	}
}

func boardFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "board_kind", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "workspace_id", Type: "string"},
	}
}

func itemFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "group_id", Type: "string"},
		{Name: "group_title", Type: "string"},
		{Name: "board_id", Type: "string"},
		{Name: "board_name", Type: "string"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "enabled", Type: "boolean"},
		{Name: "is_admin", Type: "boolean"},
		{Name: "is_guest", Type: "boolean"},
		{Name: "is_pending", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func teamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "picture_url", Type: "string"},
	}
}

func tagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "color", Type: "string"},
	}
}

func boardRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           stringField(item, "id"),
		"name":         item["name"],
		"state":        item["state"],
		"board_kind":   item["board_kind"],
		"description":  item["description"],
		"type":         item["type"],
		"updated_at":   item["updated_at"],
		"workspace_id": stringField(item, "workspace_id"),
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         stringField(item, "id"),
		"name":       item["name"],
		"email":      item["email"],
		"enabled":    item["enabled"],
		"is_admin":   item["is_admin"],
		"is_guest":   item["is_guest"],
		"is_pending": item["is_pending"],
		"created_at": item["created_at"],
	}
}

func teamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          stringField(item, "id"),
		"name":        item["name"],
		"picture_url": item["picture_url"],
	}
}

func tagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":    stringField(item, "id"),
		"name":  item["name"],
		"color": item["color"],
	}
}

// itemRecord flattens a monday item object, hoisting the nested group and board
// objects into flat group_*/board_* columns.
func itemRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":         stringField(item, "id"),
		"name":       item["name"],
		"state":      item["state"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
	if group, ok := item["group"].(map[string]any); ok {
		rec["group_id"] = stringField(group, "id")
		rec["group_title"] = group["title"]
	}
	if board, ok := item["board"].(map[string]any); ok {
		rec["board_id"] = stringField(board, "id")
		rec["board_name"] = board["name"]
	}
	return rec
}
