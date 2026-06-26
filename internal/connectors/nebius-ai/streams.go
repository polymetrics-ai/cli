package nebiusai

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Nebius AI Studio API resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects.
type streamEndpoint struct {
	// resource is the list endpoint path (e.g. "v1/models").
	resource string
	// mapRecord flattens a raw Nebius object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// nebiusStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in nebiusStreams; the read path
// is fully data-driven from this table. The Nebius AI Studio API is
// OpenAI-compatible: list endpoints return {object:"list", data:[...]}.
var nebiusStreamEndpoints = map[string]streamEndpoint{
	"models":  {resource: "v1/models", mapRecord: nebiusModelRecord},
	"files":   {resource: "v1/files", mapRecord: nebiusFileRecord},
	"batches": {resource: "v1/batches", mapRecord: nebiusBatchRecord},
}

// nebiusStreams returns the connector's published stream catalog. Every Nebius
// object exposes a string id, so the primary key is ["id"] across the board.
func nebiusStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "models",
			Description:  "Nebius AI Studio models available to the account.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       nebiusModelFields(),
		},
		{
			Name:         "files",
			Description:  "Files uploaded to Nebius AI Studio (e.g. batch inputs/outputs).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       nebiusFileFields(),
		},
		{
			Name:         "batches",
			Description:  "Nebius AI Studio batch jobs.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       nebiusBatchFields(),
		},
	}
}

func nebiusModelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "created", Type: "integer"},
		{Name: "owned_by", Type: "string"},
	}
}

func nebiusFileFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "filename", Type: "string"},
		{Name: "bytes", Type: "integer"},
		{Name: "purpose", Type: "string"},
		{Name: "status", Type: "string"},
	}
}

func nebiusBatchFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "endpoint", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "completed_at", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "input_file_id", Type: "string"},
		{Name: "output_file_id", Type: "string"},
		{Name: "error_file_id", Type: "string"},
	}
}

func nebiusModelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"object":   item["object"],
		"created":  item["created"],
		"owned_by": item["owned_by"],
	}
}

func nebiusFileRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"object":     item["object"],
		"created_at": item["created_at"],
		"filename":   item["filename"],
		"bytes":      item["bytes"],
		"purpose":    item["purpose"],
		"status":     item["status"],
	}
}

func nebiusBatchRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"object":         item["object"],
		"endpoint":       item["endpoint"],
		"created_at":     item["created_at"],
		"completed_at":   item["completed_at"],
		"status":         item["status"],
		"input_file_id":  item["input_file_id"],
		"output_file_id": item["output_file_id"],
		"error_file_id":  item["error_file_id"],
	}
}
