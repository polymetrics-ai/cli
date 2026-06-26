package nocrm

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the noCRM API resource path (relative to
// base_url, which already includes /api/v2) and the record mapper that flattens
// its objects.
type streamEndpoint struct {
	// resource is the noCRM list endpoint path segment (e.g. "leads").
	resource string
	// mapRecord flattens a raw noCRM object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// nocrmStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in nocrmStreams; the read path
// is fully data-driven from this table.
var nocrmStreamEndpoints = map[string]streamEndpoint{
	"leads":             {resource: "leads", mapRecord: nocrmLeadRecord},
	"pipelines":         {resource: "pipelines", mapRecord: nocrmPipelineRecord},
	"users":             {resource: "users", mapRecord: nocrmUserRecord},
	"teams":             {resource: "teams", mapRecord: nocrmTeamRecord},
	"prospecting_lists": {resource: "prospecting_lists", mapRecord: nocrmProspectingListRecord},
}

// nocrmStreams returns the connector's published stream catalog. Every noCRM
// object exposes a numeric id; the API supports only full_refresh syncs so no
// incremental cursor field is published.
func nocrmStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "leads",
			Description: "noCRM sales leads.",
			PrimaryKey:  []string{"id"},
			Fields:      nocrmLeadFields(),
		},
		{
			Name:        "pipelines",
			Description: "noCRM sales pipelines.",
			PrimaryKey:  []string{"id"},
			Fields:      nocrmPipelineFields(),
		},
		{
			Name:        "users",
			Description: "noCRM account users.",
			PrimaryKey:  []string{"id"},
			Fields:      nocrmUserFields(),
		},
		{
			Name:        "teams",
			Description: "noCRM teams.",
			PrimaryKey:  []string{"id"},
			Fields:      nocrmTeamFields(),
		},
		{
			Name:        "prospecting_lists",
			Description: "noCRM prospecting lists.",
			PrimaryKey:  []string{"id"},
			Fields:      nocrmProspectingListFields(),
		},
	}
}

func nocrmLeadFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "step", Type: "string"},
		{Name: "step_id", Type: "integer"},
		{Name: "pipeline", Type: "string"},
		{Name: "pipeline_id", Type: "integer"},
		{Name: "amount", Type: "number"},
		{Name: "currency", Type: "string"},
		{Name: "probability", Type: "integer"},
		{Name: "user_id", Type: "integer"},
		{Name: "client_folder_id", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "closed_at", Type: "string"},
		{Name: "remind_date", Type: "string"},
	}
}

func nocrmPipelineFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "position", Type: "integer"},
		{Name: "default", Type: "boolean"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func nocrmUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "email", Type: "string"},
		{Name: "firstname", Type: "string"},
		{Name: "lastname", Type: "string"},
		{Name: "admin", Type: "boolean"},
		{Name: "active", Type: "boolean"},
		{Name: "team_id", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func nocrmTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func nocrmProspectingListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "user_id", Type: "integer"},
		{Name: "archived", Type: "boolean"},
		{Name: "prospects_count", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func nocrmLeadRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"title":            item["title"],
		"status":           item["status"],
		"step":             item["step"],
		"step_id":          item["step_id"],
		"pipeline":         item["pipeline"],
		"pipeline_id":      item["pipeline_id"],
		"amount":           item["amount"],
		"currency":         item["currency"],
		"probability":      item["probability"],
		"user_id":          item["user_id"],
		"client_folder_id": item["client_folder_id"],
		"created_at":       item["created_at"],
		"updated_at":       item["updated_at"],
		"closed_at":        item["closed_at"],
		"remind_date":      item["remind_date"],
	}
}

func nocrmPipelineRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"position":   item["position"],
		"default":    item["default"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func nocrmUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"email":      item["email"],
		"firstname":  item["firstname"],
		"lastname":   item["lastname"],
		"admin":      item["admin"],
		"active":     item["active"],
		"team_id":    item["team_id"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func nocrmTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func nocrmProspectingListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"title":           item["title"],
		"user_id":         item["user_id"],
		"archived":        item["archived"],
		"prospects_count": item["prospects_count"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
	}
}
