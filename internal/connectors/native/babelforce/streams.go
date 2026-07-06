package babelforce

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Babelforce API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
// Every Babelforce reporting/list endpoint returns the same envelope:
// {"items":[...],"pagination":{"current":N,"max":M}} extracted at "items".
type streamEndpoint struct {
	// resource is the API path segment (relative to base_url), e.g.
	// "calls/reporting/simple".
	resource string
	// mapRecord flattens a raw Babelforce object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// babelforceStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in babelforceStreams; the
// read path is fully data-driven from this table.
//
// "calls" is the stream verified against the upstream upstream manifest
// (/calls/reporting/simple). The remaining streams target sibling Babelforce v2
// list endpoints that share the identical {items,pagination} envelope and
// dual-header auth.
var babelforceStreamEndpoints = map[string]streamEndpoint{
	"calls":          {resource: "calls/reporting/simple", mapRecord: babelforceCallRecord},
	"calls_extended": {resource: "calls/reporting/extended", mapRecord: babelforceCallRecord},
	"recordings":     {resource: "recordings", mapRecord: babelforceRecordingRecord},
	"numbers":        {resource: "numbers", mapRecord: babelforceNumberRecord},
	"users":          {resource: "users", mapRecord: babelforceUserRecord},
}

// babelforceStreams returns the connector's published stream catalog. Babelforce
// objects expose a string id and a unix-seconds dateCreated cursor, so the
// primary key is ["id"] and the incremental cursor field is ["dateCreated"].
func babelforceStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "calls",
			Description:  "Babelforce call reporting records (simple report).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"dateCreated"},
			Fields:       babelforceCallFields(),
		},
		{
			Name:         "calls_extended",
			Description:  "Babelforce call reporting records (extended report).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"dateCreated"},
			Fields:       babelforceCallFields(),
		},
		{
			Name:         "recordings",
			Description:  "Babelforce call recordings.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"dateCreated"},
			Fields:       babelforceRecordingFields(),
		},
		{
			Name:         "numbers",
			Description:  "Babelforce phone numbers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"dateCreated"},
			Fields:       babelforceNumberFields(),
		},
		{
			Name:         "users",
			Description:  "Babelforce users/agents.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"dateCreated"},
			Fields:       babelforceUserFields(),
		},
	}
}

func babelforceCallFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "domain", Type: "string"},
		{Name: "from", Type: "string"},
		{Name: "to", Type: "string"},
		{Name: "anonymous", Type: "boolean"},
		{Name: "duration", Type: "integer"},
		{Name: "finishReason", Type: "string"},
		{Name: "conversationId", Type: "string"},
		{Name: "sessionId", Type: "string"},
		{Name: "parentId", Type: "string"},
		{Name: "dateCreated", Type: "string"},
		{Name: "dateEstablished", Type: "string"},
		{Name: "dateFinished", Type: "string"},
		{Name: "lastUpdated", Type: "string"},
	}
}

func babelforceRecordingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "duration", Type: "integer"},
		{Name: "url", Type: "string"},
		{Name: "dateCreated", Type: "string"},
		{Name: "lastUpdated", Type: "string"},
	}
}

func babelforceNumberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "dateCreated", Type: "string"},
		{Name: "lastUpdated", Type: "string"},
	}
}

func babelforceUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "dateCreated", Type: "string"},
		{Name: "lastUpdated", Type: "string"},
	}
}

func babelforceCallRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"type":            item["type"],
		"state":           item["state"],
		"source":          item["source"],
		"domain":          item["domain"],
		"from":            item["from"],
		"to":              item["to"],
		"anonymous":       item["anonymous"],
		"duration":        item["duration"],
		"finishReason":    item["finishReason"],
		"conversationId":  item["conversationId"],
		"sessionId":       item["sessionId"],
		"parentId":        item["parentId"],
		"dateCreated":     item["dateCreated"],
		"dateEstablished": item["dateEstablished"],
		"dateFinished":    item["dateFinished"],
		"lastUpdated":     item["lastUpdated"],
	}
}

func babelforceRecordingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"state":       item["state"],
		"duration":    item["duration"],
		"url":         item["url"],
		"dateCreated": item["dateCreated"],
		"lastUpdated": item["lastUpdated"],
	}
}

func babelforceNumberRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"number":      item["number"],
		"name":        item["name"],
		"state":       item["state"],
		"dateCreated": item["dateCreated"],
		"lastUpdated": item["lastUpdated"],
	}
}

func babelforceUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"number":      item["number"],
		"state":       item["state"],
		"dateCreated": item["dateCreated"],
		"lastUpdated": item["lastUpdated"],
	}
}
