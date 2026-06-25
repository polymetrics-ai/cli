package acuityscheduling

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Acuity API resource path (relative to
// base_url) it reads from, the record mapper that flattens its objects, and
// whether the endpoint supports page-number pagination. Acuity returns a JSON
// array at the root for every list endpoint; only /appointments paginates with
// max/page, the rest return a single full list.
type streamEndpoint struct {
	// resource is the Acuity list endpoint path segment (e.g. "appointments").
	resource string
	// paginated marks endpoints that honor the max/page query params. When
	// false the read issues a single request.
	paginated bool
	// mapRecord flattens a raw Acuity object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// acuityStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in acuityStreams; the read path
// is fully data-driven from this table.
var acuityStreamEndpoints = map[string]streamEndpoint{
	"appointments":      {resource: "appointments", paginated: true, mapRecord: acuityAppointmentRecord},
	"clients":           {resource: "clients", paginated: false, mapRecord: acuityClientRecord},
	"appointment-types": {resource: "appointment-types", paginated: false, mapRecord: acuityAppointmentTypeRecord},
	"calendars":         {resource: "calendars", paginated: false, mapRecord: acuityCalendarRecord},
	"forms":             {resource: "forms", paginated: false, mapRecord: acuityFormRecord},
}

// acuityStreams returns the connector's published stream catalog. Acuity only
// supports full-refresh sync (its list endpoints have no incremental cursor),
// though appointments carry a datetime usable as a soft cursor field.
func acuityStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "appointments",
			Description:  "Scheduled Acuity appointments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"datetime"},
			Fields:       acuityAppointmentFields(),
		},
		{
			Name:        "clients",
			Description: "Acuity clients (people who have booked).",
			PrimaryKey:  []string{"email"},
			Fields:      acuityClientFields(),
		},
		{
			Name:        "appointment-types",
			Description: "Acuity appointment types offered for booking.",
			PrimaryKey:  []string{"id"},
			Fields:      acuityAppointmentTypeFields(),
		},
		{
			Name:        "calendars",
			Description: "Acuity calendars (staff/resources).",
			PrimaryKey:  []string{"id"},
			Fields:      acuityCalendarFields(),
		},
		{
			Name:        "forms",
			Description: "Acuity intake forms.",
			PrimaryKey:  []string{"id"},
			Fields:      acuityFormFields(),
		},
	}
}

func acuityAppointmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "datetime", Type: "string"},
		{Name: "end_time", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "time", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "appointment_type_id", Type: "integer"},
		{Name: "calendar", Type: "string"},
		{Name: "calendar_id", Type: "integer"},
		{Name: "duration", Type: "string"},
		{Name: "price", Type: "string"},
		{Name: "paid", Type: "string"},
		{Name: "amount_paid", Type: "string"},
		{Name: "canceled", Type: "boolean"},
		{Name: "datetime_created", Type: "string"},
	}
}

func acuityClientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
	}
}

func acuityAppointmentTypeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "description", Type: "string"},
		{Name: "duration", Type: "integer"},
		{Name: "price", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "private", Type: "boolean"},
		{Name: "type", Type: "string"},
	}
}

func acuityCalendarFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "replyTo", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "location", Type: "string"},
		{Name: "timezone", Type: "string"},
	}
}

func acuityFormFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "hidden", Type: "boolean"},
	}
}

func acuityAppointmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"first_name":          item["firstName"],
		"last_name":           item["lastName"],
		"email":               item["email"],
		"phone":               item["phone"],
		"datetime":            item["datetime"],
		"end_time":            item["endTime"],
		"date":                item["date"],
		"time":                item["time"],
		"type":                item["type"],
		"appointment_type_id": item["appointmentTypeID"],
		"calendar":            item["calendar"],
		"calendar_id":         item["calendarID"],
		"duration":            item["duration"],
		"price":               item["price"],
		"paid":                item["paid"],
		"amount_paid":         item["amountPaid"],
		"canceled":            item["canceled"],
		"datetime_created":    item["datetimeCreated"],
	}
}

func acuityClientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"first_name": item["firstName"],
		"last_name":  item["lastName"],
		"email":      item["email"],
		"phone":      item["phone"],
	}
}

func acuityAppointmentTypeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"active":      item["active"],
		"description": item["description"],
		"duration":    item["duration"],
		"price":       item["price"],
		"category":    item["category"],
		"color":       item["color"],
		"private":     item["private"],
		"type":        item["type"],
	}
}

func acuityCalendarRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"email":       item["email"],
		"replyTo":     item["replyTo"],
		"description": item["description"],
		"location":    item["location"],
		"timezone":    item["timezone"],
	}
}

func acuityFormRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"hidden":      item["hidden"],
	}
}
