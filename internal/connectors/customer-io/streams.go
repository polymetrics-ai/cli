package customerio

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Customer.io App API resource path
// (relative to base_url), the JSON field path holding the records array, and the
// record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the App API list endpoint path segment (e.g. "campaigns").
	resource string
	// recordsPath is the dotted JSON path to the records array in the response
	// body (e.g. "campaigns").
	recordsPath string
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// customerIOStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in customerIOStreams; the
// read path is fully data-driven from this table.
var customerIOStreamEndpoints = map[string]streamEndpoint{
	"campaigns":   {resource: "campaigns", recordsPath: "campaigns", mapRecord: campaignRecord},
	"newsletters": {resource: "newsletters", recordsPath: "newsletters", mapRecord: newsletterRecord},
	"segments":    {resource: "segments", recordsPath: "segments", mapRecord: segmentRecord},
	"broadcasts":  {resource: "broadcasts", recordsPath: "broadcasts", mapRecord: broadcastRecord},
}

// customerIOStreams returns the connector's published stream catalog. Every
// Customer.io App API object exposes a numeric id and an `updated` unix
// timestamp, so the primary key is ["id"] and the incremental cursor field is
// ["updated"] across the board.
func customerIOStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "campaigns",
			Description:  "Customer.io campaigns.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       campaignFields(),
		},
		{
			Name:         "newsletters",
			Description:  "Customer.io newsletters.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       newsletterFields(),
		},
		{
			Name:         "segments",
			Description:  "Customer.io segments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       segmentFields(),
		},
		{
			Name:         "broadcasts",
			Description:  "Customer.io broadcasts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       broadcastFields(),
		},
	}
}

func campaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "created", Type: "integer"},
		{Name: "updated", Type: "integer"},
	}
}

func newsletterFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "created", Type: "integer"},
		{Name: "updated", Type: "integer"},
	}
}

func segmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "updated", Type: "integer"},
	}
}

func broadcastFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "created", Type: "integer"},
		{Name: "updated", Type: "integer"},
	}
}

func campaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      item["id"],
		"name":    item["name"],
		"type":    item["type"],
		"state":   item["state"],
		"active":  item["active"],
		"created": item["created"],
		"updated": item["updated"],
	}
}

func newsletterRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      item["id"],
		"name":    item["name"],
		"type":    item["type"],
		"subject": item["subject"],
		"created": item["created"],
		"updated": item["updated"],
	}
}

func segmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"state":       item["state"],
		"type":        item["type"],
		"updated":     item["updated"],
	}
}

func broadcastRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      item["id"],
		"name":    item["name"],
		"type":    item["type"],
		"state":   item["state"],
		"active":  item["active"],
		"created": item["created"],
		"updated": item["updated"],
	}
}
