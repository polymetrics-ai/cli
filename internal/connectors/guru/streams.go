package guru

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Guru API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Guru list endpoint path segment (e.g. "collections").
	resource string
	// mapRecord flattens a raw Guru object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// guruStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in guruStreams; the read path
// is fully data-driven from this table. Guru list responses are top-level JSON
// arrays, so every stream extracts records at the root path ("").
var guruStreamEndpoints = map[string]streamEndpoint{
	"collections": {resource: "collections", mapRecord: guruCollectionRecord},
	"groups":      {resource: "groups", mapRecord: guruGroupRecord},
	"members":     {resource: "members", mapRecord: guruMemberRecord},
	"teams":       {resource: "teams", mapRecord: guruTeamRecord},
}

// guruStreams returns the connector's published stream catalog.
func guruStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "collections",
			Description: "Guru collections (top-level knowledge groupings).",
			PrimaryKey:  []string{"id"},
			Fields:      guruCollectionFields(),
		},
		{
			Name:        "groups",
			Description: "Guru user groups within a team.",
			PrimaryKey:  []string{"id"},
			Fields:      guruGroupFields(),
		},
		{
			Name:        "members",
			Description: "Guru team members (users).",
			PrimaryKey:  []string{"id"},
			Fields:      guruMemberFields(),
		},
		{
			Name:        "teams",
			Description: "Guru teams the authenticated user can access.",
			PrimaryKey:  []string{"id"},
			Fields:      guruTeamFields(),
		},
	}
}

func guruCollectionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "publicCardsEnabled", Type: "boolean"},
		{Name: "collectionType", Type: "string"},
		{Name: "dateCreated", Type: "string"},
	}
}

func guruGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "dateCreated", Type: "string"},
		{Name: "modifiable", Type: "boolean"},
		{Name: "groupType", Type: "string"},
		{Name: "memberCount", Type: "integer"},
	}
}

func guruMemberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "dateCreated", Type: "string"},
	}
}

func guruTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "dateCreated", Type: "string"},
	}
}

func guruCollectionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"name":               item["name"],
		"slug":               item["slug"],
		"description":        item["description"],
		"color":              item["color"],
		"publicCardsEnabled": item["publicCardsEnabled"],
		"collectionType":     item["collectionType"],
		"dateCreated":        item["dateCreated"],
	}
}

func guruGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"dateCreated": item["dateCreated"],
		"modifiable":  item["modifiable"],
		"groupType":   item["groupType"],
		"memberCount": item["memberCount"],
	}
}

// guruMemberRecord flattens a Guru member. The member's identity (id, email,
// names) is nested under a "user" object in the live API, so prefer those when
// present and fall back to top-level fields (used by fixture records).
func guruMemberRecord(item map[string]any) connectors.Record {
	user, _ := item["user"].(map[string]any)
	pick := func(key string) any {
		if user != nil {
			if v, ok := user[key]; ok {
				return v
			}
		}
		return item[key]
	}
	return connectors.Record{
		"id":          pick("id"),
		"email":       pick("email"),
		"firstName":   pick("firstName"),
		"lastName":    pick("lastName"),
		"status":      item["status"],
		"dateCreated": item["dateCreated"],
	}
}

func guruTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"status":      item["status"],
		"dateCreated": item["dateCreated"],
	}
}
