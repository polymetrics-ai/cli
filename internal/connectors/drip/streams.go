package drip

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Drip API resource path segment, the
// JSON key holding its record array, the record mapper, and whether the endpoint
// is account-scoped (prefixed with the account_id path segment) or global.
type streamEndpoint struct {
	// resource is the Drip list endpoint path segment (e.g. "subscribers").
	resource string
	// recordsKey is the JSON object key holding the array of records.
	recordsKey string
	// accountScoped is true when the path is prefixed with the account_id
	// segment (e.g. /{account_id}/subscribers). The /accounts endpoint is the
	// only global one.
	accountScoped bool
	// mapRecord flattens a raw Drip object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// dripStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in dripStreams; the read path
// is fully data-driven from this table.
var dripStreamEndpoints = map[string]streamEndpoint{
	"subscribers": {resource: "subscribers", recordsKey: "subscribers", accountScoped: true, mapRecord: dripSubscriberRecord},
	"campaigns":   {resource: "campaigns", recordsKey: "campaigns", accountScoped: true, mapRecord: dripCampaignRecord},
	"broadcasts":  {resource: "broadcasts", recordsKey: "broadcasts", accountScoped: true, mapRecord: dripBroadcastRecord},
	"accounts":    {resource: "accounts", recordsKey: "accounts", accountScoped: false, mapRecord: dripAccountRecord},
}

// dripStreams returns the connector's published stream catalog. Every Drip
// object exposes a string id and a created_at RFC3339 timestamp, so the primary
// key is ["id"] and the incremental cursor field is ["created_at"].
func dripStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "subscribers",
			Description:  "Drip subscribers (people) for the account.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       dripSubscriberFields(),
		},
		{
			Name:         "campaigns",
			Description:  "Drip email series campaigns.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       dripCampaignFields(),
		},
		{
			Name:         "broadcasts",
			Description:  "Drip single-email broadcast campaigns.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       dripBroadcastFields(),
		},
		{
			Name:         "accounts",
			Description:  "Drip accounts accessible to the API key.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       dripAccountFields(),
		},
	}
}

func dripSubscriberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "time_zone", Type: "string"},
		{Name: "utc_offset", Type: "integer"},
		{Name: "lifetime_value", Type: "integer"},
		{Name: "ip_address", Type: "string"},
		{Name: "user_agent", Type: "string"},
		{Name: "tags", Type: "array"},
		{Name: "custom_fields", Type: "object"},
	}
}

func dripCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "from_name", Type: "string"},
		{Name: "from_email", Type: "string"},
		{Name: "subscriber_count", Type: "integer"},
		{Name: "email_count", Type: "integer"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func dripBroadcastFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "from_name", Type: "string"},
		{Name: "from_email", Type: "string"},
		{Name: "send_at", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func dripAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func dripSubscriberRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"email":          item["email"],
		"status":         item["status"],
		"created_at":     item["created_at"],
		"time_zone":      item["time_zone"],
		"utc_offset":     item["utc_offset"],
		"lifetime_value": item["lifetime_value"],
		"ip_address":     item["ip_address"],
		"user_agent":     item["user_agent"],
		"tags":           item["tags"],
		"custom_fields":  item["custom_fields"],
	}
}

func dripCampaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"status":           item["status"],
		"from_name":        item["from_name"],
		"from_email":       item["from_email"],
		"subscriber_count": item["subscriber_count"],
		"email_count":      item["email_count"],
		"created_at":       item["created_at"],
	}
}

func dripBroadcastRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"status":     item["status"],
		"subject":    item["subject"],
		"from_name":  item["from_name"],
		"from_email": item["from_email"],
		"send_at":    item["send_at"],
		"created_at": item["created_at"],
	}
}

func dripAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"created_at": item["created_at"],
	}
}
