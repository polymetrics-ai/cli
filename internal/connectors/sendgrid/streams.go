package sendgrid

import "polymetrics/internal/connectors"

// paginationStyle selects how a stream walks its pages.
type paginationStyle int

const (
	// metadataNext follows SendGrid's marketing-API _metadata.next full-URL
	// cursor links (lists, segments, contacts).
	metadataNext paginationStyle = iota
	// offsetLimit advances with limit/offset until a short page is returned
	// (suppression endpoints that return a top-level array).
	offsetLimit
)

// streamEndpoint maps a stream name to the SendGrid API resource path (relative
// to base_url), the JSON path its records live under, the pagination style, and
// the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "marketing/lists").
	resource string
	// recordsPath is the dotted path to the array of records in the response
	// body ("result", "results", or "" for a top-level array).
	recordsPath string
	pagination  paginationStyle
	mapRecord   func(map[string]any) connectors.Record
}

// sendgridStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in sendgridStreams; the read
// path is fully data-driven from this table.
var sendgridStreamEndpoints = map[string]streamEndpoint{
	"lists":               {resource: "marketing/lists", recordsPath: "result", pagination: metadataNext, mapRecord: sendgridListRecord},
	"segments":            {resource: "marketing/segments/2.0", recordsPath: "results", pagination: metadataNext, mapRecord: sendgridSegmentRecord},
	"contacts":            {resource: "marketing/contacts", recordsPath: "result", pagination: metadataNext, mapRecord: sendgridContactRecord},
	"suppression_bounces": {resource: "suppression/bounces", recordsPath: "", pagination: offsetLimit, mapRecord: sendgridBounceRecord},
}

// sendgridStreams returns the connector's published stream catalog.
func sendgridStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "lists",
			Description:  "SendGrid Marketing Campaigns contact lists.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       sendgridListFields(),
		},
		{
			Name:         "segments",
			Description:  "SendGrid Marketing Campaigns segments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       sendgridSegmentFields(),
		},
		{
			Name:         "contacts",
			Description:  "SendGrid Marketing Campaigns contacts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       sendgridContactFields(),
		},
		{
			Name:         "suppression_bounces",
			Description:  "SendGrid suppression bounces.",
			PrimaryKey:   []string{"email"},
			CursorFields: []string{"created"},
			Fields:       sendgridBounceFields(),
		},
	}
}

func sendgridListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "contact_count", Type: "integer"},
	}
}

func sendgridSegmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "contacts_count", Type: "integer"},
		{Name: "query_version", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "sample_updated_at", Type: "string"},
	}
}

func sendgridContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "phone_number", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func sendgridBounceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "email", Type: "string"},
		{Name: "created", Type: "integer"},
		{Name: "reason", Type: "string"},
		{Name: "status", Type: "string"},
	}
}

func sendgridListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"contact_count": item["contact_count"],
	}
}

func sendgridSegmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"contacts_count":    item["contacts_count"],
		"query_version":     item["query_version"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
		"sample_updated_at": item["sample_updated_at"],
	}
}

func sendgridContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"email":        item["email"],
		"first_name":   item["first_name"],
		"last_name":    item["last_name"],
		"phone_number": item["phone_number"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func sendgridBounceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"email":   item["email"],
		"created": item["created"],
		"reason":  item["reason"],
		"status":  item["status"],
	}
}
