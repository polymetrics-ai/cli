package googlecalendar

import (
	"strings"

	"polymetrics.ai/internal/connectors"
)

// streamEndpoint maps a stream name to the Google Calendar API resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects. Some endpoints embed the configured calendar id; the {calendarId}
// placeholder is substituted at request time.
type streamEndpoint struct {
	// resource is the Calendar list endpoint path. It may contain the
	// {calendarId} placeholder, replaced with the resolved calendar id.
	resource string
	// mapRecord flattens a raw Calendar object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// path returns the request path with the calendar id substituted in.
func (e streamEndpoint) path(calendarID string) string {
	return strings.ReplaceAll(e.resource, "{calendarId}", calendarID)
}

// googleCalendarStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in googleCalendarStreams;
// the read path is fully data-driven from this table. All Calendar v3 list
// endpoints return {items:[...], nextPageToken:"..."}.
var googleCalendarStreamEndpoints = map[string]streamEndpoint{
	"calendar_list": {resource: "users/me/calendarList", mapRecord: calendarListRecord},
	"events":        {resource: "calendars/{calendarId}/events", mapRecord: eventRecord},
	"settings":      {resource: "users/me/settings", mapRecord: settingRecord},
	"acl":           {resource: "calendars/{calendarId}/acl", mapRecord: aclRecord},
}

// googleCalendarStreams returns the connector's published stream catalog.
func googleCalendarStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "calendar_list",
			Description:  "Calendars on the authenticated user's calendar list (users.calendarList).",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       calendarListFields(),
		},
		{
			Name:         "events",
			Description:  "Events on the configured calendar (events.list). Cursor is the event updated timestamp.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       eventFields(),
		},
		{
			Name:         "settings",
			Description:  "User calendar settings (settings.list).",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       settingFields(),
		},
		{
			Name:         "acl",
			Description:  "Access control rules for the configured calendar (acl.list).",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       aclFields(),
		},
	}
}

func calendarListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "summary", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "timeZone", Type: "string"},
		{Name: "colorId", Type: "string"},
		{Name: "accessRole", Type: "string"},
		{Name: "primary", Type: "boolean"},
		{Name: "selected", Type: "boolean"},
		{Name: "hidden", Type: "boolean"},
		{Name: "deleted", Type: "boolean"},
		{Name: "etag", Type: "string"},
		{Name: "kind", Type: "string"},
	}
}

func eventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "iCalUID", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "summary", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "location", Type: "string"},
		{Name: "htmlLink", Type: "string"},
		{Name: "created", Type: "timestamp"},
		{Name: "updated", Type: "timestamp"},
		{Name: "start", Type: "object"},
		{Name: "end", Type: "object"},
		{Name: "creator", Type: "object"},
		{Name: "organizer", Type: "object"},
		{Name: "attendees", Type: "array"},
		{Name: "recurringEventId", Type: "string"},
		{Name: "etag", Type: "string"},
		{Name: "kind", Type: "string"},
	}
}

func settingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "value", Type: "string"},
		{Name: "etag", Type: "string"},
		{Name: "kind", Type: "string"},
	}
}

func aclFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "scope", Type: "object"},
		{Name: "etag", Type: "string"},
		{Name: "kind", Type: "string"},
	}
}

func calendarListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"summary":     item["summary"],
		"description": item["description"],
		"timeZone":    item["timeZone"],
		"colorId":     item["colorId"],
		"accessRole":  item["accessRole"],
		"primary":     item["primary"],
		"selected":    item["selected"],
		"hidden":      item["hidden"],
		"deleted":     item["deleted"],
		"etag":        item["etag"],
		"kind":        item["kind"],
	}
}

func eventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"iCalUID":          item["iCalUID"],
		"status":           item["status"],
		"summary":          item["summary"],
		"description":      item["description"],
		"location":         item["location"],
		"htmlLink":         item["htmlLink"],
		"created":          item["created"],
		"updated":          item["updated"],
		"start":            item["start"],
		"end":              item["end"],
		"creator":          item["creator"],
		"organizer":        item["organizer"],
		"attendees":        item["attendees"],
		"recurringEventId": item["recurringEventId"],
		"etag":             item["etag"],
		"kind":             item["kind"],
	}
}

func settingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":    item["id"],
		"value": item["value"],
		"etag":  item["etag"],
		"kind":  item["kind"],
	}
}

func aclRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":    item["id"],
		"role":  item["role"],
		"scope": item["scope"],
		"etag":  item["etag"],
		"kind":  item["kind"],
	}
}
