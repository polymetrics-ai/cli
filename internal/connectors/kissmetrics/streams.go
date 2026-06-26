package kissmetrics

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Kissmetrics API resource and its
// record mapper. nested marks streams that live under a product partition
// (products/{product_id}/<resource>) and therefore require a product_id config.
type streamEndpoint struct {
	// resource is the path segment read for top-level streams, or the trailing
	// segment for nested streams (e.g. "events" -> products/{id}/events).
	resource string
	// nested is true when the stream is scoped to a single product partition.
	nested bool
	// mapRecord flattens a raw Kissmetrics object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// kissmetricsStreamEndpoints is the per-stream routing table. The read path is
// fully data-driven from this table: products is a top-level collection, while
// reports/events/properties are nested under a product partition.
var kissmetricsStreamEndpoints = map[string]streamEndpoint{
	"products":   {resource: "products", nested: false, mapRecord: kissmetricsProductRecord},
	"reports":    {resource: "reports", nested: true, mapRecord: kissmetricsReportRecord},
	"events":     {resource: "events", nested: true, mapRecord: kissmetricsEventRecord},
	"properties": {resource: "properties", nested: true, mapRecord: kissmetricsPropertyRecord},
}

// kissmetricsStreams returns the connector's published stream catalog. Every
// Kissmetrics object exposes a string id, so the primary key is ["id"]. The API
// supports full refresh only, so there are no incremental cursor fields.
func kissmetricsStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "products",
			Description: "Kissmetrics products (accounts) accessible to the user.",
			PrimaryKey:  []string{"id"},
			Fields:      kissmetricsProductFields(),
		},
		{
			Name:        "reports",
			Description: "Saved reports for a Kissmetrics product (requires product_id).",
			PrimaryKey:  []string{"id"},
			Fields:      kissmetricsReportFields(),
		},
		{
			Name:        "events",
			Description: "Tracked events for a Kissmetrics product (requires product_id).",
			PrimaryKey:  []string{"id"},
			Fields:      kissmetricsEventFields(),
		},
		{
			Name:        "properties",
			Description: "Properties for a Kissmetrics product (requires product_id).",
			PrimaryKey:  []string{"id"},
			Fields:      kissmetricsPropertyFields(),
		},
	}
}

func kissmetricsProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func kissmetricsReportFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "product_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func kissmetricsEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "product_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func kissmetricsPropertyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "product_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func kissmetricsProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func kissmetricsReportRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"product_id": item["product_id"],
		"name":       item["name"],
		"type":       item["type"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func kissmetricsEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"product_id":   item["product_id"],
		"name":         item["name"],
		"display_name": item["display_name"],
		"created_at":   item["created_at"],
	}
}

func kissmetricsPropertyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"product_id":   item["product_id"],
		"name":         item["name"],
		"display_name": item["display_name"],
		"type":         item["type"],
		"created_at":   item["created_at"],
	}
}
