package chift

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Chift API resource path (relative to
// base_url) it reads from, the record mapper that flattens its objects, and the
// JSON path where the records array lives in the response.
//
// Chift list endpoints return a top-level JSON array, so recordsPath is ""
// (RecordsAt treats the empty path as the document root).
type streamEndpoint struct {
	// resource is the Chift list endpoint path segment (e.g. "consumers").
	resource string
	// recordsPath is the dotted path to the records array; "" means root.
	recordsPath string
	// mapRecord flattens a raw Chift object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// chiftStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in chiftStreams; the read path
// is fully data-driven from this table.
var chiftStreamEndpoints = map[string]streamEndpoint{
	"consumers":   {resource: "consumers", mapRecord: chiftConsumerRecord},
	"connections": {resource: "connections", mapRecord: chiftConnectionRecord},
	"syncs":       {resource: "syncs", mapRecord: chiftSyncRecord},
}

// chiftStreams returns the connector's published stream catalog. Chift exposes
// only full-refresh streams (no incremental cursor), so CursorFields is empty.
func chiftStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "consumers",
			Description: "Chift consumers: the end customers whose financial systems are connected.",
			PrimaryKey:  []string{"consumerid"},
			Fields:      chiftConsumerFields(),
		},
		{
			Name:        "connections",
			Description: "Chift connections: integrations linking a consumer to an external financial system.",
			PrimaryKey:  []string{"connectionid"},
			Fields:      chiftConnectionFields(),
		},
		{
			Name:        "syncs",
			Description: "Chift syncs: configured data synchronizations between connected systems.",
			PrimaryKey:  []string{"syncid"},
			Fields:      chiftSyncFields(),
		},
	}
}

func chiftConsumerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "consumerid", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "redirect_url", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "created_on", Type: "string"},
	}
}

func chiftConnectionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "connectionid", Type: "string"},
		{Name: "consumerid", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "api", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_on", Type: "string"},
	}
}

func chiftSyncFields() []connectors.Field {
	return []connectors.Field{
		{Name: "syncid", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "consumerid", Type: "string"},
		{Name: "created_on", Type: "string"},
		{Name: "updated_on", Type: "string"},
	}
}

func chiftConsumerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"consumerid":   item["consumerid"],
		"name":         item["name"],
		"email":        item["email"],
		"phone":        item["phone"],
		"redirect_url": item["redirect_url"],
		"active":       item["active"],
		"created_on":   item["created_on"],
	}
}

func chiftConnectionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"connectionid": item["connectionid"],
		"consumerid":   item["consumerid"],
		"name":         item["name"],
		"api":          item["api"],
		"status":       item["status"],
		"created_on":   item["created_on"],
	}
}

func chiftSyncRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"syncid":     item["syncid"],
		"name":       item["name"],
		"status":     item["status"],
		"consumerid": item["consumerid"],
		"created_on": item["created_on"],
		"updated_on": item["updated_on"],
	}
}
