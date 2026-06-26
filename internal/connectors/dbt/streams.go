package dbt

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the dbt Cloud account-scoped resource path
// segment (relative to /accounts/{account_id}/) it reads from, and the record
// mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the dbt Cloud list endpoint path segment (e.g. "projects").
	// dbt Cloud v2 list paths end in a trailing slash.
	resource string
	// mapRecord flattens a raw dbt object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// dbtStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in dbtStreams; the read path is
// fully data-driven from this table.
var dbtStreamEndpoints = map[string]streamEndpoint{
	"projects":     {resource: "projects/", mapRecord: dbtProjectRecord},
	"runs":         {resource: "runs/", mapRecord: dbtRunRecord},
	"repositories": {resource: "repositories/", mapRecord: dbtRepositoryRecord},
	"users":        {resource: "users/", mapRecord: dbtUserRecord},
	"environments": {resource: "environments/", mapRecord: dbtEnvironmentRecord},
}

// dbtStreams returns the connector's published stream catalog. Every dbt Cloud
// object exposes an integer id, so the primary key is ["id"] across the board.
// The Administrative API v2 only supports full-refresh, so no cursor fields are
// declared.
func dbtStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "projects",
			Description: "dbt Cloud projects for the account.",
			PrimaryKey:  []string{"id"},
			Fields:      dbtProjectFields(),
		},
		{
			Name:        "runs",
			Description: "dbt Cloud job runs for the account.",
			PrimaryKey:  []string{"id"},
			Fields:      dbtRunFields(),
		},
		{
			Name:        "repositories",
			Description: "dbt Cloud repositories connected to the account.",
			PrimaryKey:  []string{"id"},
			Fields:      dbtRepositoryFields(),
		},
		{
			Name:        "users",
			Description: "dbt Cloud users in the account.",
			PrimaryKey:  []string{"id"},
			Fields:      dbtUserFields(),
		},
		{
			Name:        "environments",
			Description: "dbt Cloud environments for the account.",
			PrimaryKey:  []string{"id"},
			Fields:      dbtEnvironmentFields(),
		},
	}
}

func dbtProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "account_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "connection_id", Type: "integer"},
		{Name: "repository_id", Type: "integer"},
		{Name: "state", Type: "integer"},
		{Name: "dbt_project_subdirectory", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func dbtRunFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "account_id", Type: "integer"},
		{Name: "project_id", Type: "integer"},
		{Name: "job_definition_id", Type: "integer"},
		{Name: "environment_id", Type: "integer"},
		{Name: "status", Type: "integer"},
		{Name: "status_humanized", Type: "string"},
		{Name: "is_complete", Type: "boolean"},
		{Name: "is_error", Type: "boolean"},
		{Name: "is_cancelled", Type: "boolean"},
		{Name: "started_at", Type: "string"},
		{Name: "finished_at", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func dbtRepositoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "account_id", Type: "integer"},
		{Name: "project_id", Type: "integer"},
		{Name: "remote_url", Type: "string"},
		{Name: "remote_backend", Type: "string"},
		{Name: "git_clone_strategy", Type: "string"},
		{Name: "state", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func dbtUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "account_id", Type: "integer"},
		{Name: "email", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "fullname", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "created_at", Type: "string"},
	}
}

func dbtEnvironmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "account_id", Type: "integer"},
		{Name: "project_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "dbt_version", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "use_custom_branch", Type: "boolean"},
		{Name: "custom_branch", Type: "string"},
		{Name: "state", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func dbtProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                       item["id"],
		"account_id":               item["account_id"],
		"name":                     item["name"],
		"description":              item["description"],
		"connection_id":            item["connection_id"],
		"repository_id":            item["repository_id"],
		"state":                    item["state"],
		"dbt_project_subdirectory": item["dbt_project_subdirectory"],
		"created_at":               item["created_at"],
		"updated_at":               item["updated_at"],
	}
}

func dbtRunRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"account_id":        item["account_id"],
		"project_id":        item["project_id"],
		"job_definition_id": item["job_definition_id"],
		"environment_id":    item["environment_id"],
		"status":            item["status"],
		"status_humanized":  item["status_humanized"],
		"is_complete":       item["is_complete"],
		"is_error":          item["is_error"],
		"is_cancelled":      item["is_cancelled"],
		"started_at":        item["started_at"],
		"finished_at":       item["finished_at"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
	}
}

func dbtRepositoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"account_id":         item["account_id"],
		"project_id":         item["project_id"],
		"remote_url":         item["remote_url"],
		"remote_backend":     item["remote_backend"],
		"git_clone_strategy": item["git_clone_strategy"],
		"state":              item["state"],
		"created_at":         item["created_at"],
		"updated_at":         item["updated_at"],
	}
}

func dbtUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"account_id": item["account_id"],
		"email":      item["email"],
		"first_name": item["first_name"],
		"last_name":  item["last_name"],
		"fullname":   item["fullname"],
		"is_active":  item["is_active"],
		"created_at": item["created_at"],
	}
}

func dbtEnvironmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"account_id":        item["account_id"],
		"project_id":        item["project_id"],
		"name":              item["name"],
		"dbt_version":       item["dbt_version"],
		"type":              item["type"],
		"use_custom_branch": item["use_custom_branch"],
		"custom_branch":     item["custom_branch"],
		"state":             item["state"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
	}
}
