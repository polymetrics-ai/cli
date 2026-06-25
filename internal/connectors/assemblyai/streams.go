package assemblyai

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the AssemblyAI API resource path
// (relative to base_url), the JSON path to the records array in the response,
// and the record mapper that flattens its objects. The read path is fully
// data-driven from the assemblyaiStreamEndpoints table.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "v2/transcript").
	resource string
	// recordsPath is the dotted JSON path to the records array (e.g.
	// "transcripts"). Empty selects the root object as a single record.
	recordsPath string
	// mapRecord flattens a raw API object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// assemblyaiStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in assemblyaiStreams.
var assemblyaiStreamEndpoints = map[string]streamEndpoint{
	"transcript":            {resource: "v2/transcript", recordsPath: "transcripts", mapRecord: transcriptRecord},
	"transcript_sentences":  {resource: "v2/transcript", recordsPath: "transcripts", mapRecord: transcriptRefRecord},
	"transcript_paragraphs": {resource: "v2/transcript", recordsPath: "transcripts", mapRecord: transcriptRefRecord},
	"transcript_subtitle":   {resource: "v2/transcript", recordsPath: "transcripts", mapRecord: transcriptRefRecord},
}

// assemblyaiStreams returns the connector's published stream catalog. Every
// AssemblyAI transcript exposes a UUID `id` and a `created` RFC3339 timestamp,
// so the primary key is ["id"] and the incremental cursor is ["created"].
func assemblyaiStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "transcript",
			Description:  "AssemblyAI transcripts (the full transcript list with status and timestamps).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       transcriptFields(),
		},
		{
			Name:         "transcript_sentences",
			Description:  "Transcript references for retrieving per-transcript sentences.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       transcriptRefFields(),
		},
		{
			Name:         "transcript_paragraphs",
			Description:  "Transcript references for retrieving per-transcript paragraphs.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       transcriptRefFields(),
		},
		{
			Name:         "transcript_subtitle",
			Description:  "Transcript references for retrieving per-transcript subtitles (srt/vtt).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       transcriptRefFields(),
		},
	}
}

func transcriptFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "completed", Type: "string"},
		{Name: "audio_url", Type: "string"},
		{Name: "resource_url", Type: "string"},
		{Name: "error", Type: "string"},
	}
}

func transcriptRefFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "resource_url", Type: "string"},
	}
}

func transcriptRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"status":       item["status"],
		"created":      item["created"],
		"completed":    item["completed"],
		"audio_url":    item["audio_url"],
		"resource_url": item["resource_url"],
		"error":        item["error"],
	}
}

func transcriptRefRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"status":       item["status"],
		"created":      item["created"],
		"resource_url": item["resource_url"],
	}
}
