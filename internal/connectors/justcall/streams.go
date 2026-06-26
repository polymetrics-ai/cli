package justcall

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the JustCall API resource path (relative
// to base_url) it reads from, the HTTP method, whether it paginates, and the
// record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the path segment of the list endpoint (e.g. "v2.1/calls").
	resource string
	// method is the HTTP verb. JustCall's v1 list endpoints (contacts, numbers)
	// are POST; the v2.1 endpoints are GET.
	method string
	// paginated is false for endpoints that return the full set in one response
	// (phone_numbers), true for the page-increment endpoints.
	paginated bool
	// mapRecord flattens a raw JustCall object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// justcallStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in justcallStreams; the read
// path is fully data-driven from this table.
var justcallStreamEndpoints = map[string]streamEndpoint{
	"users":         {resource: "v2.1/users", method: "GET", paginated: true, mapRecord: justcallUserRecord},
	"calls":         {resource: "v2.1/calls", method: "GET", paginated: true, mapRecord: justcallCallRecord},
	"sms":           {resource: "v2.1/texts", method: "GET", paginated: true, mapRecord: justcallSMSRecord},
	"contacts":      {resource: "v1/contacts/list", method: "POST", paginated: true, mapRecord: justcallContactRecord},
	"phone_numbers": {resource: "v1/numbers/list", method: "POST", paginated: false, mapRecord: justcallPhoneNumberRecord},
}

// justcallStreams returns the connector's published stream catalog. Calls and
// SMS carry a date cursor (call_date / sms_date); the others are full-refresh
// with an id primary key.
func justcallStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "users",
			Description: "JustCall users (agents).",
			PrimaryKey:  []string{"id"},
			Fields:      justcallUserFields(),
		},
		{
			Name:         "calls",
			Description:  "JustCall call logs.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"call_date"},
			Fields:       justcallCallFields(),
		},
		{
			Name:         "sms",
			Description:  "JustCall SMS / texts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"sms_date"},
			Fields:       justcallSMSFields(),
		},
		{
			Name:        "contacts",
			Description: "JustCall contacts.",
			PrimaryKey:  []string{"id"},
			Fields:      justcallContactFields(),
		},
		{
			Name:        "phone_numbers",
			Description: "JustCall phone numbers.",
			PrimaryKey:  []string{"id"},
			Fields:      justcallPhoneNumberFields(),
		},
	}
}

func justcallUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "extension", Type: "string"},
		{Name: "available", Type: "string"},
		{Name: "on_call", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "last_login_timestamp", Type: "string"},
	}
}

func justcallCallFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "call_date", Type: "string"},
		{Name: "call_time", Type: "string"},
		{Name: "call_sid", Type: "string"},
		{Name: "agent_id", Type: "string"},
		{Name: "agent_name", Type: "string"},
		{Name: "agent_email", Type: "string"},
		{Name: "contact_name", Type: "string"},
		{Name: "contact_number", Type: "string"},
		{Name: "call_duration", Type: "string"},
		{Name: "cost_incurred", Type: "string"},
		{Name: "justcall_number", Type: "string"},
		{Name: "justcall_line_name", Type: "string"},
	}
}

func justcallSMSFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "sms_date", Type: "string"},
		{Name: "sms_time", Type: "string"},
		{Name: "direction", Type: "string"},
		{Name: "delivery_status", Type: "string"},
		{Name: "agent_id", Type: "string"},
		{Name: "agent_name", Type: "string"},
		{Name: "agent_email", Type: "string"},
		{Name: "contact_name", Type: "string"},
		{Name: "contact_number", Type: "string"},
		{Name: "justcall_number", Type: "string"},
		{Name: "justcall_line_name", Type: "string"},
		{Name: "cost_incurred", Type: "string"},
	}
}

func justcallContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "firstname", Type: "string"},
		{Name: "lastname", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "notes", Type: "string"},
	}
}

func justcallPhoneNumberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "friendly_name", Type: "string"},
		{Name: "custom_name", Type: "string"},
		{Name: "agent_id", Type: "string"},
		{Name: "capabilities", Type: "string"},
	}
}

func justcallUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                   item["id"],
		"name":                 item["name"],
		"email":                item["email"],
		"role":                 item["role"],
		"extension":            item["extension"],
		"available":            item["available"],
		"on_call":              item["on_call"],
		"timezone":             item["timezone"],
		"created_at":           item["created_at"],
		"last_login_timestamp": item["last_login_timestamp"],
	}
}

func justcallCallRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"call_date":          item["call_date"],
		"call_time":          item["call_time"],
		"call_sid":           item["call_sid"],
		"agent_id":           item["agent_id"],
		"agent_name":         item["agent_name"],
		"agent_email":        item["agent_email"],
		"contact_name":       item["contact_name"],
		"contact_number":     item["contact_number"],
		"call_duration":      item["call_duration"],
		"cost_incurred":      item["cost_incurred"],
		"justcall_number":    item["justcall_number"],
		"justcall_line_name": item["justcall_line_name"],
	}
}

func justcallSMSRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"sms_date":           item["sms_date"],
		"sms_time":           item["sms_time"],
		"direction":          item["direction"],
		"delivery_status":    item["delivery_status"],
		"agent_id":           item["agent_id"],
		"agent_name":         item["agent_name"],
		"agent_email":        item["agent_email"],
		"contact_name":       item["contact_name"],
		"contact_number":     item["contact_number"],
		"justcall_number":    item["justcall_number"],
		"justcall_line_name": item["justcall_line_name"],
		"cost_incurred":      item["cost_incurred"],
	}
}

func justcallContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"firstname": item["firstname"],
		"lastname":  item["lastname"],
		"email":     item["email"],
		"phone":     item["phone"],
		"company":   item["company"],
		"notes":     item["notes"],
	}
}

func justcallPhoneNumberRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"phone":         item["phone"],
		"friendly_name": item["friendly_name"],
		"custom_name":   item["custom_name"],
		"agent_id":      item["agent_id"],
		"capabilities":  item["capabilities"],
	}
}
