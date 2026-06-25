package clickupapi

import "polymetrics.ai/internal/connectors"

// streamShape describes how a ClickUp stream is read: the records envelope key,
// whether it paginates, and the record mapper that flattens its objects. The
// resource path is resolved at read time because several ClickUp endpoints are
// parameterised by team_id or space_id.
type streamShape struct {
	// recordsPath is the dotted JSON path to the records array in the response
	// envelope (e.g. "teams", "spaces", "tasks").
	recordsPath string
	// paginated marks endpoints that use ClickUp's page-based pagination
	// ({...,"last_page":bool}); only the tasks endpoint does this.
	paginated bool
	// mapRecord flattens a raw ClickUp object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// clickupStreamShapes is the per-stream routing table. Adding a stream means
// adding an entry here, a path builder in resolveEndpoint, and a Stream
// definition in clickupStreams.
var clickupStreamShapes = map[string]streamShape{
	"teams":   {recordsPath: "teams", mapRecord: clickupTeamRecord},
	"spaces":  {recordsPath: "spaces", mapRecord: clickupSpaceRecord},
	"folders": {recordsPath: "folders", mapRecord: clickupFolderRecord},
	"lists":   {recordsPath: "lists", mapRecord: clickupListRecord},
	"tasks":   {recordsPath: "tasks", paginated: true, mapRecord: clickupTaskRecord},
}

// clickupStreams returns the connector's published stream catalog. Every ClickUp
// object exposes a string id; tasks additionally carry a millisecond
// date_updated used as the incremental cursor.
func clickupStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "teams",
			Description: "ClickUp workspaces (called teams in the v2 API).",
			PrimaryKey:  []string{"id"},
			Fields:      clickupTeamFields(),
		},
		{
			Name:        "spaces",
			Description: "ClickUp spaces within a workspace.",
			PrimaryKey:  []string{"id"},
			Fields:      clickupSpaceFields(),
		},
		{
			Name:        "folders",
			Description: "ClickUp folders within a space.",
			PrimaryKey:  []string{"id"},
			Fields:      clickupFolderFields(),
		},
		{
			Name:        "lists",
			Description: "ClickUp folderless lists within a space.",
			PrimaryKey:  []string{"id"},
			Fields:      clickupListFields(),
		},
		{
			Name:         "tasks",
			Description:  "ClickUp tasks within a workspace.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_updated"},
			Fields:       clickupTaskFields(),
		},
	}
}

func clickupTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "avatar", Type: "string"},
	}
}

func clickupSpaceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "private", Type: "boolean"},
		{Name: "archived", Type: "boolean"},
		{Name: "multiple_assignees", Type: "boolean"},
	}
}

func clickupFolderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "orderindex", Type: "integer"},
		{Name: "hidden", Type: "boolean"},
		{Name: "archived", Type: "boolean"},
		{Name: "task_count", Type: "string"},
		{Name: "space_id", Type: "string"},
	}
}

func clickupListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "orderindex", Type: "integer"},
		{Name: "archived", Type: "boolean"},
		{Name: "task_count", Type: "integer"},
		{Name: "space_id", Type: "string"},
	}
}

func clickupTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "date_created", Type: "string"},
		{Name: "date_updated", Type: "string"},
		{Name: "date_closed", Type: "string"},
		{Name: "creator_id", Type: "string"},
		{Name: "list_id", Type: "string"},
		{Name: "folder_id", Type: "string"},
		{Name: "space_id", Type: "string"},
		{Name: "url", Type: "string"},
	}
}

func clickupTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":     item["id"],
		"name":   item["name"],
		"color":  item["color"],
		"avatar": item["avatar"],
	}
}

func clickupSpaceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"name":               item["name"],
		"private":            item["private"],
		"archived":           item["archived"],
		"multiple_assignees": item["multiple_assignees"],
	}
}

func clickupFolderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"orderindex": item["orderindex"],
		"hidden":     item["hidden"],
		"archived":   item["archived"],
		"task_count": item["task_count"],
		"space_id":   nestedID(item["space"]),
	}
}

func clickupListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"orderindex": item["orderindex"],
		"archived":   item["archived"],
		"task_count": item["task_count"],
		"space_id":   nestedID(item["space"]),
	}
}

func clickupTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"status":       statusName(item["status"]),
		"date_created": item["date_created"],
		"date_updated": item["date_updated"],
		"date_closed":  item["date_closed"],
		"creator_id":   nestedID(item["creator"]),
		"list_id":      nestedID(item["list"]),
		"folder_id":    nestedID(item["folder"]),
		"space_id":     nestedID(item["space"]),
		"url":          item["url"],
	}
}

// nestedID extracts the "id" field from a nested object (ClickUp nests space,
// list, folder, creator references as objects), returning nil if absent.
func nestedID(v any) any {
	obj, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	return obj["id"]
}

// statusName extracts the human-readable status from ClickUp's nested status
// object ({"status":"open",...}); falls back to the raw value for plain strings.
func statusName(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return t["status"]
	default:
		return t
	}
}
