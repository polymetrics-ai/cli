package airbyte

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Airbyte API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Airbyte list endpoint path segment (e.g. "connections").
	resource string
	// mapRecord flattens a raw Airbyte object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// airbyteStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in airbyteStreams; the read path
// is fully data-driven from this table.
var airbyteStreamEndpoints = map[string]streamEndpoint{
	"workspaces":   {resource: "workspaces", mapRecord: airbyteWorkspaceRecord},
	"connections":  {resource: "connections", mapRecord: airbyteConnectionRecord},
	"sources":      {resource: "sources", mapRecord: airbyteSourceRecord},
	"destinations": {resource: "destinations", mapRecord: airbyteDestinationRecord},
	"jobs":         {resource: "jobs", mapRecord: airbyteJobRecord},
}

// airbyteStreams returns the connector's published stream catalog. The Airbyte
// public API exposes UUID identifiers per resource; the primary key is the
// resource id. Only jobs carry an incremental timestamp, so it alone has a
// cursor field.
func airbyteStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "workspaces",
			Description: "Airbyte workspaces accessible to the application.",
			PrimaryKey:  []string{"workspaceId"},
			Fields:      airbyteWorkspaceFields(),
		},
		{
			Name:        "connections",
			Description: "Airbyte connections (source-to-destination syncs).",
			PrimaryKey:  []string{"connectionId"},
			Fields:      airbyteConnectionFields(),
		},
		{
			Name:        "sources",
			Description: "Airbyte configured sources.",
			PrimaryKey:  []string{"sourceId"},
			Fields:      airbyteSourceFields(),
		},
		{
			Name:        "destinations",
			Description: "Airbyte configured destinations.",
			PrimaryKey:  []string{"destinationId"},
			Fields:      airbyteDestinationFields(),
		},
		{
			Name:         "jobs",
			Description:  "Airbyte sync and reset jobs.",
			PrimaryKey:   []string{"jobId"},
			CursorFields: []string{"lastUpdatedAt"},
			Fields:       airbyteJobFields(),
		},
	}
}

func airbyteWorkspaceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "workspaceId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "dataResidency", Type: "string"},
	}
}

func airbyteConnectionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "connectionId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "sourceId", Type: "string"},
		{Name: "destinationId", Type: "string"},
		{Name: "workspaceId", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "schedule", Type: "object"},
		{Name: "dataResidency", Type: "string"},
		{Name: "namespaceFormat", Type: "string"},
		{Name: "prefix", Type: "string"},
	}
}

func airbyteSourceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "sourceId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "sourceType", Type: "string"},
		{Name: "workspaceId", Type: "string"},
		{Name: "definitionId", Type: "string"},
	}
}

func airbyteDestinationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "destinationId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "destinationType", Type: "string"},
		{Name: "workspaceId", Type: "string"},
		{Name: "definitionId", Type: "string"},
	}
}

func airbyteJobFields() []connectors.Field {
	return []connectors.Field{
		{Name: "jobId", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "jobType", Type: "string"},
		{Name: "connectionId", Type: "string"},
		{Name: "startTime", Type: "string"},
		{Name: "lastUpdatedAt", Type: "string"},
		{Name: "duration", Type: "string"},
		{Name: "bytesSynced", Type: "integer"},
		{Name: "rowsSynced", Type: "integer"},
	}
}

func airbyteWorkspaceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"workspaceId":   item["workspaceId"],
		"name":          item["name"],
		"dataResidency": item["dataResidency"],
	}
}

func airbyteConnectionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"connectionId":    item["connectionId"],
		"name":            item["name"],
		"sourceId":        item["sourceId"],
		"destinationId":   item["destinationId"],
		"workspaceId":     item["workspaceId"],
		"status":          item["status"],
		"schedule":        item["schedule"],
		"dataResidency":   item["dataResidency"],
		"namespaceFormat": item["namespaceFormat"],
		"prefix":          item["prefix"],
	}
}

func airbyteSourceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"sourceId":     item["sourceId"],
		"name":         item["name"],
		"sourceType":   item["sourceType"],
		"workspaceId":  item["workspaceId"],
		"definitionId": item["definitionId"],
	}
}

func airbyteDestinationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"destinationId":   item["destinationId"],
		"name":            item["name"],
		"destinationType": item["destinationType"],
		"workspaceId":     item["workspaceId"],
		"definitionId":    item["definitionId"],
	}
}

func airbyteJobRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"jobId":         item["jobId"],
		"status":        item["status"],
		"jobType":       item["jobType"],
		"connectionId":  item["connectionId"],
		"startTime":     item["startTime"],
		"lastUpdatedAt": item["lastUpdatedAt"],
		"duration":      item["duration"],
		"bytesSynced":   item["bytesSynced"],
		"rowsSynced":    item["rowsSynced"],
	}
}
