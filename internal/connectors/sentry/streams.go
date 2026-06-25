package sentry

import "polymetrics.ai/internal/connectors"

// scope describes the path-template scope a Sentry endpoint requires.
type scope int

const (
	// scopeGlobal endpoints take no org/project path segments (e.g. /projects/).
	scopeGlobal scope = iota
	// scopeOrg endpoints take the organization slug (e.g. /organizations/{org}/releases/).
	scopeOrg
	// scopeProject endpoints take both org and project slugs.
	scopeProject
)

// streamEndpoint maps a stream name to its Sentry API resource and the record
// mapper that flattens its objects. The read path is fully data-driven from the
// stream routing table below.
type streamEndpoint struct {
	// scope selects which path template the endpoint uses.
	scope scope
	// path is a printf template applied with org/project slugs as needed.
	// scopeGlobal: literal path. scopeOrg: one %s (org). scopeProject: two %s
	// (org, project).
	path string
	// mapRecord flattens a raw Sentry object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// sentryStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in sentryStreams.
var sentryStreamEndpoints = map[string]streamEndpoint{
	"projects": {scope: scopeGlobal, path: "projects/", mapRecord: sentryProjectRecord},
	"issues":   {scope: scopeProject, path: "projects/%s/%s/issues/", mapRecord: sentryIssueRecord},
	"events":   {scope: scopeProject, path: "projects/%s/%s/events/", mapRecord: sentryEventRecord},
	"releases": {scope: scopeOrg, path: "organizations/%s/releases/", mapRecord: sentryReleaseRecord},
}

// sentryStreams returns the connector's published stream catalog. Sentry list
// endpoints return top-level JSON arrays of objects, each with a string id.
func sentryStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "projects",
			Description:  "Sentry projects accessible to the auth token.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"dateCreated"},
			Fields:       sentryProjectFields(),
		},
		{
			Name:         "issues",
			Description:  "Sentry issues (groups) for the configured project.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"lastSeen"},
			Fields:       sentryIssueFields(),
		},
		{
			Name:         "events",
			Description:  "Sentry error events for the configured project (last 90 days).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"dateCreated"},
			Fields:       sentryEventFields(),
		},
		{
			Name:         "releases",
			Description:  "Sentry releases for the configured organization.",
			PrimaryKey:   []string{"version"},
			CursorFields: []string{"dateCreated"},
			Fields:       sentryReleaseFields(),
		},
	}
}

func sentryProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "platform", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "dateCreated", Type: "string"},
		{Name: "isPublic", Type: "boolean"},
		{Name: "isBookmarked", Type: "boolean"},
	}
}

func sentryIssueFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "shortId", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "culprit", Type: "string"},
		{Name: "level", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "count", Type: "string"},
		{Name: "userCount", Type: "integer"},
		{Name: "firstSeen", Type: "string"},
		{Name: "lastSeen", Type: "string"},
	}
}

func sentryEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "eventID", Type: "string"},
		{Name: "groupID", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "message", Type: "string"},
		{Name: "platform", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "dateCreated", Type: "string"},
	}
}

func sentryReleaseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "version", Type: "string"},
		{Name: "shortVersion", Type: "string"},
		{Name: "ref", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "dateCreated", Type: "string"},
		{Name: "dateReleased", Type: "string"},
	}
}

func sentryProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"slug":         item["slug"],
		"name":         item["name"],
		"platform":     item["platform"],
		"status":       item["status"],
		"dateCreated":  item["dateCreated"],
		"isPublic":     item["isPublic"],
		"isBookmarked": item["isBookmarked"],
	}
}

func sentryIssueRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"shortId":   item["shortId"],
		"title":     item["title"],
		"culprit":   item["culprit"],
		"level":     item["level"],
		"status":    item["status"],
		"type":      item["type"],
		"count":     item["count"],
		"userCount": item["userCount"],
		"firstSeen": item["firstSeen"],
		"lastSeen":  item["lastSeen"],
	}
}

func sentryEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"eventID":     item["eventID"],
		"groupID":     item["groupID"],
		"title":       item["title"],
		"message":     item["message"],
		"platform":    item["platform"],
		"type":        item["type"],
		"dateCreated": item["dateCreated"],
	}
}

func sentryReleaseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"version":      item["version"],
		"shortVersion": item["shortVersion"],
		"ref":          item["ref"],
		"url":          item["url"],
		"status":       item["status"],
		"dateCreated":  item["dateCreated"],
		"dateReleased": item["dateReleased"],
	}
}
