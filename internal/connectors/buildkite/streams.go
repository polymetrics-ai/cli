package buildkite

import "polymetrics.ai/internal/connectors"

// scope describes whether a stream's endpoint is rooted at the top-level API or
// scoped under a specific organization slug.
type scope int

const (
	// scopeTopLevel endpoints do not require an organization slug
	// (e.g. /organizations).
	scopeTopLevel scope = iota
	// scopeOrganization endpoints are nested under
	// /organizations/{org.slug}/... and require config organization.
	scopeOrganization
)

// streamEndpoint maps a stream name to its Buildkite REST resource and the
// record mapper that flattens the raw object. Buildkite list endpoints return a
// top-level JSON array, so the records path is the root ("").
type streamEndpoint struct {
	// resource is the path suffix. For scopeOrganization streams it is appended
	// after /organizations/{org.slug}/; for scopeTopLevel it is used as-is.
	resource  string
	scope     scope
	mapRecord func(map[string]any) connectors.Record
}

// buildkiteStreamEndpoints is the per-stream routing table. Adding a stream is a
// matter of one entry here plus a Stream definition in buildkiteStreams.
var buildkiteStreamEndpoints = map[string]streamEndpoint{
	"organizations": {resource: "organizations", scope: scopeTopLevel, mapRecord: organizationRecord},
	"pipelines":     {resource: "pipelines", scope: scopeOrganization, mapRecord: pipelineRecord},
	"builds":        {resource: "builds", scope: scopeOrganization, mapRecord: buildRecord},
	"agents":        {resource: "agents", scope: scopeOrganization, mapRecord: agentRecord},
}

// buildkiteStreams returns the connector's published stream catalog. Every
// Buildkite object exposes a string UUID id and an RFC3339 created_at, so the
// primary key is ["id"] and the incremental cursor is ["created_at"].
func buildkiteStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "organizations",
			Description:  "Buildkite organizations accessible to the API token.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       organizationFields(),
		},
		{
			Name:         "pipelines",
			Description:  "Pipelines in the configured organization.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       pipelineFields(),
		},
		{
			Name:         "builds",
			Description:  "Builds across all pipelines in the configured organization.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       buildFields(),
		},
		{
			Name:         "agents",
			Description:  "Agents registered in the configured organization.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       agentFields(),
		},
	}
}

func organizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "graphql_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "web_url", Type: "string"},
		{Name: "pipelines_url", Type: "string"},
		{Name: "agents_url", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func pipelineFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "graphql_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "web_url", Type: "string"},
		{Name: "repository", Type: "string"},
		{Name: "default_branch", Type: "string"},
		{Name: "visibility", Type: "string"},
		{Name: "builds_url", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "archived_at", Type: "timestamp"},
	}
}

func buildFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "graphql_id", Type: "string"},
		{Name: "number", Type: "integer"},
		{Name: "state", Type: "string"},
		{Name: "blocked", Type: "boolean"},
		{Name: "message", Type: "string"},
		{Name: "commit", Type: "string"},
		{Name: "branch", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "web_url", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "scheduled_at", Type: "timestamp"},
		{Name: "started_at", Type: "timestamp"},
		{Name: "finished_at", Type: "timestamp"},
	}
}

func agentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "graphql_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "connection_state", Type: "string"},
		{Name: "hostname", Type: "string"},
		{Name: "ip_address", Type: "string"},
		{Name: "user_agent", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "web_url", Type: "string"},
		{Name: "priority", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "last_job_finished_at", Type: "timestamp"},
	}
}

func organizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"graphql_id":    item["graphql_id"],
		"name":          item["name"],
		"slug":          item["slug"],
		"url":           item["url"],
		"web_url":       item["web_url"],
		"pipelines_url": item["pipelines_url"],
		"agents_url":    item["agents_url"],
		"created_at":    item["created_at"],
	}
}

func pipelineRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"graphql_id":     item["graphql_id"],
		"name":           item["name"],
		"slug":           item["slug"],
		"description":    item["description"],
		"url":            item["url"],
		"web_url":        item["web_url"],
		"repository":     item["repository"],
		"default_branch": item["default_branch"],
		"visibility":     item["visibility"],
		"builds_url":     item["builds_url"],
		"created_at":     item["created_at"],
		"archived_at":    item["archived_at"],
	}
}

func buildRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"graphql_id":   item["graphql_id"],
		"number":       item["number"],
		"state":        item["state"],
		"blocked":      item["blocked"],
		"message":      item["message"],
		"commit":       item["commit"],
		"branch":       item["branch"],
		"source":       item["source"],
		"url":          item["url"],
		"web_url":      item["web_url"],
		"created_at":   item["created_at"],
		"scheduled_at": item["scheduled_at"],
		"started_at":   item["started_at"],
		"finished_at":  item["finished_at"],
	}
}

func agentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                   item["id"],
		"graphql_id":           item["graphql_id"],
		"name":                 item["name"],
		"connection_state":     item["connection_state"],
		"hostname":             item["hostname"],
		"ip_address":           item["ip_address"],
		"user_agent":           item["user_agent"],
		"version":              item["version"],
		"url":                  item["url"],
		"web_url":              item["web_url"],
		"priority":             item["priority"],
		"created_at":           item["created_at"],
		"last_job_finished_at": item["last_job_finished_at"],
	}
}
