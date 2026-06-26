package huntr

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Huntr API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Huntr list endpoint path segment (e.g. "members").
	resource string
	// mapRecord flattens a raw Huntr object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// huntrStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in huntrStreams; the read path
// is fully data-driven from this table. Paths mirror the upstream Airbyte
// manifest (e.g. notes lives at /notes/members).
var huntrStreamEndpoints = map[string]streamEndpoint{
	"members":    {resource: "members", mapRecord: huntrMemberRecord},
	"candidates": {resource: "candidates", mapRecord: huntrCandidateRecord},
	"activities": {resource: "activities", mapRecord: huntrActivityRecord},
	"notes":      {resource: "notes/members", mapRecord: huntrNoteRecord},
	"actions":    {resource: "actions", mapRecord: huntrActionRecord},
}

// huntrStreams returns the connector's published stream catalog. Every Huntr
// object exposes a string id, so the primary key is ["id"]; the API only
// supports full_refresh, so no incremental cursor fields are declared.
func huntrStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "members",
			Description: "Huntr organization members.",
			PrimaryKey:  []string{"id"},
			Fields:      huntrMemberFields(),
		},
		{
			Name:        "candidates",
			Description: "Huntr candidates.",
			PrimaryKey:  []string{"id"},
			Fields:      huntrCandidateFields(),
		},
		{
			Name:        "activities",
			Description: "Huntr member activities.",
			PrimaryKey:  []string{"id"},
			Fields:      huntrActivityFields(),
		},
		{
			Name:        "notes",
			Description: "Huntr member notes.",
			PrimaryKey:  []string{"id"},
			Fields:      huntrNoteFields(),
		},
		{
			Name:        "actions",
			Description: "Huntr actions.",
			PrimaryKey:  []string{"id"},
			Fields:      huntrActionFields(),
		},
	}
}

func huntrMemberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "fullName", Type: "string"},
		{Name: "givenName", Type: "string"},
		{Name: "familyName", Type: "string"},
		{Name: "isActive", Type: "boolean"},
		{Name: "createdAt", Type: "number"},
		{Name: "lastSeenAt", Type: "number"},
		{Name: "boardIds", Type: "array"},
	}
}

func huntrCandidateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "memberId", Type: "string"},
	}
}

func huntrActivityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "activityCategory", Type: "string"},
		{Name: "completed", Type: "boolean"},
		{Name: "completedAt", Type: "number"},
		{Name: "createdAt", Type: "number"},
		{Name: "startAt", Type: "number"},
	}
}

func huntrNoteFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "memberId", Type: "string"},
		{Name: "text", Type: "string"},
		{Name: "htmlText", Type: "string"},
	}
}

func huntrActionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "actionType", Type: "string"},
		{Name: "memberId", Type: "string"},
		{Name: "candidateId", Type: "string"},
		{Name: "activityId", Type: "string"},
		{Name: "createdAt", Type: "number"},
		{Name: "date", Type: "number"},
	}
}

func huntrMemberRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"email":      item["email"],
		"fullName":   item["fullName"],
		"givenName":  item["givenName"],
		"familyName": item["familyName"],
		"isActive":   item["isActive"],
		"createdAt":  item["createdAt"],
		"lastSeenAt": item["lastSeenAt"],
		"boardIds":   item["boardIds"],
	}
}

func huntrCandidateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"email":     item["email"],
		"firstName": item["firstName"],
		"lastName":  item["lastName"],
		"memberId":  item["memberId"],
	}
}

func huntrActivityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"title":            item["title"],
		"activityCategory": item["activityCategory"],
		"completed":        item["completed"],
		"completedAt":      item["completedAt"],
		"createdAt":        item["createdAt"],
		"startAt":          item["startAt"],
	}
}

func huntrNoteRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"memberId": item["memberId"],
		"text":     item["text"],
		"htmlText": item["htmlText"],
	}
}

func huntrActionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"actionType":  item["actionType"],
		"memberId":    item["memberId"],
		"candidateId": item["candidateId"],
		"activityId":  item["activityId"],
		"createdAt":   item["createdAt"],
		"date":        item["date"],
	}
}
