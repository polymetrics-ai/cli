package miro

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Miro REST API v2 resource path it
// reads from, whether that path is scoped to a single board, and the record
// mapper that flattens its objects.
type streamEndpoint struct {
	// pathTemplate is the API path relative to base_url. When boardScoped is
	// true it contains a single "%s" placeholder for the board id.
	pathTemplate string
	// boardScoped marks streams nested under /v2/boards/{board_id}/...; these
	// require a board_id config value.
	boardScoped bool
	// mapRecord flattens a raw Miro object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// miroStreamEndpoints is the per-stream routing table. The read path is fully
// data-driven from this table: adding a stream means adding one entry here plus
// a Stream definition in miroStreams.
var miroStreamEndpoints = map[string]streamEndpoint{
	"boards":           {pathTemplate: "/v2/boards", boardScoped: false, mapRecord: miroBoardRecord},
	"board_users":      {pathTemplate: "/v2/boards/%s/members", boardScoped: true, mapRecord: miroBoardUserRecord},
	"board_items":      {pathTemplate: "/v2/boards/%s/items", boardScoped: true, mapRecord: miroBoardItemRecord},
	"board_tags":       {pathTemplate: "/v2/boards/%s/tags", boardScoped: true, mapRecord: miroBoardTagRecord},
	"board_connectors": {pathTemplate: "/v2/boards/%s/connectors", boardScoped: true, mapRecord: miroBoardConnectorRecord},
}

// miroStreams returns the connector's published stream catalog. Miro objects all
// carry a string id, so the primary key is ["id"] across the board. The API only
// supports full_refresh sync, so there are no incremental cursor fields.
func miroStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "boards",
			Description: "Miro boards accessible to the authenticated user.",
			PrimaryKey:  []string{"id"},
			Fields:      miroBoardFields(),
		},
		{
			Name:        "board_users",
			Description: "Members (users) of a Miro board.",
			PrimaryKey:  []string{"id"},
			Fields:      miroBoardUserFields(),
		},
		{
			Name:        "board_items",
			Description: "Items (widgets) on a Miro board.",
			PrimaryKey:  []string{"id"},
			Fields:      miroBoardItemFields(),
		},
		{
			Name:        "board_tags",
			Description: "Tags defined on a Miro board.",
			PrimaryKey:  []string{"id"},
			Fields:      miroBoardTagFields(),
		},
		{
			Name:        "board_connectors",
			Description: "Connector lines between items on a Miro board.",
			PrimaryKey:  []string{"id"},
			Fields:      miroBoardConnectorFields(),
		},
	}
}

func miroBoardFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "view_link", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
		{Name: "owner_id", Type: "string"},
		{Name: "team_id", Type: "string"},
	}
}

func miroBoardUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "board_id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "role", Type: "string"},
	}
}

func miroBoardItemFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "board_id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
	}
}

func miroBoardTagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "board_id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "fill_color", Type: "string"},
	}
}

func miroBoardConnectorFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "board_id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "shape", Type: "string"},
	}
}

func miroBoardRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":          item["id"],
		"type":        item["type"],
		"name":        item["name"],
		"description": item["description"],
		"view_link":   item["viewLink"],
		"created_at":  item["createdAt"],
		"modified_at": item["modifiedAt"],
	}
	if owner, ok := item["owner"].(map[string]any); ok {
		rec["owner_id"] = owner["id"]
	}
	if team, ok := item["team"].(map[string]any); ok {
		rec["team_id"] = team["id"]
	}
	return rec
}

func miroBoardUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"board_id": item["board_id"],
		"type":     item["type"],
		"name":     item["name"],
		"role":     item["role"],
	}
}

func miroBoardItemRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"board_id":    item["board_id"],
		"type":        item["type"],
		"created_at":  item["createdAt"],
		"modified_at": item["modifiedAt"],
	}
}

func miroBoardTagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"board_id":   item["board_id"],
		"type":       item["type"],
		"title":      item["title"],
		"fill_color": item["fillColor"],
	}
}

func miroBoardConnectorRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"board_id": item["board_id"],
		"type":     item["type"],
		"shape":    item["shape"],
	}
}
