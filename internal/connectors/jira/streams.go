package jira

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Jira REST API resource path it reads
// from, the JSON path at which its record array lives in the response, and the
// record mapper that flattens its objects into connectors.Record.
//
// Jira paginated responses are uniform offset envelopes
// ({startAt, maxResults, total, <records>[]}); the records array key differs by
// endpoint (issues -> "issues", project/search -> "values"). The users endpoint
// returns a bare top-level array, expressed here with recordsPath "".
type streamEndpoint struct {
	// resource is the path under /rest/api/3 (e.g. "search").
	resource string
	// recordsPath is the dotted JSON path to the records array in the response
	// body. Empty selects the root (a top-level array).
	recordsPath string
	// mapRecord flattens a raw Jira object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// jiraStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in jiraStreams; the read path
// is fully data-driven from this table.
var jiraStreamEndpoints = map[string]streamEndpoint{
	"issues":   {resource: "search", recordsPath: "issues", mapRecord: jiraIssueRecord},
	"projects": {resource: "project/search", recordsPath: "values", mapRecord: jiraProjectRecord},
	"users":    {resource: "users/search", recordsPath: "", mapRecord: jiraUserRecord},
}

// jiraStreams returns the connector's published stream catalog (the core set:
// issues, projects, users).
func jiraStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "issues",
			Description:  "Jira issues from GET /rest/api/3/search.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       jiraIssueFields(),
		},
		{
			Name:        "projects",
			Description: "Jira projects from GET /rest/api/3/project/search.",
			PrimaryKey:  []string{"id"},
			Fields:      jiraProjectFields(),
		},
		{
			Name:        "users",
			Description: "Jira users from GET /rest/api/3/users/search.",
			PrimaryKey:  []string{"accountId"},
			Fields:      jiraUserFields(),
		},
	}
}

func jiraIssueFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "key", Type: "string"},
		{Name: "self", Type: "string"},
		{Name: "summary", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "issuetype", Type: "string"},
		{Name: "priority", Type: "string"},
		{Name: "assignee", Type: "string"},
		{Name: "reporter", Type: "string"},
		{Name: "project", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func jiraProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "key", Type: "string"},
		{Name: "self", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "projectTypeKey", Type: "string"},
		{Name: "simplified", Type: "boolean"},
		{Name: "style", Type: "string"},
		{Name: "isPrivate", Type: "boolean"},
	}
}

func jiraUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "accountId", Type: "string"},
		{Name: "accountType", Type: "string"},
		{Name: "self", Type: "string"},
		{Name: "displayName", Type: "string"},
		{Name: "emailAddress", Type: "string"},
		{Name: "active", Type: "boolean"},
	}
}

// jiraIssueRecord flattens an issue: top-level id/key/self plus a curated subset
// of the nested `fields` object lifted to the record root.
func jiraIssueRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":   item["id"],
		"key":  item["key"],
		"self": item["self"],
	}
	fields, _ := item["fields"].(map[string]any)
	rec["summary"] = mapGet(fields, "summary")
	rec["created"] = mapGet(fields, "created")
	rec["updated"] = mapGet(fields, "updated")
	rec["status"] = nestedName(fields, "status")
	rec["issuetype"] = nestedName(fields, "issuetype")
	rec["priority"] = nestedName(fields, "priority")
	rec["assignee"] = nestedDisplayName(fields, "assignee")
	rec["reporter"] = nestedDisplayName(fields, "reporter")
	rec["project"] = nestedKey(fields, "project")
	return rec
}

func jiraProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"key":            item["key"],
		"self":           item["self"],
		"name":           item["name"],
		"projectTypeKey": item["projectTypeKey"],
		"simplified":     item["simplified"],
		"style":          item["style"],
		"isPrivate":      item["isPrivate"],
	}
}

func jiraUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"accountId":    item["accountId"],
		"accountType":  item["accountType"],
		"self":         item["self"],
		"displayName":  item["displayName"],
		"emailAddress": item["emailAddress"],
		"active":       item["active"],
	}
}

// mapGet reads a key from a possibly-nil map.
func mapGet(m map[string]any, key string) any {
	if m == nil {
		return nil
	}
	return m[key]
}

// nestedName lifts the "name" of a nested object field (e.g. status.name).
func nestedName(fields map[string]any, key string) any {
	if obj, ok := mapGet(fields, key).(map[string]any); ok {
		return obj["name"]
	}
	return nil
}

// nestedDisplayName lifts the "displayName" of a nested user field.
func nestedDisplayName(fields map[string]any, key string) any {
	if obj, ok := mapGet(fields, key).(map[string]any); ok {
		return obj["displayName"]
	}
	return nil
}

// nestedKey lifts the "key" of a nested object field (e.g. project.key).
func nestedKey(fields map[string]any, key string) any {
	if obj, ok := mapGet(fields, key).(map[string]any); ok {
		return obj["key"]
	}
	return nil
}
