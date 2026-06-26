package eventbrite

import "polymetrics.ai/internal/connectors"

// scope describes the resource a stream is rooted under. Eventbrite list
// endpoints are addressed under either the authenticated user, an organization,
// or an event, so the read path needs to know which id (if any) to splice into
// the path before requesting.
type scope int

const (
	// scopeUser reads from /users/me/... and needs no configured id.
	scopeUser scope = iota
	// scopeOrg reads from /organizations/{organization_id}/...
	scopeOrg
	// scopeEvent reads from /events/{event_id}/...
	scopeEvent
)

// streamEndpoint maps a stream name to the Eventbrite API resource it reads
// from, the scope id it requires, the JSON key holding the records array in the
// list response, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// pathTemplate is the endpoint path with a single %s placeholder for the
	// scope id (scopeUser templates contain no placeholder).
	pathTemplate string
	scope        scope
	// recordsKey is the top-level JSON key whose array holds the records.
	recordsKey string
	mapRecord  func(map[string]any) connectors.Record
}

// eventbriteStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in eventbriteStreams; the
// read path is fully data-driven from this table.
var eventbriteStreamEndpoints = map[string]streamEndpoint{
	"organizations":  {pathTemplate: "users/me/organizations/", scope: scopeUser, recordsKey: "organizations", mapRecord: organizationRecord},
	"events":         {pathTemplate: "organizations/%s/events/", scope: scopeOrg, recordsKey: "events", mapRecord: eventRecord},
	"attendees":      {pathTemplate: "events/%s/attendees/", scope: scopeEvent, recordsKey: "attendees", mapRecord: attendeeRecord},
	"orders":         {pathTemplate: "events/%s/orders/", scope: scopeEvent, recordsKey: "orders", mapRecord: orderRecord},
	"ticket_classes": {pathTemplate: "events/%s/ticket_classes/", scope: scopeEvent, recordsKey: "ticket_classes", mapRecord: ticketClassRecord},
}

// eventbriteStreams returns the connector's published stream catalog. Eventbrite
// objects expose a string id; mutable objects carry a `changed` RFC3339
// timestamp used as the incremental cursor where available.
func eventbriteStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "organizations",
			Description: "Eventbrite organizations the authenticated user belongs to.",
			PrimaryKey:  []string{"id"},
			Fields:      organizationFields(),
		},
		{
			Name:         "events",
			Description:  "Events under the configured organization.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"changed"},
			Fields:       eventFields(),
		},
		{
			Name:         "attendees",
			Description:  "Attendees of the configured event.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"changed"},
			Fields:       attendeeFields(),
		},
		{
			Name:         "orders",
			Description:  "Orders placed for the configured event.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"changed"},
			Fields:       orderFields(),
		},
		{
			Name:         "ticket_classes",
			Description:  "Ticket classes defined for the configured event.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"changed"},
			Fields:       ticketClassFields(),
		},
	}
}

func organizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "vertical", Type: "string"},
		{Name: "locale", Type: "string"},
		{Name: "image_id", Type: "string"},
	}
}

func eventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "start", Type: "string"},
		{Name: "end", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "changed", Type: "string"},
		{Name: "published", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "online_event", Type: "boolean"},
		{Name: "listed", Type: "boolean"},
		{Name: "organization_id", Type: "string"},
		{Name: "venue_id", Type: "string"},
		{Name: "capacity", Type: "integer"},
	}
}

func attendeeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "event_id", Type: "string"},
		{Name: "order_id", Type: "string"},
		{Name: "ticket_class_id", Type: "string"},
		{Name: "ticket_class_name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "changed", Type: "string"},
		{Name: "quantity", Type: "integer"},
		{Name: "checked_in", Type: "boolean"},
		{Name: "cancelled", Type: "boolean"},
		{Name: "refunded", Type: "boolean"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
	}
}

func orderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "event_id", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "changed", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "time_remaining", Type: "integer"},
	}
}

func ticketClassFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "event_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "cost", Type: "string"},
		{Name: "fee", Type: "string"},
		{Name: "quantity_total", Type: "integer"},
		{Name: "quantity_sold", Type: "integer"},
		{Name: "free", Type: "boolean"},
		{Name: "hidden", Type: "boolean"},
		{Name: "on_sale_status", Type: "string"},
	}
}

func organizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"name":     flatten(item["name"]),
		"vertical": item["vertical"],
		"locale":   item["locale"],
		"image_id": item["image_id"],
	}
}

func eventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            flattenText(item["name"]),
		"description":     flattenText(item["description"]),
		"url":             item["url"],
		"start":           flattenUTC(item["start"]),
		"end":             flattenUTC(item["end"]),
		"created":         item["created"],
		"changed":         item["changed"],
		"published":       item["published"],
		"status":          item["status"],
		"currency":        item["currency"],
		"online_event":    item["online_event"],
		"listed":          item["listed"],
		"organization_id": item["organization_id"],
		"venue_id":        item["venue_id"],
		"capacity":        item["capacity"],
	}
}

func attendeeRecord(item map[string]any) connectors.Record {
	profile, _ := item["profile"].(map[string]any)
	var attendeeName, attendeeEmail any
	if profile != nil {
		attendeeName = flatten(profile["name"])
		attendeeEmail = flatten(profile["email"])
	}
	return connectors.Record{
		"id":                item["id"],
		"event_id":          item["event_id"],
		"order_id":          item["order_id"],
		"ticket_class_id":   item["ticket_class_id"],
		"ticket_class_name": item["ticket_class_name"],
		"status":            item["status"],
		"created":           item["created"],
		"changed":           item["changed"],
		"quantity":          item["quantity"],
		"checked_in":        item["checked_in"],
		"cancelled":         item["cancelled"],
		"refunded":          item["refunded"],
		"name":              attendeeName,
		"email":             attendeeEmail,
	}
}

func orderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"event_id":       item["event_id"],
		"created":        item["created"],
		"changed":        item["changed"],
		"status":         item["status"],
		"name":           flatten(item["name"]),
		"email":          flatten(item["email"]),
		"time_remaining": item["time_remaining"],
	}
}

func ticketClassRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"event_id":       item["event_id"],
		"name":           flattenText(item["name"]),
		"description":    flattenText(item["description"]),
		"cost":           flattenCurrency(item["cost"]),
		"fee":            flattenCurrency(item["fee"]),
		"quantity_total": item["quantity_total"],
		"quantity_sold":  item["quantity_sold"],
		"free":           item["free"],
		"hidden":         item["hidden"],
		"on_sale_status": item["on_sale_status"],
	}
}

// flatten returns a scalar value as-is, or the inner "text"/"email" string of a
// nested multipart object when Eventbrite wraps a string in {text, html} or
// similar. Unknown maps fall through to nil.
func flatten(v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case string:
		return t
	case map[string]any:
		if s, ok := t["text"].(string); ok {
			return s
		}
		if s, ok := t["email"].(string); ok {
			return s
		}
		if s, ok := t["name"].(string); ok {
			return s
		}
		return nil
	default:
		return v
	}
}

// flattenText flattens an Eventbrite multipart-text object ({text, html}) to its
// text value, leaving plain strings untouched.
func flattenText(v any) any {
	if m, ok := v.(map[string]any); ok {
		if s, ok := m["text"].(string); ok {
			return s
		}
		return nil
	}
	return v
}

// flattenUTC flattens an Eventbrite datetime-tz object ({timezone, local, utc})
// to its utc value, leaving plain strings untouched.
func flattenUTC(v any) any {
	if m, ok := v.(map[string]any); ok {
		if s, ok := m["utc"].(string); ok {
			return s
		}
		return nil
	}
	return v
}

// flattenCurrency flattens an Eventbrite currency object ({display, value,
// currency}) to its display string, leaving plain strings untouched.
func flattenCurrency(v any) any {
	if m, ok := v.(map[string]any); ok {
		if s, ok := m["display"].(string); ok {
			return s
		}
		return nil
	}
	return v
}
