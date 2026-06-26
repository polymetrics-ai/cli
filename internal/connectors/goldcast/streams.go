package goldcast

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Goldcast API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Goldcast list endpoint path (e.g. "event/").
	resource string
	// mapRecord flattens a raw Goldcast object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// goldcastStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in goldcastStreams; the read
// path is fully data-driven from this table.
//
// Only the "parent" Goldcast list endpoints are exposed (no partition-routed
// children like per-event webinars/members) to keep the core set tight and
// credential-free fixture mode simple. Every object exposes a string "id".
var goldcastStreamEndpoints = map[string]streamEndpoint{
	"organizations":     {resource: "core/organization/", mapRecord: goldcastOrganizationRecord},
	"events":            {resource: "event/", mapRecord: goldcastEventRecord},
	"agenda_items":      {resource: "event/agenda-item/", mapRecord: goldcastAgendaItemRecord},
	"discussion_groups": {resource: "event/discussion-groups/", mapRecord: goldcastDiscussionGroupRecord},
	"tracks":            {resource: "event/tracks/", mapRecord: goldcastTrackRecord},
}

// goldcastStreams returns the connector's published stream catalog. The Goldcast
// API supports only full-refresh syncs (no cursor field), so CursorFields is
// empty across the board and the primary key is ["id"].
func goldcastStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "organizations",
			Description: "Goldcast organizations.",
			PrimaryKey:  []string{"id"},
			Fields:      goldcastOrganizationFields(),
		},
		{
			Name:        "events",
			Description: "Goldcast events.",
			PrimaryKey:  []string{"id"},
			Fields:      goldcastEventFields(),
		},
		{
			Name:        "agenda_items",
			Description: "Goldcast event agenda items.",
			PrimaryKey:  []string{"id"},
			Fields:      goldcastAgendaItemFields(),
		},
		{
			Name:        "discussion_groups",
			Description: "Goldcast event discussion groups.",
			PrimaryKey:  []string{"id"},
			Fields:      goldcastDiscussionGroupFields(),
		},
		{
			Name:        "tracks",
			Description: "Goldcast event tracks.",
			PrimaryKey:  []string{"id"},
			Fields:      goldcastTrackFields(),
		},
	}
}

func goldcastOrganizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "domain", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func goldcastEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "organization", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "start_time", Type: "string"},
		{Name: "end_time", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func goldcastAgendaItemFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "event", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "start_time", Type: "string"},
		{Name: "end_time", Type: "string"},
	}
}

func goldcastDiscussionGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "event", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "capacity", Type: "integer"},
		{Name: "created_at", Type: "string"},
	}
}

func goldcastTrackFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "event", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "color", Type: "string"},
	}
}

func goldcastOrganizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"slug":       item["slug"],
		"domain":     item["domain"],
		"created_at": item["created_at"],
	}
}

func goldcastEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"title":        item["title"],
		"organization": item["organization"],
		"status":       item["status"],
		"start_time":   item["start_time"],
		"end_time":     item["end_time"],
		"timezone":     item["timezone"],
		"created_at":   item["created_at"],
	}
}

func goldcastAgendaItemRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"event":       item["event"],
		"title":       item["title"],
		"description": item["description"],
		"start_time":  item["start_time"],
		"end_time":    item["end_time"],
	}
}

func goldcastDiscussionGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"event":      item["event"],
		"name":       item["name"],
		"capacity":   item["capacity"],
		"created_at": item["created_at"],
	}
}

func goldcastTrackRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":    item["id"],
		"event": item["event"],
		"name":  item["name"],
		"color": item["color"],
	}
}
