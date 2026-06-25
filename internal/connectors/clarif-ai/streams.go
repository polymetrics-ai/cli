package clarifai

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Clarifai API resource path (relative
// to the user-scoped base, i.e. "users/{user_id}/<resource>"), the JSON key the
// records array lives under in the response body, and the record mapper that
// flattens its objects.
type streamEndpoint struct {
	// resource is the path segment after users/{user_id}/ (e.g. "apps").
	resource string
	// recordsKey is the top-level JSON key holding the array (e.g. "apps").
	recordsKey string
	// mapRecord flattens a raw Clarifai object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// clarifaiStreamEndpoints is the per-stream routing table. Every Clarifai list
// endpoint is scoped under /users/{user_id}/ and returns its array under a
// resource-named key. Adding a stream means adding one entry here plus a Stream
// definition in clarifaiStreams; the read path is fully data-driven from this
// table.
var clarifaiStreamEndpoints = map[string]streamEndpoint{
	"applications":   {resource: "apps", recordsKey: "apps", mapRecord: clarifaiApplicationRecord},
	"datasets":       {resource: "datasets", recordsKey: "datasets", mapRecord: clarifaiDatasetRecord},
	"models":         {resource: "models", recordsKey: "models", mapRecord: clarifaiModelRecord},
	"model_versions": {resource: "models/versions", recordsKey: "model_versions", mapRecord: clarifaiModelVersionRecord},
	"workflows":      {resource: "workflows", recordsKey: "workflows", mapRecord: clarifaiWorkflowRecord},
}

// clarifaiStreams returns the connector's published stream catalog. Every
// Clarifai object exposes a string id and a created_at timestamp, so the primary
// key is ["id"] across the board. The Clarifai source is full-refresh only, so
// no cursor fields are published.
func clarifaiStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "applications",
			Description: "Clarifai applications owned by the user.",
			PrimaryKey:  []string{"id"},
			Fields:      clarifaiApplicationFields(),
		},
		{
			Name:        "datasets",
			Description: "Clarifai datasets across the user's applications.",
			PrimaryKey:  []string{"id"},
			Fields:      clarifaiDatasetFields(),
		},
		{
			Name:        "models",
			Description: "Clarifai models owned by the user.",
			PrimaryKey:  []string{"id"},
			Fields:      clarifaiModelFields(),
		},
		{
			Name:        "model_versions",
			Description: "Versions of the user's Clarifai models.",
			PrimaryKey:  []string{"id"},
			Fields:      clarifaiModelVersionFields(),
		},
		{
			Name:        "workflows",
			Description: "Clarifai workflows owned by the user.",
			PrimaryKey:  []string{"id"},
			Fields:      clarifaiWorkflowFields(),
		},
	}
}

func clarifaiApplicationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "default_language", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
	}
}

func clarifaiDatasetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "app_id", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "default_processing_info", Type: "object"},
		{Name: "created_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
	}
}

func clarifaiModelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "app_id", Type: "string"},
		{Name: "model_type_id", Type: "string"},
		{Name: "visibility", Type: "object"},
		{Name: "created_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
	}
}

func clarifaiModelVersionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "app_id", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "status", Type: "object"},
		{Name: "created_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
	}
}

func clarifaiWorkflowFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "app_id", Type: "string"},
		{Name: "version", Type: "object"},
		{Name: "created_at", Type: "string"},
		{Name: "modified_at", Type: "string"},
	}
}

func clarifaiApplicationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"user_id":          item["user_id"],
		"description":      item["description"],
		"default_language": item["default_language"],
		"created_at":       item["created_at"],
		"modified_at":      item["modified_at"],
	}
}

func clarifaiDatasetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                      item["id"],
		"user_id":                 item["user_id"],
		"app_id":                  item["app_id"],
		"description":             item["description"],
		"default_processing_info": item["default_processing_info"],
		"created_at":              item["created_at"],
		"modified_at":             item["modified_at"],
	}
}

func clarifaiModelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"user_id":       item["user_id"],
		"app_id":        item["app_id"],
		"model_type_id": item["model_type_id"],
		"visibility":    item["visibility"],
		"created_at":    item["created_at"],
		"modified_at":   item["modified_at"],
	}
}

func clarifaiModelVersionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"user_id":     item["user_id"],
		"app_id":      item["app_id"],
		"description": item["description"],
		"status":      item["status"],
		"created_at":  item["created_at"],
		"modified_at": item["modified_at"],
	}
}

func clarifaiWorkflowRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"user_id":     item["user_id"],
		"app_id":      item["app_id"],
		"version":     item["version"],
		"created_at":  item["created_at"],
		"modified_at": item["modified_at"],
	}
}
