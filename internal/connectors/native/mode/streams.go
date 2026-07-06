package mode

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the workspace-scoped Mode API resource
// path segment it reads from, the HAL+JSON _embedded key holding its records,
// and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the path segment after /{workspace}/ (e.g. "spaces").
	resource string
	// embedded is the key under "_embedded" holding the records array
	// (Mode HAL responses look like {"_embedded":{"spaces":[...]}}).
	embedded string
	// mapRecord flattens a raw Mode object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// modeStreamEndpoints is the per-stream routing table. Every stream is a
// top-level, workspace-scoped list endpoint. Adding a stream means adding one
// entry here plus a Stream definition in modeStreams; the read path is fully
// data-driven from this table.
var modeStreamEndpoints = map[string]streamEndpoint{
	"spaces":       {resource: "spaces", embedded: "spaces", mapRecord: modeSpaceRecord},
	"reports":      {resource: "reports", embedded: "reports", mapRecord: modeReportRecord},
	"data_sources": {resource: "data_sources", embedded: "data_sources", mapRecord: modeDataSourceRecord},
	"groups":       {resource: "groups", embedded: "groups", mapRecord: modeGroupRecord},
	"memberships":  {resource: "memberships", embedded: "memberships", mapRecord: modeMembershipRecord},
}

// modeStreams returns the connector's published stream catalog. Every Mode
// object exposes a string `token` (its stable identifier) and most carry
// `created_at`/`updated_at` RFC3339 timestamps, so the primary key is ["token"]
// and the incremental cursor field is ["updated_at"] where present.
func modeStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "spaces",
			Description:  "Mode collections (spaces) in the workspace.",
			PrimaryKey:   []string{"token"},
			CursorFields: []string{"updated_at"},
			Fields:       modeSpaceFields(),
		},
		{
			Name:         "reports",
			Description:  "Mode reports in the workspace.",
			PrimaryKey:   []string{"token"},
			CursorFields: []string{"updated_at"},
			Fields:       modeReportFields(),
		},
		{
			Name:         "data_sources",
			Description:  "Mode data source connections in the workspace.",
			PrimaryKey:   []string{"token"},
			CursorFields: []string{"updated_at"},
			Fields:       modeDataSourceFields(),
		},
		{
			Name:         "groups",
			Description:  "Mode user groups in the workspace.",
			PrimaryKey:   []string{"token"},
			CursorFields: []string{"updated_at"},
			Fields:       modeGroupFields(),
		},
		{
			Name:         "memberships",
			Description:  "Mode workspace memberships.",
			PrimaryKey:   []string{"token"},
			CursorFields: []string{"created_at"},
			Fields:       modeMembershipFields(),
		},
	}
}

func modeSpaceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "token", Type: "string"},
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "space_type", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "restricted", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func modeReportFields() []connectors.Field {
	return []connectors.Field{
		{Name: "token", Type: "string"},
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "space_token", Type: "string"},
		{Name: "account_username", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "public", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "last_run_at", Type: "timestamp"},
	}
}

func modeDataSourceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "token", Type: "string"},
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "adapter", Type: "string"},
		{Name: "database", Type: "string"},
		{Name: "host", Type: "string"},
		{Name: "queryable", Type: "boolean"},
		{Name: "public", Type: "boolean"},
		{Name: "asleep", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func modeGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "token", Type: "string"},
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func modeMembershipFields() []connectors.Field {
	return []connectors.Field{
		{Name: "token", Type: "string"},
		{Name: "id", Type: "integer"},
		{Name: "admin", Type: "boolean"},
		{Name: "username", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func modeSpaceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"token":       item["token"],
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"space_type":  item["space_type"],
		"state":       item["state"],
		"restricted":  item["restricted"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func modeReportRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"token":            item["token"],
		"id":               item["id"],
		"name":             item["name"],
		"description":      item["description"],
		"space_token":      item["space_token"],
		"account_username": item["account_username"],
		"archived":         item["archived"],
		"public":           item["public"],
		"created_at":       item["created_at"],
		"updated_at":       item["updated_at"],
		"last_run_at":      item["last_run_at"],
	}
}

func modeDataSourceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"token":       item["token"],
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"adapter":     item["adapter"],
		"database":    item["database"],
		"host":        item["host"],
		"queryable":   item["queryable"],
		"public":      item["public"],
		"asleep":      item["asleep"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func modeGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"token":       item["token"],
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"state":       item["state"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func modeMembershipRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"token":      item["token"],
		"id":         item["id"],
		"admin":      item["admin"],
		"username":   item["username"],
		"email":      item["email"],
		"created_at": item["created_at"],
	}
}
