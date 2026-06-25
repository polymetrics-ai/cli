package calendly

import "polymetrics.ai/internal/connectors"

// scope describes the query parameter a Calendly list endpoint requires to be
// bound to the authenticated user's organization or user URI.
type scope int

const (
	scopeNone scope = iota // endpoint needs no org/user binding (e.g. /users/me)
	scopeOrg               // endpoint takes ?organization=<org uri>
	scopeUser              // endpoint takes ?user=<user uri>
)

// streamEndpoint maps a stream name to the Calendly API resource path (relative
// to base_url), the scope param it requires, and the record mapper that flattens
// its objects. The read path is fully data-driven from this table.
type streamEndpoint struct {
	// resource is the list endpoint path segment (e.g. "scheduled_events").
	resource string
	// scope selects the org/user query binding applied to list requests.
	scope scope
	// single marks endpoints that return a single {resource:{...}} object
	// rather than a {collection:[...]} list (e.g. /users/me).
	single bool
	// mapRecord flattens a raw Calendly object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// cursor is the field used as the incremental cursor, if any.
	cursor string
}

// calendlyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in calendlyStreams.
var calendlyStreamEndpoints = map[string]streamEndpoint{
	"event_types":              {resource: "event_types", scope: scopeOrg, mapRecord: calendlyEventTypeRecord, cursor: "updated_at"},
	"scheduled_events":         {resource: "scheduled_events", scope: scopeOrg, mapRecord: calendlyScheduledEventRecord, cursor: "start_time"},
	"organization_memberships": {resource: "organization_memberships", scope: scopeOrg, mapRecord: calendlyMembershipRecord, cursor: "updated_at"},
	"groups":                   {resource: "groups", scope: scopeOrg, mapRecord: calendlyGroupRecord, cursor: "updated_at"},
	"users":                    {resource: "users/me", scope: scopeNone, single: true, mapRecord: calendlyUserRecord},
}

// calendlyStreams returns the connector's published stream catalog. Calendly
// objects are addressed by a `uri`; the connector derives a stable `id` (the
// trailing URI segment) used as the primary key.
func calendlyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "scheduled_events",
			Description:  "Calendly scheduled (booked) events for the organization.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"start_time"},
			Fields:       calendlyScheduledEventFields(),
		},
		{
			Name:         "event_types",
			Description:  "Calendly event types available in the organization.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       calendlyEventTypeFields(),
		},
		{
			Name:         "organization_memberships",
			Description:  "Members of the Calendly organization.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       calendlyMembershipFields(),
		},
		{
			Name:         "groups",
			Description:  "Calendly organization groups.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       calendlyGroupFields(),
		},
		{
			Name:        "users",
			Description: "The authenticated Calendly user (current user).",
			PrimaryKey:  []string{"id"},
			Fields:      calendlyUserFields(),
		},
	}
}

func calendlyScheduledEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "uri", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "start_time", Type: "timestamp"},
		{Name: "end_time", Type: "timestamp"},
		{Name: "event_type", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func calendlyEventTypeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "uri", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "duration", Type: "integer"},
		{Name: "kind", Type: "string"},
		{Name: "scheduling_url", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func calendlyMembershipFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "uri", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "user", Type: "string"},
		{Name: "user_name", Type: "string"},
		{Name: "user_email", Type: "string"},
		{Name: "organization", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func calendlyGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "uri", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "organization", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func calendlyUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "uri", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "current_organization", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func calendlyScheduledEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         idFromURI(item["uri"]),
		"uri":        item["uri"],
		"name":       item["name"],
		"status":     item["status"],
		"start_time": item["start_time"],
		"end_time":   item["end_time"],
		"event_type": item["event_type"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func calendlyEventTypeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             idFromURI(item["uri"]),
		"uri":            item["uri"],
		"name":           item["name"],
		"slug":           item["slug"],
		"active":         item["active"],
		"duration":       item["duration"],
		"kind":           item["kind"],
		"scheduling_url": item["scheduling_url"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
	}
}

func calendlyMembershipRecord(item map[string]any) connectors.Record {
	user := asObject(item["user"])
	return connectors.Record{
		"id":           idFromURI(item["uri"]),
		"uri":          item["uri"],
		"role":         item["role"],
		"user":         user["uri"],
		"user_name":    user["name"],
		"user_email":   user["email"],
		"organization": item["organization"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func calendlyGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           idFromURI(item["uri"]),
		"uri":          item["uri"],
		"name":         item["name"],
		"organization": item["organization"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func calendlyUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                   idFromURI(item["uri"]),
		"uri":                  item["uri"],
		"name":                 item["name"],
		"email":                item["email"],
		"slug":                 item["slug"],
		"timezone":             item["timezone"],
		"current_organization": item["current_organization"],
		"created_at":           item["created_at"],
		"updated_at":           item["updated_at"],
	}
}
