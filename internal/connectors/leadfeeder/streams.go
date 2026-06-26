package leadfeeder

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to how its Leadfeeder JSON:API resource is
// addressed and the record mapper that flattens its objects. Leadfeeder nests
// leads/visits/custom-feeds under an account, so accountScoped endpoints build
// their path from the configured account_id.
type streamEndpoint struct {
	// resource is the trailing path segment (e.g. "accounts", "leads").
	resource string
	// accountScoped is true when the endpoint lives under
	// /accounts/{account_id}/<resource> and therefore needs an account_id config.
	accountScoped bool
	// dateScoped is true when the endpoint accepts/requires start_date & end_date
	// query params (leads and visits).
	dateScoped bool
	// mapRecord flattens a raw JSON:API resource object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// leadfeederStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in leadfeederStreams; the read
// path is fully data-driven from this table.
var leadfeederStreamEndpoints = map[string]streamEndpoint{
	"accounts":     {resource: "accounts", mapRecord: leadfeederAccountRecord},
	"leads":        {resource: "leads", accountScoped: true, dateScoped: true, mapRecord: leadfeederLeadRecord},
	"visits":       {resource: "visits", accountScoped: true, dateScoped: true, mapRecord: leadfeederVisitRecord},
	"custom_feeds": {resource: "custom-feeds", accountScoped: true, mapRecord: leadfeederCustomFeedRecord},
}

// leadfeederStreams returns the connector's published stream catalog. Every
// Leadfeeder JSON:API resource carries a string id, so the primary key is ["id"]
// across the board.
func leadfeederStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "accounts",
			Description: "Leadfeeder accounts available to the API token.",
			PrimaryKey:  []string{"id"},
			Fields:      leadfeederAccountFields(),
		},
		{
			Name:         "leads",
			Description:  "Leads (companies) detected for an account; requires account_id config.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_visit_date"},
			Fields:       leadfeederLeadFields(),
		},
		{
			Name:         "visits",
			Description:  "Individual visits for an account; requires account_id config.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"visit_date"},
			Fields:       leadfeederVisitFields(),
		},
		{
			Name:        "custom_feeds",
			Description: "Custom feeds configured for an account; requires account_id config.",
			PrimaryKey:  []string{"id"},
			Fields:      leadfeederCustomFeedFields(),
		},
	}
}

func leadfeederAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "industry", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "time_zone", Type: "string"},
		{Name: "currency", Type: "string"},
	}
}

func leadfeederLeadFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "industry", Type: "string"},
		{Name: "website", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "employee_count", Type: "string"},
		{Name: "quality", Type: "integer"},
		{Name: "visits", Type: "integer"},
		{Name: "first_visit_date", Type: "string"},
		{Name: "last_visit_date", Type: "string"},
	}
}

func leadfeederVisitFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "visit_length", Type: "integer"},
		{Name: "pageviews", Type: "integer"},
		{Name: "source", Type: "string"},
		{Name: "referring_url", Type: "string"},
		{Name: "hostname", Type: "string"},
		{Name: "visit_date", Type: "string"},
		{Name: "started_at", Type: "string"},
		{Name: "ended_at", Type: "string"},
	}
}

func leadfeederCustomFeedFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
	}
}

func leadfeederAccountRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{"id": item["id"], "type": item["type"]}
	attrs := attributes(item)
	for _, f := range []string{"name", "industry", "status", "time_zone", "currency"} {
		rec[f] = attrs[f]
	}
	return rec
}

func leadfeederLeadRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{"id": item["id"], "type": item["type"]}
	attrs := attributes(item)
	for _, f := range []string{"name", "industry", "website", "city", "country", "employee_count", "quality", "visits", "first_visit_date", "last_visit_date"} {
		rec[f] = attrs[f]
	}
	return rec
}

func leadfeederVisitRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{"id": item["id"], "type": item["type"]}
	attrs := attributes(item)
	for _, f := range []string{"visit_length", "pageviews", "source", "referring_url", "hostname", "visit_date", "started_at", "ended_at"} {
		rec[f] = attrs[f]
	}
	return rec
}

func leadfeederCustomFeedRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{"id": item["id"], "type": item["type"]}
	attrs := attributes(item)
	rec["name"] = attrs["name"]
	return rec
}

// attributes returns the JSON:API "attributes" sub-object of a resource, or an
// empty map if absent. Leadfeeder nests resource fields under attributes.
func attributes(item map[string]any) map[string]any {
	if attrs, ok := item["attributes"].(map[string]any); ok {
		return attrs
	}
	return map[string]any{}
}
