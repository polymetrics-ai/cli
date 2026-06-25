package braze

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Braze list export endpoint it reads
// from, the top-level JSON field that holds the records array, and the mapper
// that flattens each object into a connectors.Record.
//
// Braze list endpoints share one shape: GET <resource> returns
// {"message":"success", "<recordsField>":[ ... ]} and paginate via a 0-based
// ?page= parameter, ~100 items per page (a short page ends pagination).
type streamEndpoint struct {
	resource     string
	recordsField string
	mapRecord    func(map[string]any) connectors.Record
}

// brazeStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in brazeStreams; the read path
// is fully data-driven from this table.
var brazeStreamEndpoints = map[string]streamEndpoint{
	"campaigns": {resource: "campaigns/list", recordsField: "campaigns", mapRecord: brazeCampaignRecord},
	"canvases":  {resource: "canvas/list", recordsField: "canvases", mapRecord: brazeCanvasRecord},
	"segments":  {resource: "segments/list", recordsField: "segments", mapRecord: brazeSegmentRecord},
	"events":    {resource: "events/list", recordsField: "events", mapRecord: brazeEventRecord},
}

// brazeStreams returns the connector's published stream catalog. Braze list
// objects expose a string id and a last_edited timestamp (events expose only a
// name), so the primary key is ["id"] and the incremental cursor is
// ["last_edited"] where available.
func brazeStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "campaigns",
			Description:  "Braze campaigns (campaign list export).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_edited"},
			Fields:       brazeCampaignFields(),
		},
		{
			Name:         "canvases",
			Description:  "Braze Canvases (canvas list export).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_edited"},
			Fields:       brazeCanvasFields(),
		},
		{
			Name:         "segments",
			Description:  "Braze segments (segment list export).",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       brazeSegmentFields(),
		},
		{
			Name:         "events",
			Description:  "Braze custom events (event list export).",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       brazeEventFields(),
		},
	}
}

func brazeCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "is_api_campaign", Type: "boolean"},
		{Name: "tags", Type: "array"},
		{Name: "last_edited", Type: "string"},
	}
}

func brazeCanvasFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "tags", Type: "array"},
		{Name: "last_edited", Type: "string"},
	}
}

func brazeSegmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "analytics_tracking_enabled", Type: "boolean"},
		{Name: "tags", Type: "array"},
	}
}

func brazeEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
	}
}

func brazeCampaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            item["name"],
		"is_api_campaign": item["is_api_campaign"],
		"tags":            item["tags"],
		"last_edited":     item["last_edited"],
	}
}

func brazeCanvasRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"tags":        item["tags"],
		"last_edited": item["last_edited"],
	}
}

func brazeSegmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                         item["id"],
		"name":                       item["name"],
		"analytics_tracking_enabled": item["analytics_tracking_enabled"],
		"tags":                       item["tags"],
	}
}

// brazeEventRecord handles the /events/list shape. Braze returns events as an
// array of event-name strings; RecordsAt only yields object elements, so the
// read path wraps each string into {"name": <string>} before mapping. The id is
// derived from the name (events have no separate id).
func brazeEventRecord(item map[string]any) connectors.Record {
	name := item["name"]
	id := item["id"]
	if id == nil {
		id = name
	}
	return connectors.Record{
		"id":   id,
		"name": name,
	}
}
