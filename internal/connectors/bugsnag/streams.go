package bugsnag

import "polymetrics.ai/internal/connectors"

// scope describes the parent resource a stream's endpoint is nested under in the
// Bugsnag Data Access API. Bugsnag resources form a hierarchy:
//
//	/user/organizations                                (root)
//	/organizations/{organization_id}/projects          (per-org)
//	/organizations/{organization_id}/collaborators      (per-org)
//	/projects/{project_id}/errors                       (per-project)
//	/projects/{project_id}/events                       (per-project)
//	/projects/{project_id}/releases                     (per-project)
//
// The read path resolves the required parent id(s) from config, falling back to
// auto-discovery via the parent endpoint(s) when not configured.
type scope int

const (
	scopeRoot         scope = iota // no parent; endpoint is a fixed path
	scopeOrganization              // requires an organization id
	scopeProject                   // requires a project id
)

// streamEndpoint maps a stream to its Bugsnag endpoint shape. pathFor builds the
// resource path given the resolved parent id (empty for scopeRoot). mapRecord
// flattens a raw Bugsnag object into a connectors.Record.
type streamEndpoint struct {
	scope     scope
	pathFor   func(parentID string) string
	mapRecord func(map[string]any) connectors.Record
}

// bugsnagStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in bugsnagStreams; the read
// path is fully data-driven from this table.
var bugsnagStreamEndpoints = map[string]streamEndpoint{
	"organizations": {
		scope:     scopeRoot,
		pathFor:   func(string) string { return "user/organizations" },
		mapRecord: organizationRecord,
	},
	"projects": {
		scope:     scopeOrganization,
		pathFor:   func(org string) string { return "organizations/" + org + "/projects" },
		mapRecord: projectRecord,
	},
	"collaborators": {
		scope:     scopeOrganization,
		pathFor:   func(org string) string { return "organizations/" + org + "/collaborators" },
		mapRecord: collaboratorRecord,
	},
	"errors": {
		scope:     scopeProject,
		pathFor:   func(proj string) string { return "projects/" + proj + "/errors" },
		mapRecord: errorRecord,
	},
	"events": {
		scope:     scopeProject,
		pathFor:   func(proj string) string { return "projects/" + proj + "/events" },
		mapRecord: eventRecord,
	},
	"releases": {
		scope:     scopeProject,
		pathFor:   func(proj string) string { return "projects/" + proj + "/releases" },
		mapRecord: releaseRecord,
	},
}

// bugsnagStreams returns the connector's published stream catalog. Every Bugsnag
// resource exposes a string id, so the primary key is ["id"] across the board.
// Errors/events/releases carry timestamps usable as incremental cursors.
func bugsnagStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "organizations",
			Description: "Bugsnag organizations the authenticated user belongs to.",
			PrimaryKey:  []string{"id"},
			Fields:      organizationFields(),
		},
		{
			Name:        "projects",
			Description: "Projects within an organization.",
			PrimaryKey:  []string{"id"},
			Fields:      projectFields(),
		},
		{
			Name:        "collaborators",
			Description: "Collaborators within an organization.",
			PrimaryKey:  []string{"id"},
			Fields:      collaboratorFields(),
		},
		{
			Name:         "errors",
			Description:  "Errors recorded in a project.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_seen"},
			Fields:       errorFields(),
		},
		{
			Name:         "events",
			Description:  "Individual error events in a project.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"received_at"},
			Fields:       eventFields(),
		},
		{
			Name:         "releases",
			Description:  "Releases recorded in a project.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"release_time"},
			Fields:       releaseFields(),
		},
	}
}

func organizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "api_key", Type: "string"},
		{Name: "auto_upgrade", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "projects_url", Type: "string"},
		{Name: "collaborators_url", Type: "string"},
	}
}

func projectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "organization_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "api_key", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "open_error_count", Type: "integer"},
		{Name: "for_review_error_count", Type: "integer"},
		{Name: "collaborators_count", Type: "integer"},
		{Name: "errors_url", Type: "string"},
		{Name: "events_url", Type: "string"},
		{Name: "html_url", Type: "string"},
	}
}

func collaboratorFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "is_admin", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "last_request_at", Type: "timestamp"},
		{Name: "two_factor_enabled", Type: "boolean"},
		{Name: "pending_invitation", Type: "boolean"},
	}
}

func errorFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "project_id", Type: "string"},
		{Name: "error_class", Type: "string"},
		{Name: "message", Type: "string"},
		{Name: "context", Type: "string"},
		{Name: "severity", Type: "string"},
		{Name: "original_severity", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "first_seen", Type: "timestamp"},
		{Name: "last_seen", Type: "timestamp"},
		{Name: "events_count", Type: "integer"},
		{Name: "comment_count", Type: "integer"},
		{Name: "url", Type: "string"},
	}
}

func eventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "project_id", Type: "string"},
		{Name: "error_id", Type: "string"},
		{Name: "context", Type: "string"},
		{Name: "severity", Type: "string"},
		{Name: "unhandled", Type: "boolean"},
		{Name: "received_at", Type: "timestamp"},
		{Name: "is_full_report", Type: "boolean"},
		{Name: "url", Type: "string"},
	}
}

func releaseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "project_id", Type: "string"},
		{Name: "release_group_id", Type: "string"},
		{Name: "app_version", Type: "string"},
		{Name: "app_version_code", Type: "string"},
		{Name: "app_bundle_version", Type: "string"},
		{Name: "build_label", Type: "string"},
		{Name: "release_stage", Type: "string"},
		{Name: "release_source", Type: "string"},
		{Name: "release_time", Type: "timestamp"},
		{Name: "errors_seen_count", Type: "integer"},
		{Name: "errors_introduced_count", Type: "integer"},
	}
}

func organizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"slug":              item["slug"],
		"api_key":           item["api_key"],
		"auto_upgrade":      item["auto_upgrade"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
		"projects_url":      item["projects_url"],
		"collaborators_url": item["collaborators_url"],
	}
}

func projectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                     item["id"],
		"organization_id":        item["organization_id"],
		"name":                   item["name"],
		"slug":                   item["slug"],
		"api_key":                item["api_key"],
		"type":                   item["type"],
		"language":               item["language"],
		"created_at":             item["created_at"],
		"updated_at":             item["updated_at"],
		"open_error_count":       item["open_error_count"],
		"for_review_error_count": item["for_review_error_count"],
		"collaborators_count":    item["collaborators_count"],
		"errors_url":             item["errors_url"],
		"events_url":             item["events_url"],
		"html_url":               item["html_url"],
	}
}

func collaboratorRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"name":               item["name"],
		"email":              item["email"],
		"is_admin":           item["is_admin"],
		"created_at":         item["created_at"],
		"last_request_at":    item["last_request_at"],
		"two_factor_enabled": item["two_factor_enabled"],
		"pending_invitation": item["pending_invitation"],
	}
}

func errorRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"project_id":        item["project_id"],
		"error_class":       item["error_class"],
		"message":           item["message"],
		"context":           item["context"],
		"severity":          item["severity"],
		"original_severity": item["original_severity"],
		"status":            item["status"],
		"first_seen":        item["first_seen"],
		"last_seen":         item["last_seen"],
		"events_count":      item["events"],
		"comment_count":     item["comment_count"],
		"url":               item["url"],
	}
}

func eventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"project_id":     item["project_id"],
		"error_id":       item["error_id"],
		"context":        item["context"],
		"severity":       item["severity"],
		"unhandled":      item["unhandled"],
		"received_at":    item["received_at"],
		"is_full_report": item["is_full_report"],
		"url":            item["url"],
	}
}

func releaseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                      item["id"],
		"project_id":              item["project_id"],
		"release_group_id":        item["release_group_id"],
		"app_version":             item["app_version"],
		"app_version_code":        item["app_version_code"],
		"app_bundle_version":      item["app_bundle_version"],
		"build_label":             item["build_label"],
		"release_stage":           item["release_stage"],
		"release_source":          item["release_source"],
		"release_time":            item["release_time"],
		"errors_seen_count":       item["errors_seen_count"],
		"errors_introduced_count": item["errors_introduced_count"],
	}
}
