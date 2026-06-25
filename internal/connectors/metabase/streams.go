package metabase

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Metabase API resource path (relative
// to instance_api_url) it reads from, and the record mapper that projects its
// objects.
type streamEndpoint struct {
	// resource is the Metabase list endpoint path segment (e.g. "card").
	resource string
	// mapRecord projects a raw Metabase object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// metabaseStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in metabaseStreams; the read
// path is fully data-driven from this table.
//
// Metabase list endpoints are singular paths (/api/card, /api/dashboard, ...)
// and the connector publishes pluralized, snake-cased stream names.
var metabaseStreamEndpoints = map[string]streamEndpoint{
	"cards":       {resource: "card", mapRecord: metabaseCardRecord},
	"dashboards":  {resource: "dashboard", mapRecord: metabaseDashboardRecord},
	"collections": {resource: "collection", mapRecord: metabaseCollectionRecord},
	"databases":   {resource: "database", mapRecord: metabaseDatabaseRecord},
	"users":       {resource: "user", mapRecord: metabaseUserRecord},
}

// metabaseStreams returns the connector's published stream catalog. Every
// Metabase object exposes an integer id, so the primary key is ["id"] across the
// board. Metabase's API is full-refresh only (no incremental cursors), but
// updated_at is exposed where present for downstream convenience.
func metabaseStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "cards",
			Description: "Metabase questions (saved queries), known as cards.",
			PrimaryKey:  []string{"id"},
			Fields:      metabaseCardFields(),
		},
		{
			Name:        "dashboards",
			Description: "Metabase dashboards.",
			PrimaryKey:  []string{"id"},
			Fields:      metabaseDashboardFields(),
		},
		{
			Name:        "collections",
			Description: "Metabase collections that organize cards and dashboards.",
			PrimaryKey:  []string{"id"},
			Fields:      metabaseCollectionFields(),
		},
		{
			Name:        "databases",
			Description: "Databases connected to the Metabase instance.",
			PrimaryKey:  []string{"id"},
			Fields:      metabaseDatabaseFields(),
		},
		{
			Name:        "users",
			Description: "Metabase user accounts.",
			PrimaryKey:  []string{"id"},
			Fields:      metabaseUserFields(),
		},
	}
}

func metabaseCardFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "collection_id", Type: "integer"},
		{Name: "database_id", Type: "integer"},
		{Name: "query_type", Type: "string"},
		{Name: "display", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "creator_id", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func metabaseDashboardFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "collection_id", Type: "integer"},
		{Name: "archived", Type: "boolean"},
		{Name: "creator_id", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func metabaseCollectionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "personal_owner_id", Type: "integer"},
		{Name: "location", Type: "string"},
	}
}

func metabaseDatabaseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "engine", Type: "string"},
		{Name: "is_sample", Type: "boolean"},
		{Name: "is_on_demand", Type: "boolean"},
		{Name: "timezone", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func metabaseUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "email", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "common_name", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "is_superuser", Type: "boolean"},
		{Name: "last_login", Type: "string"},
		{Name: "date_joined", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func metabaseCardRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"description":   item["description"],
		"collection_id": item["collection_id"],
		"database_id":   item["database_id"],
		"query_type":    item["query_type"],
		"display":       item["display"],
		"archived":      item["archived"],
		"creator_id":    item["creator_id"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func metabaseDashboardRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"description":   item["description"],
		"collection_id": item["collection_id"],
		"archived":      item["archived"],
		"creator_id":    item["creator_id"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func metabaseCollectionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"description":       item["description"],
		"slug":              item["slug"],
		"archived":          item["archived"],
		"personal_owner_id": item["personal_owner_id"],
		"location":          item["location"],
	}
}

func metabaseDatabaseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"engine":       item["engine"],
		"is_sample":    item["is_sample"],
		"is_on_demand": item["is_on_demand"],
		"timezone":     item["timezone"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func metabaseUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"email":        item["email"],
		"first_name":   item["first_name"],
		"last_name":    item["last_name"],
		"common_name":  item["common_name"],
		"is_active":    item["is_active"],
		"is_superuser": item["is_superuser"],
		"last_login":   item["last_login"],
		"date_joined":  item["date_joined"],
		"updated_at":   item["updated_at"],
	}
}
