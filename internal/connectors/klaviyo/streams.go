package klaviyo

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Klaviyo API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its JSON:API
// objects. Adding a stream means adding one entry here plus a Stream definition
// in klaviyoStreams; the read path is fully data-driven from this table.
type streamEndpoint struct {
	// resource is the Klaviyo list endpoint path segment (e.g. "profiles").
	resource string
	// mapRecord flattens a raw Klaviyo JSON:API object ({id, type, attributes})
	// into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// klaviyoStreamEndpoints is the per-stream routing table for the CORE set of
// Klaviyo streams supported by this connector.
var klaviyoStreamEndpoints = map[string]streamEndpoint{
	"profiles":  {resource: "profiles", mapRecord: klaviyoProfileRecord},
	"events":    {resource: "events", mapRecord: klaviyoEventRecord},
	"campaigns": {resource: "campaigns", mapRecord: klaviyoCampaignRecord},
	"lists":     {resource: "lists", mapRecord: klaviyoListRecord},
	"metrics":   {resource: "metrics", mapRecord: klaviyoMetricRecord},
	"segments":  {resource: "segments", mapRecord: klaviyoSegmentRecord},
}

// klaviyoStreams returns the connector's published stream catalog. Every Klaviyo
// JSON:API object exposes a string id; most carry attributes.created and
// attributes.updated, so the primary key is ["id"] and the incremental cursor is
// ["updated"] (or ["datetime"] for events) where applicable.
func klaviyoStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "profiles",
			Description:  "Klaviyo customer profiles.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       klaviyoProfileFields(),
		},
		{
			Name:         "events",
			Description:  "Klaviyo events (metric activity per profile).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"datetime"},
			Fields:       klaviyoEventFields(),
		},
		{
			Name:         "campaigns",
			Description:  "Klaviyo email/SMS campaigns.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       klaviyoCampaignFields(),
		},
		{
			Name:         "lists",
			Description:  "Klaviyo lists.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       klaviyoListFields(),
		},
		{
			Name:         "metrics",
			Description:  "Klaviyo metrics.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       klaviyoMetricFields(),
		},
		{
			Name:         "segments",
			Description:  "Klaviyo segments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       klaviyoSegmentFields(),
		},
	}
}

func klaviyoProfileFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone_number", Type: "string"},
		{Name: "external_id", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "organization", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func klaviyoEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "timestamp", Type: "integer"},
		{Name: "datetime", Type: "string"},
		{Name: "uuid", Type: "string"},
	}
}

func klaviyoCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "channel", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "scheduled_at", Type: "string"},
		{Name: "send_time", Type: "string"},
	}
}

func klaviyoListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func klaviyoMetricFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "integration_name", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func klaviyoSegmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "is_processing", Type: "boolean"},
		{Name: "created", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func klaviyoProfileRecord(item map[string]any) connectors.Record {
	attr := attributes(item)
	return connectors.Record{
		"id":           item["id"],
		"type":         item["type"],
		"email":        attr["email"],
		"phone_number": attr["phone_number"],
		"external_id":  attr["external_id"],
		"first_name":   attr["first_name"],
		"last_name":    attr["last_name"],
		"organization": attr["organization"],
		"created":      attr["created"],
		"updated":      attr["updated"],
	}
}

func klaviyoEventRecord(item map[string]any) connectors.Record {
	attr := attributes(item)
	return connectors.Record{
		"id":        item["id"],
		"type":      item["type"],
		"timestamp": attr["timestamp"],
		"datetime":  attr["datetime"],
		"uuid":      attr["uuid"],
	}
}

func klaviyoCampaignRecord(item map[string]any) connectors.Record {
	attr := attributes(item)
	return connectors.Record{
		"id":           item["id"],
		"type":         item["type"],
		"name":         attr["name"],
		"status":       attr["status"],
		"archived":     attr["archived"],
		"channel":      attr["channel"],
		"created_at":   attr["created_at"],
		"updated_at":   attr["updated_at"],
		"scheduled_at": attr["scheduled_at"],
		"send_time":    attr["send_time"],
	}
}

func klaviyoListRecord(item map[string]any) connectors.Record {
	attr := attributes(item)
	return connectors.Record{
		"id":      item["id"],
		"type":    item["type"],
		"name":    attr["name"],
		"created": attr["created"],
		"updated": attr["updated"],
	}
}

func klaviyoMetricRecord(item map[string]any) connectors.Record {
	attr := attributes(item)
	return connectors.Record{
		"id":               item["id"],
		"type":             item["type"],
		"name":             attr["name"],
		"integration_name": integrationName(attr),
		"created":          attr["created"],
		"updated":          attr["updated"],
	}
}

func klaviyoSegmentRecord(item map[string]any) connectors.Record {
	attr := attributes(item)
	return connectors.Record{
		"id":            item["id"],
		"type":          item["type"],
		"name":          attr["name"],
		"is_active":     attr["is_active"],
		"is_processing": attr["is_processing"],
		"created":       attr["created"],
		"updated":       attr["updated"],
	}
}

// attributes returns the nested JSON:API "attributes" object, or an empty map if
// absent. Klaviyo wraps resource fields under attributes; the record mappers
// promote a curated subset to the top level.
func attributes(item map[string]any) map[string]any {
	if attr, ok := item["attributes"].(map[string]any); ok {
		return attr
	}
	return map[string]any{}
}

// integrationName extracts the nested integration display name from a metric's
// attributes.integration.name, returning "" when not present.
func integrationName(attr map[string]any) any {
	integration, ok := attr["integration"].(map[string]any)
	if !ok {
		return nil
	}
	return integration["name"]
}
