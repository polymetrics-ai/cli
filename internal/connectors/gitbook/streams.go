package gitbook

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the GitBook API resource it reads from,
// how its records are extracted, and the mapper that flattens each item.
type streamEndpoint struct {
	// resource is the path template relative to base_url. The {space} token is
	// substituted with the configured space_id at read time.
	resource string
	// recordsPath is the dotted JSON path to the records array in a list
	// response. Empty means the response body itself is the (single) record,
	// e.g. GET /user returns one user object.
	recordsPath string
	// list is true for paginated list endpoints (items[] + next.page) and false
	// for single-object endpoints (no pagination, one record).
	list bool
	// mapRecord flattens a raw GitBook object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// gitbookStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in gitbookStreams; the read
// path is fully data-driven from this table.
//
// GitBook list endpoints return {"items":[...],"next":{"page":"<cursor>"}} and
// accept ?page=<cursor>&limit=<n>. The /user endpoint returns a single object.
var gitbookStreamEndpoints = map[string]streamEndpoint{
	"users":         {resource: "user", recordsPath: "", list: false, mapRecord: gitbookUserRecord},
	"organizations": {resource: "orgs", recordsPath: "items", list: true, mapRecord: gitbookOrganizationRecord},
	"org_members":   {resource: "orgs/{org}/members", recordsPath: "items", list: true, mapRecord: gitbookOrgMemberRecord},
	"content":       {resource: "spaces/{space}/content/pages", recordsPath: "pages", list: true, mapRecord: gitbookContentRecord},
}

// gitbookStreams returns the connector's published stream catalog. GitBook
// objects are identified by a string id and are full-refresh only (the Airbyte
// source declares no incremental cursors), so CursorFields is empty.
func gitbookStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "users",
			Description: "The authenticated GitBook user.",
			PrimaryKey:  []string{"id"},
			Fields:      gitbookUserFields(),
		},
		{
			Name:        "organizations",
			Description: "Organizations the authenticated user belongs to.",
			PrimaryKey:  []string{"id"},
			Fields:      gitbookOrganizationFields(),
		},
		{
			Name:        "org_members",
			Description: "Members of the configured organization.",
			PrimaryKey:  []string{"id"},
			Fields:      gitbookOrgMemberFields(),
		},
		{
			Name:        "content",
			Description: "Pages (content tree) of the configured space.",
			PrimaryKey:  []string{"id"},
			Fields:      gitbookContentFields(),
		},
	}
}

func gitbookUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "photo_url", Type: "string"},
	}
}

func gitbookOrganizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "url", Type: "string"},
	}
}

func gitbookOrgMemberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "role", Type: "string"},
	}
}

func gitbookContentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "kind", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "path", Type: "string"},
	}
}

func gitbookUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"display_name": firstNonNil(item["displayName"], item["display_name"]),
		"email":        item["email"],
		"photo_url":    pickNested(item, "photoURL", "photo_url"),
	}
}

func gitbookOrganizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"title":      item["title"],
		"type":       item["type"],
		"created_at": firstNonNil(item["createdAt"], item["created_at"]),
		"url":        item["urls"],
	}
}

func gitbookOrgMemberRecord(item map[string]any) connectors.Record {
	// A member object nests the user under "user" on real payloads but may be
	// flattened in fixtures/tests; tolerate both.
	user, _ := item["user"].(map[string]any)
	id := item["id"]
	displayName := firstNonNil(item["displayName"], item["display_name"])
	email := item["email"]
	if user != nil {
		if id == nil {
			id = user["id"]
		}
		if displayName == nil {
			displayName = firstNonNil(user["displayName"], user["display_name"])
		}
		if email == nil {
			email = user["email"]
		}
	}
	return connectors.Record{
		"id":           id,
		"display_name": displayName,
		"email":        email,
		"role":         item["role"],
	}
}

func gitbookContentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":    item["id"],
		"title": item["title"],
		"type":  item["type"],
		"kind":  item["kind"],
		"slug":  item["slug"],
		"path":  item["path"],
	}
}

// firstNonNil returns the first non-nil value from the supplied candidates.
func firstNonNil(values ...any) any {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

// pickNested returns the first present key's value from item.
func pickNested(item map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := item[k]; ok && v != nil {
			return v
		}
	}
	return nil
}
