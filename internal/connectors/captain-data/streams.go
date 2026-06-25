package captaindata

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Captain Data v3 API resource it reads
// from and the record mapper that flattens its objects.
//
// Captain Data list endpoints come in two shapes:
//   - top-level streams (workspace, workflows) return a JSON array at the root,
//     so recordsPath is "" and they are read with a single request.
//   - the job_results stream returns {results:[...], paging:{next, have_next_page}}
//     and is read with cursor pagination; recordsPath is "results".
//
// The jobs and job_results endpoints are scoped by a parent uid that is supplied
// via config (workflow_uid / job_uid); pathFor builds the concrete path.
type streamEndpoint struct {
	// resource is the static resource path for top-level streams (e.g.
	// "workflows"). For scoped streams it is empty and pathFor is used instead.
	resource string
	// recordsPath is the dotted JSON path to the records array ("" = root array).
	recordsPath string
	// paginated is true when the endpoint returns a paging cursor and must be
	// harvested across pages.
	paginated bool
	// scopeParam is the config key holding the parent uid for scoped streams
	// (empty for top-level streams).
	scopeParam string
	// pathFor builds the request path from a scope uid (only for scoped streams).
	pathFor func(scope string) string
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// captainDataStreamEndpoints is the per-stream routing table. The read path is
// fully data-driven from this table; adding a stream means adding one entry here
// plus a Stream definition in captainDataStreams.
var captainDataStreamEndpoints = map[string]streamEndpoint{
	"workspace": {resource: "workspace", recordsPath: "", mapRecord: workspaceRecord},
	"workflows": {resource: "workflows", recordsPath: "", mapRecord: workflowRecord},
	"jobs": {
		recordsPath: "",
		scopeParam:  "workflow_uid",
		pathFor:     func(scope string) string { return "workflows/" + scope + "/jobs" },
		mapRecord:   jobRecord,
	},
	"job_results": {
		recordsPath: "results",
		paginated:   true,
		scopeParam:  "job_uid",
		pathFor:     func(scope string) string { return "jobs/" + scope + "/results" },
		mapRecord:   jobResultRecord,
	},
}

// captainDataStreams returns the connector's published stream catalog. Every
// Captain Data object is keyed by a string "uid", so the primary key is ["uid"].
// The Captain Data source supports full-refresh only, so no cursor fields are
// published.
func captainDataStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "workspace",
			Description: "The Captain Data workspace the API key belongs to.",
			PrimaryKey:  []string{"uid"},
			Fields:      workspaceFields(),
		},
		{
			Name:        "workflows",
			Description: "Captain Data automation workflows.",
			PrimaryKey:  []string{"uid"},
			Fields:      workflowFields(),
		},
		{
			Name:        "jobs",
			Description: "Jobs (runs) for a workflow. Scoped by config workflow_uid.",
			PrimaryKey:  []string{"uid"},
			Fields:      jobFields(),
		},
		{
			Name:        "job_results",
			Description: "Result rows produced by a job. Scoped by config job_uid.",
			PrimaryKey:  []string{"uid"},
			Fields:      jobResultFields(),
		},
	}
}

func workspaceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uid", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "plan", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func workflowFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uid", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func jobFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uid", Type: "string"},
		{Name: "workflow_uid", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "ended_at", Type: "string"},
	}
}

func jobResultFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uid", Type: "string"},
		{Name: "job_uid", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "data", Type: "object"},
		{Name: "created_at", Type: "string"},
	}
}

func workspaceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uid":        item["uid"],
		"name":       item["name"],
		"plan":       item["plan"],
		"created_at": item["created_at"],
	}
}

func workflowRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uid":        item["uid"],
		"name":       item["name"],
		"status":     item["status"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func jobRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uid":          item["uid"],
		"workflow_uid": item["workflow_uid"],
		"status":       item["status"],
		"created_at":   item["created_at"],
		"ended_at":     item["ended_at"],
	}
}

func jobResultRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uid":        item["uid"],
		"job_uid":    item["job_uid"],
		"status":     item["status"],
		"data":       item["data"],
		"created_at": item["created_at"],
	}
}
