package gitlab

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the GitLab API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects. Every
// GitLab collection endpoint returns a top-level JSON array, so the records path
// is the root ("") in all cases.
type streamEndpoint struct {
	// resource is the GitLab list endpoint path segment (e.g. "projects").
	resource string
	// mapRecord flattens a raw GitLab object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// gitlabStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in gitlabStreams; the read path
// is fully data-driven from this table. The chosen streams are instance-global
// collections that do not require a project/group context to list.
var gitlabStreamEndpoints = map[string]streamEndpoint{
	"projects": {resource: "projects", mapRecord: gitlabProjectRecord},
	"groups":   {resource: "groups", mapRecord: gitlabGroupRecord},
	"users":    {resource: "users", mapRecord: gitlabUserRecord},
	"issues":   {resource: "issues", mapRecord: gitlabIssueRecord},
}

// gitlabStreams returns the connector's published stream catalog. Every GitLab
// object exposes a numeric id and (for most) a created_at timestamp, so the
// primary key is ["id"] and the incremental cursor field is ["created_at"].
func gitlabStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "projects",
			Description:  "GitLab projects visible to the authenticated token.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_activity_at"},
			Fields:       gitlabProjectFields(),
		},
		{
			Name:         "groups",
			Description:  "GitLab groups visible to the authenticated token.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       gitlabGroupFields(),
		},
		{
			Name:         "users",
			Description:  "GitLab users visible to the authenticated token.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       gitlabUserFields(),
		},
		{
			Name:         "issues",
			Description:  "GitLab issues across projects visible to the authenticated token.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       gitlabIssueFields(),
		},
	}
}

func gitlabProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "path", Type: "string"},
		{Name: "path_with_namespace", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "default_branch", Type: "string"},
		{Name: "visibility", Type: "string"},
		{Name: "web_url", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "last_activity_at", Type: "timestamp"},
		{Name: "archived", Type: "boolean"},
		{Name: "star_count", Type: "integer"},
		{Name: "forks_count", Type: "integer"},
		{Name: "open_issues_count", Type: "integer"},
	}
}

func gitlabGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "path", Type: "string"},
		{Name: "full_path", Type: "string"},
		{Name: "full_name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "visibility", Type: "string"},
		{Name: "web_url", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "parent_id", Type: "integer"},
	}
}

func gitlabUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "username", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "web_url", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "bot", Type: "boolean"},
		{Name: "is_admin", Type: "boolean"},
	}
}

func gitlabIssueFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "iid", Type: "integer"},
		{Name: "project_id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "web_url", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "closed_at", Type: "timestamp"},
		{Name: "author_id", Type: "integer"},
		{Name: "upvotes", Type: "integer"},
		{Name: "downvotes", Type: "integer"},
		{Name: "user_notes_count", Type: "integer"},
	}
}

func gitlabProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"name":                item["name"],
		"path":                item["path"],
		"path_with_namespace": item["path_with_namespace"],
		"description":         item["description"],
		"default_branch":      item["default_branch"],
		"visibility":          item["visibility"],
		"web_url":             item["web_url"],
		"created_at":          item["created_at"],
		"last_activity_at":    item["last_activity_at"],
		"archived":            item["archived"],
		"star_count":          item["star_count"],
		"forks_count":         item["forks_count"],
		"open_issues_count":   item["open_issues_count"],
	}
}

func gitlabGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"path":        item["path"],
		"full_path":   item["full_path"],
		"full_name":   item["full_name"],
		"description": item["description"],
		"visibility":  item["visibility"],
		"web_url":     item["web_url"],
		"created_at":  item["created_at"],
		"parent_id":   item["parent_id"],
	}
}

func gitlabUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"username":   item["username"],
		"name":       item["name"],
		"state":      item["state"],
		"web_url":    item["web_url"],
		"created_at": item["created_at"],
		"bot":        item["bot"],
		"is_admin":   item["is_admin"],
	}
}

func gitlabIssueRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"iid":              item["iid"],
		"project_id":       item["project_id"],
		"title":            item["title"],
		"state":            item["state"],
		"web_url":          item["web_url"],
		"created_at":       item["created_at"],
		"updated_at":       item["updated_at"],
		"closed_at":        item["closed_at"],
		"author_id":        gitlabAuthorID(item["author"]),
		"upvotes":          item["upvotes"],
		"downvotes":        item["downvotes"],
		"user_notes_count": item["user_notes_count"],
	}
}

// gitlabAuthorID extracts the nested author.id from a GitLab issue object, which
// embeds the author as a sub-object rather than a flat id.
func gitlabAuthorID(author any) any {
	if obj, ok := author.(map[string]any); ok {
		return obj["id"]
	}
	return nil
}
