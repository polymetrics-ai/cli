package instatus

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to its Instatus API resource and record
// mapper. Endpoints are either top-level (e.g. "v2/pages") or parent-scoped: a
// scoped endpoint embeds the page id, so its path is built at read time from the
// page_id config via pathFor.
type streamEndpoint struct {
	// pathSegments is the API path template. For parent-scoped streams the
	// page id is inserted between version and resource, e.g. ["v2", "components"]
	// resolves to "v2/{page_id}/components".
	version string
	// resource is the trailing path segment (e.g. "components").
	resource string
	// scoped indicates the path is "/{version}/{page_id}/{resource}" and so
	// requires a page_id config value. Top-level streams ("v2/pages") set false.
	scoped bool
	// topLevelPath, when set, is the exact path used for non-scoped streams and
	// overrides version/resource composition (e.g. "v2/pages").
	topLevelPath string
	// mapRecord flattens a raw Instatus object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// instatusStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in instatusStreams. The Airbyte
// source exposes pages plus parent-scoped components/incidents/maintenances; we
// ship that core set.
var instatusStreamEndpoints = map[string]streamEndpoint{
	"pages":        {topLevelPath: "v2/pages", mapRecord: instatusPageRecord},
	"components":   {version: "v2", resource: "components", scoped: true, mapRecord: instatusComponentRecord},
	"incidents":    {version: "v1", resource: "incidents", scoped: true, mapRecord: instatusIncidentRecord},
	"maintenances": {version: "v2", resource: "maintenances", scoped: true, mapRecord: instatusMaintenanceRecord},
}

// instatusStreams returns the connector's published stream catalog. Instatus
// objects carry a string id; pages also carry createdAt/updatedAt timestamps,
// though the API only supports full-refresh syncs (no incremental cursor).
func instatusStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "pages",
			Description: "Instatus status pages owned by the workspace.",
			PrimaryKey:  []string{"id"},
			Fields:      instatusPageFields(),
		},
		{
			Name:        "components",
			Description: "Components of an Instatus status page (requires page_id config).",
			PrimaryKey:  []string{"id"},
			Fields:      instatusComponentFields(),
		},
		{
			Name:        "incidents",
			Description: "Incidents of an Instatus status page (requires page_id config).",
			PrimaryKey:  []string{"id"},
			Fields:      instatusIncidentFields(),
		},
		{
			Name:        "maintenances",
			Description: "Scheduled maintenances of an Instatus status page (requires page_id config).",
			PrimaryKey:  []string{"id"},
			Fields:      instatusMaintenanceFields(),
		},
	}
}

func instatusPageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "subdomain", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "websiteUrl", Type: "string"},
		{Name: "customDomain", Type: "string"},
		{Name: "publicEmail", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "createdAt", Type: "timestamp"},
		{Name: "updatedAt", Type: "timestamp"},
	}
}

func instatusComponentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "uniqueEmail", Type: "string"},
		{Name: "showUptime", Type: "boolean"},
		{Name: "order", Type: "integer"},
		{Name: "group", Type: "string"},
	}
}

func instatusIncidentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "started", Type: "timestamp"},
		{Name: "resolved", Type: "timestamp"},
		{Name: "createdAt", Type: "timestamp"},
		{Name: "updatedAt", Type: "timestamp"},
	}
}

func instatusMaintenanceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "start", Type: "timestamp"},
		{Name: "duration", Type: "integer"},
		{Name: "autoStart", Type: "boolean"},
		{Name: "autoEnd", Type: "boolean"},
	}
}

func instatusPageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"subdomain":    item["subdomain"],
		"name":         item["name"],
		"status":       item["status"],
		"websiteUrl":   item["websiteUrl"],
		"customDomain": item["customDomain"],
		"publicEmail":  item["publicEmail"],
		"language":     item["language"],
		"createdAt":    item["createdAt"],
		"updatedAt":    item["updatedAt"],
	}
}

func instatusComponentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"status":      item["status"],
		"description": item["description"],
		"uniqueEmail": item["uniqueEmail"],
		"showUptime":  item["showUptime"],
		"order":       item["order"],
		"group":       item["group"],
	}
}

func instatusIncidentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"status":    item["status"],
		"started":   item["started"],
		"resolved":  item["resolved"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
	}
}

func instatusMaintenanceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"status":    item["status"],
		"start":     item["start"],
		"duration":  item["duration"],
		"autoStart": item["autoStart"],
		"autoEnd":   item["autoEnd"],
	}
}
