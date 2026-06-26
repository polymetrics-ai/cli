package granola

import "polymetrics.ai/internal/connectors"

// granolaStreams returns the connector's published stream catalog. Every Granola
// note exposes a string id and an RFC3339 created_at timestamp, so the primary
// key is ["id"] and the incremental cursor field is ["created_at"].
//
// The two core streams mirror the upstream Airbyte connector:
//   - notes: lightweight list metadata from GET /notes.
//   - detailed_notes: the full note (summary, owner, attendees, calendar event)
//     from GET /notes/{id}, fanned out from the notes list.
func granolaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "notes",
			Description:  "Granola meeting notes metadata (list view).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       granolaNoteFields(),
		},
		{
			Name:         "detailed_notes",
			Description:  "Granola notes with full detail: summary, owner, attendees, and calendar event.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       granolaDetailedNoteFields(),
		},
	}
}

func granolaNoteFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "owner_name", Type: "string"},
		{Name: "owner_email", Type: "string"},
	}
}

func granolaDetailedNoteFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "owner_name", Type: "string"},
		{Name: "owner_email", Type: "string"},
		{Name: "summary", Type: "string"},
		{Name: "transcript", Type: "json"},
		{Name: "attendees", Type: "json"},
		{Name: "calendar_event", Type: "json"},
		{Name: "folders", Type: "json"},
	}
}

// granolaNoteRecord flattens a note list item into a connectors.Record. The
// owner sub-object (if present) is flattened to owner_name / owner_email.
func granolaNoteRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":         item["id"],
		"title":      item["title"],
		"object":     item["object"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
	name, email := ownerFields(item["owner"])
	rec["owner_name"] = name
	rec["owner_email"] = email
	return rec
}

// granolaDetailedNoteRecord flattens a single full note into a Record, preserving
// the richer nested structures (transcript, attendees, calendar_event, folders)
// as-is so downstream consumers can project what they need.
func granolaDetailedNoteRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":             item["id"],
		"title":          item["title"],
		"object":         item["object"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
		"summary":        item["summary"],
		"transcript":     item["transcript"],
		"attendees":      item["attendees"],
		"calendar_event": item["calendar_event"],
		"folders":        item["folders"],
	}
	name, email := ownerFields(item["owner"])
	rec["owner_name"] = name
	rec["owner_email"] = email
	return rec
}

// ownerFields extracts name/email from a Granola owner sub-object, tolerating a
// missing or non-object value.
func ownerFields(raw any) (any, any) {
	owner, ok := raw.(map[string]any)
	if !ok {
		return nil, nil
	}
	return owner["name"], owner["email"]
}
