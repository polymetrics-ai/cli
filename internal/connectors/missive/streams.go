package missive

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Missive API resource path (relative to
// base_url) it reads from, the JSON key holding the records array (which equals
// the resource name for Missive's list endpoints), and the record mapper.
type streamEndpoint struct {
	// resource is the Missive list endpoint path segment (e.g. "contacts").
	resource string
	// recordsPath is the dotted JSON path to the records array. Missive wraps
	// list responses as {"<resource>":[...]}, so this equals resource.
	recordsPath string
	// mapRecord flattens a raw Missive object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// missiveStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in missiveStreams; the read path
// is fully data-driven from this table.
var missiveStreamEndpoints = map[string]streamEndpoint{
	"contacts":       {resource: "contacts", recordsPath: "contacts", mapRecord: missiveContactRecord},
	"contact_groups": {resource: "contact_groups", recordsPath: "contact_groups", mapRecord: missiveContactGroupRecord},
	"users":          {resource: "users", recordsPath: "users", mapRecord: missiveUserRecord},
	"teams":          {resource: "teams", recordsPath: "teams", mapRecord: missiveTeamRecord},
	"shared_labels":  {resource: "shared_labels", recordsPath: "shared_labels", mapRecord: missiveSharedLabelRecord},
}

// missiveStreams returns the connector's published stream catalog. Every Missive
// object exposes a string id, so the primary key is ["id"] across the board.
// Missive's source is full-refresh only, so no cursor fields are published.
func missiveStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "contacts",
			Description: "Missive contacts.",
			PrimaryKey:  []string{"id"},
			Fields:      missiveContactFields(),
		},
		{
			Name:        "contact_groups",
			Description: "Missive contact groups (groups or organizations, selected by the kind config).",
			PrimaryKey:  []string{"id"},
			Fields:      missiveContactGroupFields(),
		},
		{
			Name:        "users",
			Description: "Missive organization users.",
			PrimaryKey:  []string{"id"},
			Fields:      missiveUserFields(),
		},
		{
			Name:        "teams",
			Description: "Missive teams.",
			PrimaryKey:  []string{"id"},
			Fields:      missiveTeamFields(),
		},
		{
			Name:        "shared_labels",
			Description: "Missive shared labels.",
			PrimaryKey:  []string{"id"},
			Fields:      missiveSharedLabelFields(),
		},
	}
}

func missiveContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "modified_at", Type: "integer"},
	}
}

func missiveContactGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "kind", Type: "string"},
	}
}

func missiveUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
	}
}

func missiveTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "organization", Type: "string"},
	}
}

func missiveSharedLabelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "name_with_parent_names", Type: "string"},
		{Name: "color", Type: "string"},
	}
}

func missiveContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"first_name":  item["first_name"],
		"last_name":   item["last_name"],
		"modified_at": item["modified_at"],
	}
}

func missiveContactGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   item["id"],
		"name": item["name"],
		"kind": item["kind"],
	}
}

func missiveUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":    item["id"],
		"name":  item["name"],
		"email": item["email"],
	}
}

func missiveTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"organization": item["organization"],
	}
}

func missiveSharedLabelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                     item["id"],
		"name":                   item["name"],
		"name_with_parent_names": item["name_with_parent_names"],
		"color":                  item["color"],
	}
}
