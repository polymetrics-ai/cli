package linear

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Linear GraphQL connection it reads.
//
// Linear is a single-endpoint GraphQL API: every list query targets a "root
// connection" (issues, teams, projects, users, ...) shaped as
// { <connection>: { nodes: [...], pageInfo: { hasNextPage, endCursor } } }.
// connection is the root field name; selection is the GraphQL field selection
// for each node; mapRecord flattens a node into a connectors.Record.
type streamEndpoint struct {
	connection string
	selection  string
	mapRecord  func(map[string]any) connectors.Record
}

// linearStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in linearStreams; the read path
// is fully data-driven from this table.
var linearStreamEndpoints = map[string]streamEndpoint{
	"issues":   {connection: "issues", selection: issueSelection, mapRecord: linearIssueRecord},
	"teams":    {connection: "teams", selection: teamSelection, mapRecord: linearTeamRecord},
	"projects": {connection: "projects", selection: projectSelection, mapRecord: linearProjectRecord},
	"users":    {connection: "users", selection: userSelection, mapRecord: linearUserRecord},
}

const issueSelection = `id
identifier
title
description
priority
estimate
url
branchName
createdAt
updatedAt
completedAt
canceledAt
state { id name type }
team { id key name }
assignee { id name email }
creator { id name email }`

const teamSelection = `id
key
name
description
private
createdAt
updatedAt`

const projectSelection = `id
name
description
state
progress
url
createdAt
updatedAt
startedAt
completedAt
canceledAt`

const userSelection = `id
name
displayName
email
active
admin
createdAt
updatedAt`

// linearStreams returns the connector's published stream catalog. Every Linear
// node exposes a string id and ISO-8601 createdAt/updatedAt timestamps, so the
// primary key is ["id"] and the incremental cursor field is ["updatedAt"].
func linearStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "issues",
			Description:  "Linear issues.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       linearIssueFields(),
		},
		{
			Name:         "teams",
			Description:  "Linear teams.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       linearTeamFields(),
		},
		{
			Name:         "projects",
			Description:  "Linear projects.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       linearProjectFields(),
		},
		{
			Name:         "users",
			Description:  "Linear workspace users.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       linearUserFields(),
		},
	}
}

func linearIssueFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "identifier", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "priority", Type: "integer"},
		{Name: "estimate", Type: "number"},
		{Name: "url", Type: "string"},
		{Name: "branch_name", Type: "string"},
		{Name: "state_id", Type: "string"},
		{Name: "state_name", Type: "string"},
		{Name: "state_type", Type: "string"},
		{Name: "team_id", Type: "string"},
		{Name: "team_key", Type: "string"},
		{Name: "assignee_id", Type: "string"},
		{Name: "assignee_email", Type: "string"},
		{Name: "creator_id", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "completed_at", Type: "timestamp"},
		{Name: "canceled_at", Type: "timestamp"},
	}
}

func linearTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "key", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "private", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func linearProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "progress", Type: "number"},
		{Name: "url", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "started_at", Type: "timestamp"},
		{Name: "completed_at", Type: "timestamp"},
		{Name: "canceled_at", Type: "timestamp"},
	}
}

func linearUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "admin", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

// linearIssueRecord flattens an issue node, lifting nested state/team/assignee
// objects into prefixed scalar columns. The raw `identifier`, `createdAt`, and
// `updatedAt` keys are preserved so tests and cursor logic can read them
// directly.
func linearIssueRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":           item["id"],
		"identifier":   item["identifier"],
		"title":        item["title"],
		"description":  item["description"],
		"priority":     item["priority"],
		"estimate":     item["estimate"],
		"url":          item["url"],
		"branch_name":  item["branchName"],
		"created_at":   item["createdAt"],
		"updated_at":   item["updatedAt"],
		"completed_at": item["completedAt"],
		"canceled_at":  item["canceledAt"],
		// Preserve raw cursor/identifier keys for state + assertions.
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
	}
	if state := nestedObject(item, "state"); state != nil {
		rec["state_id"] = state["id"]
		rec["state_name"] = state["name"]
		rec["state_type"] = state["type"]
	}
	if team := nestedObject(item, "team"); team != nil {
		rec["team_id"] = team["id"]
		rec["team_key"] = team["key"]
	}
	if assignee := nestedObject(item, "assignee"); assignee != nil {
		rec["assignee_id"] = assignee["id"]
		rec["assignee_email"] = assignee["email"]
	}
	if creator := nestedObject(item, "creator"); creator != nil {
		rec["creator_id"] = creator["id"]
	}
	return rec
}

func linearTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"key":         item["key"],
		"name":        item["name"],
		"description": item["description"],
		"private":     item["private"],
		"created_at":  item["createdAt"],
		"updated_at":  item["updatedAt"],
		"createdAt":   item["createdAt"],
		"updatedAt":   item["updatedAt"],
	}
}

func linearProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"description":  item["description"],
		"state":        item["state"],
		"progress":     item["progress"],
		"url":          item["url"],
		"created_at":   item["createdAt"],
		"updated_at":   item["updatedAt"],
		"started_at":   item["startedAt"],
		"completed_at": item["completedAt"],
		"canceled_at":  item["canceledAt"],
		"createdAt":    item["createdAt"],
		"updatedAt":    item["updatedAt"],
	}
}

func linearUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"display_name": item["displayName"],
		"email":        item["email"],
		"active":       item["active"],
		"admin":        item["admin"],
		"created_at":   item["createdAt"],
		"updated_at":   item["updatedAt"],
		"createdAt":    item["createdAt"],
		"updatedAt":    item["updatedAt"],
	}
}

// nestedObject returns item[key] as a map when it is a JSON object, else nil.
func nestedObject(item map[string]any, key string) map[string]any {
	if obj, ok := item[key].(map[string]any); ok {
		return obj
	}
	return nil
}
