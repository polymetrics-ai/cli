package circleci

import "polymetrics.ai/internal/connectors"

// streamShape distinguishes the two response shapes the CircleCI v2 API returns:
// a paginated list ({"items":[...],"next_page_token":...}) versus a single
// resource object (the project endpoint).
type streamShape int

const (
	// shapeList is a paginated {"items":[...],"next_page_token":...} response.
	shapeList streamShape = iota
	// shapeObject is a single resource object returned at the response root.
	shapeObject
)

// streamEndpoint maps a stream name to its CircleCI API endpoint, the response
// shape, and the record mapper. paths are resolved against base_url at read
// time; segments wrapped in braces are substituted from config (project slug,
// pipeline id, workflow id). The read path is fully data-driven from this table.
type streamEndpoint struct {
	// resolvePath builds the request path from the runtime config, substituting
	// the configured project slug / pipeline id / workflow id. It returns an
	// error if a required config value is missing.
	resolvePath func(p pathParams) (string, error)
	// shape selects how the response body is decoded.
	shape streamShape
	// mapRecord flattens a raw CircleCI object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// pathParams holds the config-derived values used to build endpoint paths.
type pathParams struct {
	projectSlug string
	pipelineID  string
	workflowID  string
}

// circleciStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in circleciStreams.
var circleciStreamEndpoints = map[string]streamEndpoint{
	"projects": {
		resolvePath: func(p pathParams) (string, error) {
			if p.projectSlug == "" {
				return "", errMissingProjectSlug
			}
			return "project/" + p.projectSlug, nil
		},
		shape:     shapeObject,
		mapRecord: circleciProjectRecord,
	},
	"pipelines": {
		resolvePath: func(p pathParams) (string, error) {
			if p.projectSlug == "" {
				return "", errMissingProjectSlug
			}
			return "project/" + p.projectSlug + "/pipeline", nil
		},
		shape:     shapeList,
		mapRecord: circleciPipelineRecord,
	},
	"workflows": {
		resolvePath: func(p pathParams) (string, error) {
			if p.pipelineID == "" {
				return "", errMissingPipelineID
			}
			return "pipeline/" + p.pipelineID + "/workflow", nil
		},
		shape:     shapeList,
		mapRecord: circleciWorkflowRecord,
	},
	"jobs": {
		resolvePath: func(p pathParams) (string, error) {
			if p.workflowID == "" {
				return "", errMissingWorkflowID
			}
			return "workflow/" + p.workflowID + "/job", nil
		},
		shape:     shapeList,
		mapRecord: circleciJobRecord,
	},
}

// circleciStreams returns the connector's published stream catalog.
func circleciStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "projects",
			Description:  "CircleCI project metadata for the configured project_slug.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       circleciProjectFields(),
		},
		{
			Name:         "pipelines",
			Description:  "CircleCI pipelines for the configured project_slug.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       circleciPipelineFields(),
		},
		{
			Name:         "workflows",
			Description:  "CircleCI workflows for the configured pipeline_id.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       circleciWorkflowFields(),
		},
		{
			Name:         "jobs",
			Description:  "CircleCI jobs for the configured workflow_id.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"started_at"},
			Fields:       circleciJobFields(),
		},
	}
}

func circleciProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "organization_name", Type: "string"},
		{Name: "organization_slug", Type: "string"},
		{Name: "organization_id", Type: "string"},
		{Name: "vcs_url", Type: "string"},
		{Name: "default_branch", Type: "string"},
	}
}

func circleciPipelineFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "number", Type: "integer"},
		{Name: "project_slug", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func circleciWorkflowFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "pipeline_id", Type: "string"},
		{Name: "pipeline_number", Type: "integer"},
		{Name: "project_slug", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "stopped_at", Type: "timestamp"},
	}
}

func circleciJobFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "job_number", Type: "integer"},
		{Name: "project_slug", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "started_at", Type: "timestamp"},
		{Name: "stopped_at", Type: "timestamp"},
	}
}

func circleciProjectRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":                item["id"],
		"slug":              item["slug"],
		"name":              item["name"],
		"organization_name": item["organization_name"],
		"organization_slug": item["organization_slug"],
		"organization_id":   item["organization_id"],
		"vcs_url":           item["vcs_url"],
	}
	// vcs_info nests default_branch and vcs_url on the project payload.
	if vcs, ok := item["vcs_info"].(map[string]any); ok {
		rec["default_branch"] = vcs["default_branch"]
		if rec["vcs_url"] == nil {
			rec["vcs_url"] = vcs["vcs_url"]
		}
	}
	return rec
}

func circleciPipelineRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"number":       item["number"],
		"project_slug": item["project_slug"],
		"state":        item["state"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func circleciWorkflowRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            item["name"],
		"pipeline_id":     item["pipeline_id"],
		"pipeline_number": item["pipeline_number"],
		"project_slug":    item["project_slug"],
		"status":          item["status"],
		"created_at":      item["created_at"],
		"stopped_at":      item["stopped_at"],
	}
}

func circleciJobRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"job_number":   item["job_number"],
		"project_slug": item["project_slug"],
		"type":         item["type"],
		"status":       item["status"],
		"started_at":   item["started_at"],
		"stopped_at":   item["stopped_at"],
	}
}
