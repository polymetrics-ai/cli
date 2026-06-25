package trello

import "polymetrics.ai/internal/connectors"

// scope describes how a Trello stream is fetched relative to the API base URL.
//   - boardScoped streams are read once per board at /boards/{id}/<resource>.
//   - the boards stream itself is read either from configured board IDs
//     (/boards/{id}) or from /members/me/boards.
type scope int

const (
	scopeBoards scope = iota // the boards stream
	scopeBoard               // a per-board sub-resource (lists, cards, ...)
)

// streamEndpoint maps a stream name to its API resource and the record mapper
// that flattens its objects. The read path is fully data-driven from this table.
type streamEndpoint struct {
	// resource is the per-board sub-resource path segment (e.g. "cards").
	// Unused for the boards stream.
	resource string
	scope    scope
	// paginated is true for resources that support id-cursor `before` paging
	// (cards and actions). Lists/checklists return the full set in one call.
	paginated bool
	mapRecord func(map[string]any) connectors.Record
}

// trelloStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in trelloStreams.
var trelloStreamEndpoints = map[string]streamEndpoint{
	"boards":     {scope: scopeBoards, mapRecord: trelloBoardRecord},
	"lists":      {resource: "lists", scope: scopeBoard, mapRecord: trelloListRecord},
	"cards":      {resource: "cards", scope: scopeBoard, paginated: true, mapRecord: trelloCardRecord},
	"checklists": {resource: "checklists", scope: scopeBoard, mapRecord: trelloChecklistRecord},
	"actions":    {resource: "actions", scope: scopeBoard, paginated: true, mapRecord: trelloActionRecord},
}

// trelloStreams returns the connector's published stream catalog. Every Trello
// object exposes a string id, so the primary key is ["id"] across the board.
func trelloStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "boards",
			Description: "Trello boards the authenticated member can access (or the configured board IDs).",
			PrimaryKey:  []string{"id"},
			Fields:      trelloBoardFields(),
		},
		{
			Name:        "lists",
			Description: "Lists (columns) belonging to each board.",
			PrimaryKey:  []string{"id"},
			Fields:      trelloListFields(),
		},
		{
			Name:         "cards",
			Description:  "Cards belonging to each board.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"dateLastActivity"},
			Fields:       trelloCardFields(),
		},
		{
			Name:        "checklists",
			Description: "Checklists belonging to each board.",
			PrimaryKey:  []string{"id"},
			Fields:      trelloChecklistFields(),
		},
		{
			Name:         "actions",
			Description:  "Board actions (activity feed). Incremental on date.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date"},
			Fields:       trelloActionFields(),
		},
	}
}

func trelloBoardFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "desc", Type: "string"},
		{Name: "closed", Type: "boolean"},
		{Name: "idOrganization", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "shortUrl", Type: "string"},
		{Name: "dateLastActivity", Type: "string"},
	}
}

func trelloListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "closed", Type: "boolean"},
		{Name: "idBoard", Type: "string"},
		{Name: "pos", Type: "number"},
		{Name: "subscribed", Type: "boolean"},
	}
}

func trelloCardFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "desc", Type: "string"},
		{Name: "closed", Type: "boolean"},
		{Name: "idBoard", Type: "string"},
		{Name: "idList", Type: "string"},
		{Name: "due", Type: "string"},
		{Name: "dueComplete", Type: "boolean"},
		{Name: "url", Type: "string"},
		{Name: "shortUrl", Type: "string"},
		{Name: "dateLastActivity", Type: "string"},
	}
}

func trelloChecklistFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "idBoard", Type: "string"},
		{Name: "idCard", Type: "string"},
		{Name: "pos", Type: "number"},
	}
}

func trelloActionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "idMemberCreator", Type: "string"},
		{Name: "idBoard", Type: "string"},
	}
}

func trelloBoardRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"desc":             item["desc"],
		"closed":           item["closed"],
		"idOrganization":   item["idOrganization"],
		"url":              item["url"],
		"shortUrl":         item["shortUrl"],
		"dateLastActivity": item["dateLastActivity"],
	}
}

func trelloListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"closed":     item["closed"],
		"idBoard":    item["idBoard"],
		"pos":        item["pos"],
		"subscribed": item["subscribed"],
	}
}

func trelloCardRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"desc":             item["desc"],
		"closed":           item["closed"],
		"idBoard":          item["idBoard"],
		"idList":           item["idList"],
		"due":              item["due"],
		"dueComplete":      item["dueComplete"],
		"url":              item["url"],
		"shortUrl":         item["shortUrl"],
		"dateLastActivity": item["dateLastActivity"],
	}
}

func trelloChecklistRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      item["id"],
		"name":    item["name"],
		"idBoard": item["idBoard"],
		"idCard":  item["idCard"],
		"pos":     item["pos"],
	}
}

func trelloActionRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":              item["id"],
		"type":            item["type"],
		"date":            item["date"],
		"idMemberCreator": item["idMemberCreator"],
	}
	// Trello nests the board id under data.board.id on actions; surface it flat.
	if data, ok := item["data"].(map[string]any); ok {
		if board, ok := data["board"].(map[string]any); ok {
			rec["idBoard"] = board["id"]
		}
	}
	if rec["idBoard"] == nil {
		rec["idBoard"] = item["idBoard"]
	}
	return rec
}
