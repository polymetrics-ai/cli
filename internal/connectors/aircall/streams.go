package aircall

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Aircall API resource path (relative
// to base_url), the JSON key holding the array of records in the list response,
// and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Aircall list endpoint path segment (e.g. "calls").
	resource string
	// recordsKey is the top-level JSON key holding the records array. Aircall
	// keys each list by its resource name (e.g. {"calls":[...]}).
	recordsKey string
	// mapRecord flattens a raw Aircall object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// aircallStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in aircallStreams; the read
// path is fully data-driven from this table.
var aircallStreamEndpoints = map[string]streamEndpoint{
	"calls":    {resource: "calls", recordsKey: "calls", mapRecord: aircallCallRecord},
	"users":    {resource: "users", recordsKey: "users", mapRecord: aircallUserRecord},
	"contacts": {resource: "contacts", recordsKey: "contacts", mapRecord: aircallContactRecord},
	"numbers":  {resource: "numbers", recordsKey: "numbers", mapRecord: aircallNumberRecord},
	"teams":    {resource: "teams", recordsKey: "teams", mapRecord: aircallTeamRecord},
}

// aircallStreams returns the connector's published stream catalog. Most Aircall
// objects expose a numeric id and a unix timestamp; the primary key is ["id"]
// and the incremental cursor uses the resource's own timestamp where the API
// supports a `from` filter (calls/contacts use started_at/created_at).
func aircallStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "calls",
			Description:  "Aircall calls (inbound and outbound).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"started_at"},
			Fields:       aircallCallFields(),
		},
		{
			Name:         "users",
			Description:  "Aircall users (agents).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       aircallUserFields(),
		},
		{
			Name:         "contacts",
			Description:  "Aircall contacts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       aircallContactFields(),
		},
		{
			Name:        "numbers",
			Description: "Aircall phone numbers.",
			PrimaryKey:  []string{"id"},
			Fields:      aircallNumberFields(),
		},
		{
			Name:        "teams",
			Description: "Aircall teams.",
			PrimaryKey:  []string{"id"},
			Fields:      aircallTeamFields(),
		},
	}
}

func aircallCallFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "sid", Type: "string"},
		{Name: "direction", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "started_at", Type: "integer"},
		{Name: "answered_at", Type: "integer"},
		{Name: "ended_at", Type: "integer"},
		{Name: "duration", Type: "integer"},
		{Name: "raw_digits", Type: "string"},
		{Name: "missed_call_reason", Type: "string"},
		{Name: "recording", Type: "string"},
		{Name: "voicemail", Type: "string"},
		{Name: "archived", Type: "boolean"},
	}
}

func aircallUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "available", Type: "boolean"},
		{Name: "availability_status", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "time_zone", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "wrap_up_time", Type: "integer"},
	}
}

func aircallContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "company_name", Type: "string"},
		{Name: "information", Type: "string"},
		{Name: "is_shared", Type: "boolean"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func aircallNumberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "digits", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "time_zone", Type: "string"},
		{Name: "open", Type: "boolean"},
		{Name: "is_ivr", Type: "boolean"},
		{Name: "live_recording_activated", Type: "boolean"},
		{Name: "created_at", Type: "string"},
	}
}

func aircallTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func aircallCallRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"sid":                item["sid"],
		"direction":          item["direction"],
		"status":             item["status"],
		"started_at":         item["started_at"],
		"answered_at":        item["answered_at"],
		"ended_at":           item["ended_at"],
		"duration":           item["duration"],
		"raw_digits":         item["raw_digits"],
		"missed_call_reason": item["missed_call_reason"],
		"recording":          item["recording"],
		"voicemail":          item["voicemail"],
		"archived":           item["archived"],
	}
}

func aircallUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"name":                item["name"],
		"email":               item["email"],
		"available":           item["available"],
		"availability_status": item["availability_status"],
		"created_at":          item["created_at"],
		"time_zone":           item["time_zone"],
		"language":            item["language"],
		"wrap_up_time":        item["wrap_up_time"],
	}
}

func aircallContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"first_name":   item["first_name"],
		"last_name":    item["last_name"],
		"company_name": item["company_name"],
		"information":  item["information"],
		"is_shared":    item["is_shared"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func aircallNumberRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                       item["id"],
		"name":                     item["name"],
		"digits":                   item["digits"],
		"country":                  item["country"],
		"time_zone":                item["time_zone"],
		"open":                     item["open"],
		"is_ivr":                   item["is_ivr"],
		"live_recording_activated": item["live_recording_activated"],
		"created_at":               item["created_at"],
	}
}

func aircallTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"created_at": item["created_at"],
	}
}
