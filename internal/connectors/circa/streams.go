package circa

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Circa API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Circa list endpoint path segment (e.g. "events").
	resource string
	// incremental reports whether the stream supports updated_at[min] filtering.
	incremental bool
	// mapRecord flattens a raw Circa object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// circaStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in circaStreams; the read path
// is fully data-driven from this table.
var circaStreamEndpoints = map[string]streamEndpoint{
	"events":    {resource: "events", incremental: true, mapRecord: circaEventRecord},
	"contacts":  {resource: "contacts", incremental: true, mapRecord: circaContactRecord},
	"companies": {resource: "companies", incremental: true, mapRecord: circaCompanyRecord},
	"teams":     {resource: "teams", incremental: false, mapRecord: circaTeamRecord},
}

// circaStreams returns the connector's published stream catalog. Every Circa
// object exposes a string id; events/contacts/companies additionally carry an
// updated_at RFC3339 timestamp used as the incremental cursor.
func circaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "events",
			Description:  "Circa events.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       circaEventFields(),
		},
		{
			Name:         "contacts",
			Description:  "Circa contacts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       circaContactFields(),
		},
		{
			Name:         "companies",
			Description:  "Circa companies.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       circaCompanyFields(),
		},
		{
			Name:        "teams",
			Description: "Circa teams.",
			PrimaryKey:  []string{"id"},
			Fields:      circaTeamFields(),
		},
	}
}

func circaEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "website", Type: "string"},
		{Name: "brief_url", Type: "string"},
		{Name: "time_zone", Type: "string"},
		{Name: "paid_total", Type: "number"},
		{Name: "actual_total", Type: "number"},
		{Name: "planned_total", Type: "number"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func circaContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "company", Type: "object"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func circaCompanyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email_opt_in", Type: "boolean"},
		{Name: "sync_status", Type: "object"},
		{Name: "created_method", Type: "string"},
		{Name: "updated_method", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func circaTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "created_by", Type: "object"},
	}
}

func circaEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"status":        item["status"],
		"website":       item["website"],
		"brief_url":     item["brief_url"],
		"time_zone":     item["time_zone"],
		"paid_total":    item["paid_total"],
		"actual_total":  item["actual_total"],
		"planned_total": item["planned_total"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func circaContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"email":      item["email"],
		"first_name": item["first_name"],
		"last_name":  item["last_name"],
		"company":    item["company"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func circaCompanyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"email_opt_in":   item["email_opt_in"],
		"sync_status":    item["sync_status"],
		"created_method": item["created_method"],
		"updated_method": item["updated_method"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
	}
}

func circaTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"created_at": item["created_at"],
		"created_by": item["created_by"],
	}
}
