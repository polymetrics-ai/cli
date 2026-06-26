package mailosaur

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Mailosaur API resource path (relative
// to base_url), the JSON path where its records live, whether it is paginated by
// page/itemsPerPage, whether it requires a server query param, and the mapper
// that flattens its objects into connectors.Record values.
type streamEndpoint struct {
	// resource is the API path segment, e.g. "messages".
	resource string
	// recordsPath is the dotted path to the array in the response body. Empty
	// (or ".") means the body is itself the array (e.g. /servers).
	recordsPath string
	// paginated is true for endpoints that support page/itemsPerPage.
	paginated bool
	// needsServer is true for endpoints scoped to a single server (messages).
	needsServer bool
	// mapRecord flattens a raw Mailosaur object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// mailosaurStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in mailosaurStreams; the read
// path is fully data-driven from this table.
var mailosaurStreamEndpoints = map[string]streamEndpoint{
	"servers":      {resource: "servers", recordsPath: ".", mapRecord: mailosaurServerRecord},
	"messages":     {resource: "messages", recordsPath: "items", paginated: true, needsServer: true, mapRecord: mailosaurMessageRecord},
	"transactions": {resource: "usage/transactions", recordsPath: "items", mapRecord: mailosaurTransactionRecord},
}

// mailosaurStreams returns the connector's published stream catalog.
func mailosaurStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "servers",
			Description: "Mailosaur virtual SMTP/SMS servers.",
			PrimaryKey:  []string{"id"},
			Fields:      mailosaurServerFields(),
		},
		{
			Name:         "messages",
			Description:  "Email and SMS message summaries for a server (config server=<id>).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"received"},
			Fields:       mailosaurMessageFields(),
		},
		{
			Name:         "transactions",
			Description:  "Account transactional usage over the last 31 days.",
			PrimaryKey:   []string{"timestamp"},
			CursorFields: []string{"timestamp"},
			Fields:       mailosaurTransactionFields(),
		},
	}
}

func mailosaurServerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "users", Type: "array"},
		{Name: "messages", Type: "integer"},
	}
}

func mailosaurMessageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "received", Type: "timestamp"},
		{Name: "type", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "server", Type: "string"},
		{Name: "from", Type: "array"},
		{Name: "to", Type: "array"},
		{Name: "cc", Type: "array"},
		{Name: "bcc", Type: "array"},
	}
}

func mailosaurTransactionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "timestamp", Type: "timestamp"},
		{Name: "email", Type: "integer"},
		{Name: "sms", Type: "integer"},
		{Name: "previews", Type: "integer"},
	}
}

func mailosaurServerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"name":     item["name"],
		"users":    item["users"],
		"messages": item["messages"],
	}
}

func mailosaurMessageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"received": item["received"],
		"type":     item["type"],
		"subject":  item["subject"],
		"server":   item["server"],
		"from":     item["from"],
		"to":       item["to"],
		"cc":       item["cc"],
		"bcc":      item["bcc"],
	}
}

func mailosaurTransactionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"timestamp": item["timestamp"],
		"email":     item["email"],
		"sms":       item["sms"],
		"previews":  item["previews"],
	}
}
