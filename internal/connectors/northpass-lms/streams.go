package northpasslms

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Northpass API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its JSON:API
// objects ({id,type,attributes}) into a flat connectors.Record.
type streamEndpoint struct {
	// resource is the Northpass list endpoint path segment (e.g. "courses").
	resource string
	// mapRecord flattens a raw JSON:API object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// northpassStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in northpassStreams; the read
// path is fully data-driven from this table.
var northpassStreamEndpoints = map[string]streamEndpoint{
	"people":             {resource: "people", mapRecord: northpassPersonRecord},
	"courses":            {resource: "courses", mapRecord: northpassCourseRecord},
	"course_enrollments": {resource: "course_enrollments", mapRecord: northpassEnrollmentRecord},
	"groups":             {resource: "groups", mapRecord: northpassGroupRecord},
}

// northpassStreams returns the connector's published stream catalog. Northpass
// objects expose a string id; primary key is ["id"]. Only full-refresh sync is
// supported upstream, so no cursor fields are advertised.
func northpassStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "people",
			Description: "Northpass people (learners).",
			PrimaryKey:  []string{"id"},
			Fields:      northpassPersonFields(),
		},
		{
			Name:        "courses",
			Description: "Northpass courses.",
			PrimaryKey:  []string{"id"},
			Fields:      northpassCourseFields(),
		},
		{
			Name:        "course_enrollments",
			Description: "Northpass course enrollments.",
			PrimaryKey:  []string{"id"},
			Fields:      northpassEnrollmentFields(),
		},
		{
			Name:        "groups",
			Description: "Northpass groups.",
			PrimaryKey:  []string{"id"},
			Fields:      northpassGroupFields(),
		},
	}
}

func northpassPersonFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func northpassCourseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func northpassEnrollmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "percentage", Type: "integer"},
		{Name: "learner_id", Type: "string"},
		{Name: "course_id", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "completed_at", Type: "timestamp"},
	}
}

func northpassGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func northpassPersonRecord(item map[string]any) connectors.Record {
	attrs := attributesOf(item)
	return connectors.Record{
		"id":         item["id"],
		"type":       item["type"],
		"email":      attrs["email"],
		"first_name": attrs["first_name"],
		"last_name":  attrs["last_name"],
		"status":     attrs["status"],
		"created_at": attrs["created_at"],
		"updated_at": attrs["updated_at"],
	}
}

func northpassCourseRecord(item map[string]any) connectors.Record {
	attrs := attributesOf(item)
	return connectors.Record{
		"id":         item["id"],
		"type":       item["type"],
		"name":       attrs["name"],
		"slug":       attrs["slug"],
		"status":     attrs["status"],
		"created_at": attrs["created_at"],
		"updated_at": attrs["updated_at"],
	}
}

func northpassEnrollmentRecord(item map[string]any) connectors.Record {
	attrs := attributesOf(item)
	return connectors.Record{
		"id":           item["id"],
		"type":         item["type"],
		"status":       attrs["status"],
		"percentage":   attrs["percentage"],
		"learner_id":   attrs["learner_id"],
		"course_id":    attrs["course_id"],
		"created_at":   attrs["created_at"],
		"updated_at":   attrs["updated_at"],
		"completed_at": attrs["completed_at"],
	}
}

func northpassGroupRecord(item map[string]any) connectors.Record {
	attrs := attributesOf(item)
	return connectors.Record{
		"id":         item["id"],
		"type":       item["type"],
		"name":       attrs["name"],
		"slug":       attrs["slug"],
		"created_at": attrs["created_at"],
		"updated_at": attrs["updated_at"],
	}
}

// attributesOf returns the JSON:API "attributes" object of a record, or an empty
// map when absent. This lets the mappers flatten {id,type,attributes:{...}} into
// a flat record without panicking on shape differences.
func attributesOf(item map[string]any) map[string]any {
	if attrs, ok := item["attributes"].(map[string]any); ok {
		return attrs
	}
	return map[string]any{}
}
