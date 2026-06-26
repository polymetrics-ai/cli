package firehydrant

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the FireHydrant API resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects.
type streamEndpoint struct {
	// resource is the FireHydrant list endpoint path segment (e.g. "incidents").
	resource string
	// mapRecord flattens a raw FireHydrant object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// firehydrantStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in firehydrantStreams; the
// read path is fully data-driven from this table.
var firehydrantStreamEndpoints = map[string]streamEndpoint{
	"incidents":       {resource: "incidents", mapRecord: firehydrantIncidentRecord},
	"services":        {resource: "services", mapRecord: firehydrantServiceRecord},
	"teams":           {resource: "teams", mapRecord: firehydrantTeamRecord},
	"environments":    {resource: "environments", mapRecord: firehydrantEnvironmentRecord},
	"functionalities": {resource: "functionalities", mapRecord: firehydrantFunctionalityRecord},
}

// firehydrantStreams returns the connector's published stream catalog. Every
// FireHydrant object exposes a string UUID id and RFC3339 created_at/updated_at
// timestamps, so the primary key is ["id"] and the incremental cursor field is
// ["updated_at"] across the board.
func firehydrantStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "incidents",
			Description:  "FireHydrant incidents.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       firehydrantIncidentFields(),
		},
		{
			Name:         "services",
			Description:  "FireHydrant catalog services.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       firehydrantServiceFields(),
		},
		{
			Name:         "teams",
			Description:  "FireHydrant teams.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       firehydrantTeamFields(),
		},
		{
			Name:         "environments",
			Description:  "FireHydrant environments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       firehydrantEnvironmentFields(),
		},
		{
			Name:         "functionalities",
			Description:  "FireHydrant functionalities.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       firehydrantFunctionalityFields(),
		},
	}
}

func firehydrantIncidentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "number", Type: "integer"},
		{Name: "description", Type: "string"},
		{Name: "summary", Type: "string"},
		{Name: "current_milestone", Type: "string"},
		{Name: "severity", Type: "string"},
		{Name: "priority", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "started_at", Type: "timestamp"},
		{Name: "resolved_at", Type: "timestamp"},
	}
}

func firehydrantServiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "service_tier", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func firehydrantTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func firehydrantEnvironmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func firehydrantFunctionalityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func firehydrantIncidentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"number":            item["number"],
		"description":       item["description"],
		"summary":           item["summary"],
		"current_milestone": item["current_milestone"],
		"severity":          item["severity"],
		"priority":          item["priority"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
		"started_at":        item["started_at"],
		"resolved_at":       item["resolved_at"],
	}
}

func firehydrantServiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"description":  item["description"],
		"slug":         item["slug"],
		"service_tier": item["service_tier"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func firehydrantTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"slug":        item["slug"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func firehydrantEnvironmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func firehydrantFunctionalityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"slug":        item["slug"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}
