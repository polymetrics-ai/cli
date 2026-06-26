package gong

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Gong API resource path (relative to
// base_url), the cursor query param's start filter, and the record mapper that
// flattens its objects. All core streams use GET + cursor pagination where the
// next-page token lives at records.cursor in the response body and is sent back
// as the "cursor" query parameter.
type streamEndpoint struct {
	// resource is the Gong list endpoint path segment (e.g. "users", "calls",
	// "settings/scorecards").
	resource string
	// recordsPath is the JSON path to the array of records in the response body
	// (e.g. "users", "calls", "scorecards").
	recordsPath string
	// startParam, when non-empty, is the query parameter that carries the
	// start_date lower bound (e.g. "fromDateTime").
	startParam string
	// mapRecord flattens a raw Gong object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// gongStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in gongStreams; the read path
// is fully data-driven from this table.
var gongStreamEndpoints = map[string]streamEndpoint{
	"users":      {resource: "users", recordsPath: "users", startParam: "fromDateTime", mapRecord: gongUserRecord},
	"calls":      {resource: "calls", recordsPath: "calls", startParam: "fromDateTime", mapRecord: gongCallRecord},
	"scorecards": {resource: "settings/scorecards", recordsPath: "scorecards", startParam: "fromDateTime", mapRecord: gongScorecardRecord},
}

// gongStreams returns the connector's published stream catalog.
func gongStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "users",
			Description:  "Gong users (profiles, settings, manager hierarchy).",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       gongUserFields(),
		},
		{
			Name:         "calls",
			Description:  "Gong calls metadata (participants, duration, timestamps).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"started"},
			Fields:       gongCallFields(),
		},
		{
			Name:         "scorecards",
			Description:  "Gong scorecard definitions and configurations.",
			PrimaryKey:   []string{"scorecardId"},
			CursorFields: nil,
			Fields:       gongScorecardFields(),
		},
	}
}

func gongUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email_address", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "phone_number", Type: "string"},
		{Name: "manager_id", Type: "string"},
		{Name: "created", Type: "string"},
	}
}

func gongCallFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "started", Type: "string"},
		{Name: "scheduled", Type: "string"},
		{Name: "duration", Type: "integer"},
		{Name: "direction", Type: "string"},
		{Name: "system", Type: "string"},
		{Name: "scope", Type: "string"},
		{Name: "media", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "is_private", Type: "boolean"},
	}
}

func gongScorecardFields() []connectors.Field {
	return []connectors.Field{
		{Name: "scorecardId", Type: "string"},
		{Name: "scorecardName", Type: "string"},
		{Name: "workspaceId", Type: "string"},
		{Name: "enabled", Type: "boolean"},
		{Name: "created", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func gongUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"email_address": item["emailAddress"],
		"first_name":    item["firstName"],
		"last_name":     item["lastName"],
		"title":         item["title"],
		"active":        item["active"],
		"phone_number":  item["phoneNumber"],
		"manager_id":    item["managerId"],
		"created":       item["created"],
	}
}

func gongCallRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"title":      item["title"],
		"started":    item["started"],
		"scheduled":  item["scheduled"],
		"duration":   item["duration"],
		"direction":  item["direction"],
		"system":     item["system"],
		"scope":      item["scope"],
		"media":      item["media"],
		"language":   item["language"],
		"url":        item["url"],
		"is_private": item["isPrivate"],
	}
}

func gongScorecardRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"scorecardId":   item["scorecardId"],
		"scorecardName": item["scorecardName"],
		"workspaceId":   item["workspaceId"],
		"enabled":       item["enabled"],
		"created":       item["created"],
		"updated":       item["updated"],
	}
}
