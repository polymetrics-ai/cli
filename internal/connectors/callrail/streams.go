package callrail

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the CallRail API resource file (relative
// to the account-scoped path), the JSON key holding the record array in the
// paginated response, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the account-scoped path segment, e.g. "calls.json".
	resource string
	// arrayKey is the top-level response key holding the records, e.g. "calls".
	arrayKey string
	// mapRecord flattens a raw CallRail object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// callrailStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in callrailStreams; the read
// path is fully data-driven from this table.
var callrailStreamEndpoints = map[string]streamEndpoint{
	"calls":         {resource: "calls.json", arrayKey: "calls", mapRecord: callRecord},
	"companies":     {resource: "companies.json", arrayKey: "companies", mapRecord: companyRecord},
	"users":         {resource: "users.json", arrayKey: "users", mapRecord: userRecord},
	"text_messages": {resource: "text-messages.json", arrayKey: "text_messages", mapRecord: textMessageRecord},
}

// callrailStreams returns the connector's published stream catalog. Every
// CallRail object exposes a string id; the cursor field is the per-stream
// timestamp the API filters and orders on.
func callrailStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "calls",
			Description:  "CallRail call detail records.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"start_time"},
			Fields:       callFields(),
		},
		{
			Name:         "companies",
			Description:  "CallRail companies.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       companyFields(),
		},
		{
			Name:         "users",
			Description:  "CallRail users.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       userFields(),
		},
		{
			Name:         "text_messages",
			Description:  "CallRail text message conversations.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_message_at"},
			Fields:       textMessageFields(),
		},
	}
}

func callFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "answered", Type: "boolean"},
		{Name: "business_phone_number", Type: "string"},
		{Name: "customer_city", Type: "string"},
		{Name: "customer_country", Type: "string"},
		{Name: "customer_name", Type: "string"},
		{Name: "customer_phone_number", Type: "string"},
		{Name: "customer_state", Type: "string"},
		{Name: "direction", Type: "string"},
		{Name: "duration", Type: "integer"},
		{Name: "recording", Type: "string"},
		{Name: "start_time", Type: "timestamp"},
		{Name: "tracking_phone_number", Type: "string"},
		{Name: "voicemail", Type: "boolean"},
		{Name: "company_id", Type: "string"},
	}
}

func companyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "time_zone", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "disabled_at", Type: "timestamp"},
		{Name: "dni_active", Type: "boolean"},
		{Name: "callscore_enabled", Type: "boolean"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func textMessageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "company_id", Type: "string"},
		{Name: "initial_tracker_id", Type: "string"},
		{Name: "customer_name", Type: "string"},
		{Name: "customer_phone_number", Type: "string"},
		{Name: "tracking_phone_number", Type: "string"},
		{Name: "last_message_at", Type: "timestamp"},
		{Name: "state", Type: "string"},
	}
}

func callRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                    item["id"],
		"answered":              item["answered"],
		"business_phone_number": item["business_phone_number"],
		"customer_city":         item["customer_city"],
		"customer_country":      item["customer_country"],
		"customer_name":         item["customer_name"],
		"customer_phone_number": item["customer_phone_number"],
		"customer_state":        item["customer_state"],
		"direction":             item["direction"],
		"duration":              item["duration"],
		"recording":             item["recording"],
		"start_time":            item["start_time"],
		"tracking_phone_number": item["tracking_phone_number"],
		"voicemail":             item["voicemail"],
		"company_id":            item["company_id"],
	}
}

func companyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"status":            item["status"],
		"time_zone":         item["time_zone"],
		"created_at":        item["created_at"],
		"disabled_at":       item["disabled_at"],
		"dni_active":        item["dni_active"],
		"callscore_enabled": item["callscore_enabled"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"email":      item["email"],
		"first_name": item["first_name"],
		"last_name":  item["last_name"],
		"name":       item["name"],
		"role":       item["role"],
		"created_at": item["created_at"],
	}
}

func textMessageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                    item["id"],
		"company_id":            item["company_id"],
		"initial_tracker_id":    item["initial_tracker_id"],
		"customer_name":         item["customer_name"],
		"customer_phone_number": item["customer_phone_number"],
		"tracking_phone_number": item["tracking_phone_number"],
		"last_message_at":       item["last_message_at"],
		"state":                 item["state"],
	}
}
