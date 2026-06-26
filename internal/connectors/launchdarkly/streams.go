package launchdarkly

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the LaunchDarkly API resource path
// (relative to base_url) it reads from, the record mapper that flattens its
// objects, and whether the path requires the project_key config to be templated
// in. Adding a stream means adding one entry here plus a Stream definition in
// launchdarklyStreams; the read path is fully data-driven from this table.
type streamEndpoint struct {
	// resource is the path template. "{project_key}" is substituted with the
	// project_key config value when needsProject is true.
	resource string
	// needsProject indicates the path embeds the project_key config.
	needsProject bool
	// primaryKey is the field carrying the record's identity.
	primaryKey string
	// mapRecord flattens a raw LaunchDarkly object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// launchdarklyStreamEndpoints is the per-stream routing table.
var launchdarklyStreamEndpoints = map[string]streamEndpoint{
	"projects":     {resource: "projects", primaryKey: "_id", mapRecord: projectRecord},
	"members":      {resource: "members", primaryKey: "_id", mapRecord: memberRecord},
	"auditlog":     {resource: "auditlog", primaryKey: "_id", mapRecord: auditLogRecord},
	"flags":        {resource: "flags/{project_key}", needsProject: true, primaryKey: "key", mapRecord: flagRecord},
	"environments": {resource: "projects/{project_key}/environments", needsProject: true, primaryKey: "_id", mapRecord: environmentRecord},
}

// launchdarklyStreams returns the connector's published stream catalog. All
// LaunchDarkly list responses wrap their rows under the "items" key.
func launchdarklyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "projects",
			Description: "LaunchDarkly projects.",
			PrimaryKey:  []string{"_id"},
			Fields:      projectFields(),
		},
		{
			Name:        "members",
			Description: "LaunchDarkly account members.",
			PrimaryKey:  []string{"_id"},
			Fields:      memberFields(),
		},
		{
			Name:         "auditlog",
			Description:  "LaunchDarkly audit log entries.",
			PrimaryKey:   []string{"_id"},
			CursorFields: []string{"date"},
			Fields:       auditLogFields(),
		},
		{
			Name:        "flags",
			Description: "LaunchDarkly feature flags for the configured project_key.",
			PrimaryKey:  []string{"key"},
			Fields:      flagFields(),
		},
		{
			Name:        "environments",
			Description: "LaunchDarkly environments for the configured project_key.",
			PrimaryKey:  []string{"_id"},
			Fields:      environmentFields(),
		},
	}
}

func projectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "key", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "tags", Type: "array"},
	}
}

func memberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "_pendingInvite", Type: "boolean"},
	}
}

func auditLogFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "date", Type: "integer"},
		{Name: "kind", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "shortDescription", Type: "string"},
	}
}

func flagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "key", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "kind", Type: "string"},
		{Name: "creationDate", Type: "integer"},
		{Name: "temporary", Type: "boolean"},
		{Name: "tags", Type: "array"},
	}
}

func environmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "key", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "defaultTtl", Type: "integer"},
		{Name: "tags", Type: "array"},
	}
}

func projectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":  item["_id"],
		"key":  item["key"],
		"name": item["name"],
		"tags": item["tags"],
	}
}

func memberRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":            item["_id"],
		"email":          item["email"],
		"firstName":      item["firstName"],
		"lastName":       item["lastName"],
		"role":           item["role"],
		"_pendingInvite": item["_pendingInvite"],
	}
}

func auditLogRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":              item["_id"],
		"date":             item["date"],
		"kind":             item["kind"],
		"name":             item["name"],
		"description":      item["description"],
		"shortDescription": item["shortDescription"],
	}
}

func flagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"key":          item["key"],
		"name":         item["name"],
		"description":  item["description"],
		"kind":         item["kind"],
		"creationDate": item["creationDate"],
		"temporary":    item["temporary"],
		"tags":         item["tags"],
	}
}

func environmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":        item["_id"],
		"key":        item["key"],
		"name":       item["name"],
		"color":      item["color"],
		"defaultTtl": item["defaultTtl"],
		"tags":       item["tags"],
	}
}
