package agilecrm

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the AgileCRM API resource path (relative
// to base_url) it reads from, whether that resource paginates via AgileCRM's
// last-record cursor, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the AgileCRM list endpoint path segment (e.g. "contacts").
	resource string
	// paginated is true when the endpoint supports cursor/page_size paging.
	paginated bool
	// mapRecord flattens a raw AgileCRM object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// agilecrmStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in agilecrmStreams; the read
// path is fully data-driven from this table.
//
// Paths and pagination semantics come from the official REST API docs
// (github.com/agilecrm/rest-api) and the upstream source manifest: contacts and
// deals (opportunity) paginate via a cursor read off the last array element;
// tasks and milestone pipelines return a bounded top-level array with no paging.
var agilecrmStreamEndpoints = map[string]streamEndpoint{
	"contacts":  {resource: "contacts", paginated: true, mapRecord: agilecrmContactRecord},
	"deals":     {resource: "opportunity", paginated: true, mapRecord: agilecrmDealRecord},
	"tasks":     {resource: "tasks", paginated: false, mapRecord: agilecrmTaskRecord},
	"milestone": {resource: "milestone/pipelines", paginated: false, mapRecord: agilecrmMilestoneRecord},
}

// agilecrmStreams returns the connector's published stream catalog. Every
// AgileCRM object exposes a numeric id; contacts/deals/tasks also carry a
// numeric created_time used as the incremental cursor field (the live API only
// supports full refresh today, but the field is published for downstream use).
func agilecrmStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "AgileCRM contacts (people and the companies they belong to).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_time"},
			Fields:       agilecrmContactFields(),
		},
		{
			Name:         "deals",
			Description:  "AgileCRM deals (sales opportunities).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_time"},
			Fields:       agilecrmDealFields(),
		},
		{
			Name:         "tasks",
			Description:  "AgileCRM tasks (action items and to-dos).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_time"},
			Fields:       agilecrmTaskFields(),
		},
		{
			Name:        "milestone",
			Description: "AgileCRM deal milestone pipelines.",
			PrimaryKey:  []string{"id"},
			Fields:      agilecrmMilestoneFields(),
		},
	}
}

func agilecrmContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "type", Type: "string"},
		{Name: "created_time", Type: "integer"},
		{Name: "updated_time", Type: "integer"},
		{Name: "star_value", Type: "integer"},
		{Name: "lead_score", Type: "integer"},
		{Name: "tags", Type: "array"},
		{Name: "properties", Type: "array"},
		{Name: "owner_id", Type: "string"},
	}
}

func agilecrmDealFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "created_time", Type: "integer"},
		{Name: "expected_value", Type: "number"},
		{Name: "probability", Type: "integer"},
		{Name: "milestone", Type: "string"},
		{Name: "close_date", Type: "integer"},
		{Name: "pipeline_id", Type: "integer"},
		{Name: "owner_id", Type: "string"},
	}
}

func agilecrmTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "subject", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "priority_type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_time", Type: "integer"},
		{Name: "due", Type: "integer"},
		{Name: "is_complete", Type: "boolean"},
		{Name: "owner_id", Type: "string"},
	}
}

func agilecrmMilestoneFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "milestones", Type: "string"},
		{Name: "pipeline_default", Type: "boolean"},
	}
}

func agilecrmContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"type":         item["type"],
		"created_time": item["created_time"],
		"updated_time": item["updated_time"],
		"star_value":   item["star_value"],
		"lead_score":   item["lead_score"],
		"tags":         item["tags"],
		"properties":   item["properties"],
		"owner_id":     ownerID(item),
	}
}

func agilecrmDealRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"created_time":   item["created_time"],
		"expected_value": item["expected_value"],
		"probability":    item["probability"],
		"milestone":      item["milestone"],
		"close_date":     item["close_date"],
		"pipeline_id":    item["pipeline_id"],
		"owner_id":       ownerID(item),
	}
}

func agilecrmTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"subject":       item["subject"],
		"type":          item["type"],
		"priority_type": item["priority_type"],
		"status":        item["status"],
		"created_time":  item["created_time"],
		"due":           item["due"],
		"is_complete":   item["is_complete"],
		"owner_id":      ownerID(item),
	}
}

func agilecrmMilestoneRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"milestones":       item["milestones"],
		"pipeline_default": item["pipeline_default"],
	}
}

// ownerID extracts a stable owner identifier. AgileCRM nests the owner as an
// object ({"owner":{"id":...}}); fall back to a top-level owner_id when present.
func ownerID(item map[string]any) any {
	if owner, ok := item["owner"].(map[string]any); ok {
		if id, ok := owner["id"]; ok {
			return id
		}
	}
	return item["owner_id"]
}
