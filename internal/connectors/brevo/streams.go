package brevo

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Brevo API resource path (relative to
// base_url), the JSON path to the records array in the response body, and the
// record mapper that flattens its objects. supportsModifiedSince marks streams
// whose list endpoint accepts the modifiedSince incremental filter.
type streamEndpoint struct {
	// resource is the Brevo list endpoint path segment (e.g. "contacts").
	resource string
	// recordsPath is the dotted JSON path to the array of records in the
	// response body (e.g. "contacts", "campaigns", "lists", "senders").
	recordsPath string
	// paginated is true when the endpoint accepts limit/offset pagination.
	paginated bool
	// supportsModifiedSince is true when the endpoint accepts a modifiedSince
	// incremental cursor query parameter.
	supportsModifiedSince bool
	// cursorField is the response field carrying the incremental cursor value.
	cursorField string
	// mapRecord flattens a raw Brevo object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// brevoStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in brevoStreams; the read path
// is fully data-driven from this table.
var brevoStreamEndpoints = map[string]streamEndpoint{
	"contacts": {
		resource:              "contacts",
		recordsPath:           "contacts",
		paginated:             true,
		supportsModifiedSince: true,
		cursorField:           "modifiedAt",
		mapRecord:             brevoContactRecord,
	},
	"emailCampaigns": {
		resource:              "emailCampaigns",
		recordsPath:           "campaigns",
		paginated:             true,
		supportsModifiedSince: true,
		cursorField:           "modifiedAt",
		mapRecord:             brevoCampaignRecord,
	},
	"contacts_lists": {
		resource:    "contacts/lists",
		recordsPath: "lists",
		paginated:   true,
		mapRecord:   brevoListRecord,
	},
	"senders": {
		resource:    "senders",
		recordsPath: "senders",
		paginated:   false,
		mapRecord:   brevoSenderRecord,
	},
}

// brevoStreams returns the connector's published stream catalog.
func brevoStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "Brevo contacts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modifiedAt"},
			Fields:       brevoContactFields(),
		},
		{
			Name:         "emailCampaigns",
			Description:  "Brevo email campaigns.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"modifiedAt"},
			Fields:       brevoCampaignFields(),
		},
		{
			Name:        "contacts_lists",
			Description: "Brevo contact lists.",
			PrimaryKey:  []string{"id"},
			Fields:      brevoListFields(),
		},
		{
			Name:        "senders",
			Description: "Brevo senders.",
			PrimaryKey:  []string{"id"},
			Fields:      brevoSenderFields(),
		},
	}
}

func brevoContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "email", Type: "string"},
		{Name: "emailBlacklisted", Type: "boolean"},
		{Name: "smsBlacklisted", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
		{Name: "modifiedAt", Type: "string"},
		{Name: "listIds", Type: "array"},
		{Name: "attributes", Type: "object"},
	}
}

func brevoCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "modifiedAt", Type: "string"},
	}
}

func brevoListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "totalSubscribers", Type: "integer"},
		{Name: "totalBlacklisted", Type: "integer"},
		{Name: "uniqueSubscribers", Type: "integer"},
		{Name: "folderId", Type: "integer"},
	}
}

func brevoSenderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "active", Type: "boolean"},
	}
}

func brevoContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"email":            item["email"],
		"emailBlacklisted": item["emailBlacklisted"],
		"smsBlacklisted":   item["smsBlacklisted"],
		"createdAt":        item["createdAt"],
		"modifiedAt":       item["modifiedAt"],
		"listIds":          item["listIds"],
		"attributes":       item["attributes"],
	}
}

func brevoCampaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"subject":    item["subject"],
		"type":       item["type"],
		"status":     item["status"],
		"createdAt":  item["createdAt"],
		"modifiedAt": item["modifiedAt"],
	}
}

func brevoListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"totalSubscribers":  item["totalSubscribers"],
		"totalBlacklisted":  item["totalBlacklisted"],
		"uniqueSubscribers": item["uniqueSubscribers"],
		"folderId":          item["folderId"],
	}
}

func brevoSenderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":     item["id"],
		"name":   item["name"],
		"email":  item["email"],
		"active": item["active"],
	}
}
