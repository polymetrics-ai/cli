package mux

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Mux API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Mux list endpoint path (e.g. "video/v1/assets").
	resource string
	// mapRecord flattens a raw Mux object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// muxStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in muxStreams; the read path is
// fully data-driven from this table.
var muxStreamEndpoints = map[string]streamEndpoint{
	"assets":       {resource: "video/v1/assets", mapRecord: muxAssetRecord},
	"live_streams": {resource: "video/v1/live-streams", mapRecord: muxLiveStreamRecord},
	"uploads":      {resource: "video/v1/uploads", mapRecord: muxUploadRecord},
	"signing_keys": {resource: "system/v1/signing-keys", mapRecord: muxSigningKeyRecord},
}

// muxStreams returns the connector's published stream catalog. Every Mux object
// exposes a string id and (where present) a `created_at` unix-seconds string, so
// the primary key is ["id"] across the board.
func muxStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "assets",
			Description:  "Mux Video assets.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       muxAssetFields(),
		},
		{
			Name:         "live_streams",
			Description:  "Mux Video live streams.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       muxLiveStreamFields(),
		},
		{
			Name:        "uploads",
			Description: "Mux Video direct uploads.",
			PrimaryKey:  []string{"id"},
			Fields:      muxUploadFields(),
		},
		{
			Name:         "signing_keys",
			Description:  "Mux system signing keys.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       muxSigningKeyFields(),
		},
	}
}

func muxAssetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "duration", Type: "number"},
		{Name: "max_resolution_tier", Type: "string"},
		{Name: "encoding_tier", Type: "string"},
		{Name: "mp4_support", Type: "string"},
		{Name: "master_access", Type: "string"},
		{Name: "test", Type: "boolean"},
	}
}

func muxLiveStreamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "stream_key", Type: "string"},
		{Name: "latency_mode", Type: "string"},
		{Name: "reconnect_window", Type: "number"},
		{Name: "max_continuous_duration", Type: "number"},
		{Name: "test", Type: "boolean"},
	}
}

func muxUploadFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "asset_id", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "timeout", Type: "number"},
		{Name: "cors_origin", Type: "string"},
		{Name: "test", Type: "boolean"},
	}
}

func muxSigningKeyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func muxAssetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"status":              item["status"],
		"created_at":          item["created_at"],
		"duration":            item["duration"],
		"max_resolution_tier": item["max_resolution_tier"],
		"encoding_tier":       item["encoding_tier"],
		"mp4_support":         item["mp4_support"],
		"master_access":       item["master_access"],
		"test":                item["test"],
	}
}

func muxLiveStreamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                      item["id"],
		"status":                  item["status"],
		"created_at":              item["created_at"],
		"stream_key":              item["stream_key"],
		"latency_mode":            item["latency_mode"],
		"reconnect_window":        item["reconnect_window"],
		"max_continuous_duration": item["max_continuous_duration"],
		"test":                    item["test"],
	}
}

func muxUploadRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"status":      item["status"],
		"asset_id":    item["asset_id"],
		"url":         item["url"],
		"timeout":     item["timeout"],
		"cors_origin": item["cors_origin"],
		"test":        item["test"],
	}
}

func muxSigningKeyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"created_at": item["created_at"],
	}
}
