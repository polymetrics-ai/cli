package height

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Height API resource path (relative to
// base_url), the dotted JSON path the records live at in the response, and the
// record mapper that flattens its objects.
//
// Most Height list endpoints return a {"list":[...]} envelope, so recordsPath is
// "list". The workspace endpoint returns a single object at the root, so its
// recordsPath is "" (connsdk.RecordsAt treats a root object as a one-element set).
type streamEndpoint struct {
	resource    string
	recordsPath string
	// paginated is true for list endpoints that honor the nextPageToken/after
	// cursor pagination (tasks). Single-object or unpaginated endpoints set false.
	paginated bool
	mapRecord func(map[string]any) connectors.Record
}

// heightStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in heightStreams; the read path
// is fully data-driven from this table.
var heightStreamEndpoints = map[string]streamEndpoint{
	"tasks":           {resource: "tasks", recordsPath: "list", paginated: true, mapRecord: heightTaskRecord},
	"lists":           {resource: "lists", recordsPath: "list", mapRecord: heightListRecord},
	"field_templates": {resource: "fieldTemplates", recordsPath: "list", mapRecord: heightFieldTemplateRecord},
	"users":           {resource: "users", recordsPath: "list", mapRecord: heightUserRecord},
	"workspace":       {resource: "workspace", recordsPath: "", mapRecord: heightWorkspaceRecord},
}

// heightStreams returns the connector's published stream catalog. Every Height
// object exposes a string id; most also carry an RFC3339 createdAt timestamp used
// as the incremental cursor field.
func heightStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "tasks",
			Description:  "Height tasks.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       heightTaskFields(),
		},
		{
			Name:         "lists",
			Description:  "Height lists (projects, smartlists, views).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       heightListFields(),
		},
		{
			Name:        "field_templates",
			Description: "Height field templates (custom fields).",
			PrimaryKey:  []string{"id"},
			Fields:      heightFieldTemplateFields(),
		},
		{
			Name:         "users",
			Description:  "Height workspace users.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       heightUserFields(),
		},
		{
			Name:         "workspace",
			Description:  "Height workspace.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       heightWorkspaceFields(),
		},
	}
}

func heightTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "model", Type: "string"},
		{Name: "index", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "completed", Type: "boolean"},
		{Name: "completedAt", Type: "string"},
		{Name: "deleted", Type: "boolean"},
		{Name: "url", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "createdUserId", Type: "string"},
		{Name: "lastActivityAt", Type: "string"},
		{Name: "listIds", Type: "array"},
		{Name: "assigneesIds", Type: "array"},
		{Name: "parentTaskId", Type: "string"},
	}
}

func heightListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "model", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "key", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
		{Name: "userId", Type: "string"},
		{Name: "defaultList", Type: "boolean"},
		{Name: "visualization", Type: "string"},
	}
}

func heightFieldTemplateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "model", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "standardType", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "hidden", Type: "boolean"},
		{Name: "required", Type: "boolean"},
		{Name: "labels", Type: "array"},
	}
}

func heightUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "model", Type: "string"},
		{Name: "key", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "username", Type: "string"},
		{Name: "firstname", Type: "string"},
		{Name: "lastname", Type: "string"},
		{Name: "admin", Type: "boolean"},
		{Name: "state", Type: "string"},
		{Name: "deleted", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
		{Name: "signedUpAt", Type: "string"},
	}
}

func heightWorkspaceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "model", Type: "string"},
		{Name: "key", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "urlType", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "createdUserId", Type: "string"},
		{Name: "frozen", Type: "boolean"},
	}
}

func heightTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"model":          item["model"],
		"index":          item["index"],
		"name":           item["name"],
		"description":    item["description"],
		"status":         item["status"],
		"completed":      item["completed"],
		"completedAt":    item["completedAt"],
		"deleted":        item["deleted"],
		"url":            item["url"],
		"createdAt":      item["createdAt"],
		"createdUserId":  item["createdUserId"],
		"lastActivityAt": item["lastActivityAt"],
		"listIds":        item["listIds"],
		"assigneesIds":   item["assigneesIds"],
		"parentTaskId":   item["parentTaskId"],
	}
}

func heightListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"model":         item["model"],
		"type":          item["type"],
		"key":           item["key"],
		"name":          item["name"],
		"description":   item["description"],
		"url":           item["url"],
		"createdAt":     item["createdAt"],
		"updatedAt":     item["updatedAt"],
		"userId":        item["userId"],
		"defaultList":   item["defaultList"],
		"visualization": item["visualization"],
	}
}

func heightFieldTemplateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"model":        item["model"],
		"type":         item["type"],
		"name":         item["name"],
		"standardType": item["standardType"],
		"archived":     item["archived"],
		"hidden":       item["hidden"],
		"required":     item["required"],
		"labels":       item["labels"],
	}
}

func heightUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"model":      item["model"],
		"key":        item["key"],
		"email":      item["email"],
		"username":   item["username"],
		"firstname":  item["firstname"],
		"lastname":   item["lastname"],
		"admin":      item["admin"],
		"state":      item["state"],
		"deleted":    item["deleted"],
		"createdAt":  item["createdAt"],
		"signedUpAt": item["signedUpAt"],
	}
}

func heightWorkspaceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"model":         item["model"],
		"key":           item["key"],
		"name":          item["name"],
		"url":           item["url"],
		"urlType":       item["urlType"],
		"createdAt":     item["createdAt"],
		"createdUserId": item["createdUserId"],
		"frozen":        item["frozen"],
	}
}
