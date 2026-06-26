package dropboxsign

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to its Dropbox Sign API resource path
// (relative to base_url), the JSON path to the records array in the response,
// and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "signature_request/list").
	resource string
	// recordsPath is the dotted JSON path to the array of records (or "account"
	// for the single-object account endpoint).
	recordsPath string
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// dropboxSignStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in dropboxSignStreams; the
// read path is fully data-driven from this table.
var dropboxSignStreamEndpoints = map[string]streamEndpoint{
	"signature_requests": {resource: "signature_request/list", recordsPath: "signature_requests", mapRecord: signatureRequestRecord},
	"templates":          {resource: "template/list", recordsPath: "templates", mapRecord: templateRecord},
	"team_members":       {resource: "team/members", recordsPath: "team_members", mapRecord: teamMemberRecord},
	"account":            {resource: "account", recordsPath: "account", mapRecord: accountRecord},
}

// dropboxSignStreams returns the connector's published stream catalog.
func dropboxSignStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "signature_requests",
			Description:  "Dropbox Sign signature requests.",
			PrimaryKey:   []string{"signature_request_id"},
			CursorFields: []string{"created_at"},
			Fields:       signatureRequestFields(),
		},
		{
			Name:         "templates",
			Description:  "Dropbox Sign reusable templates.",
			PrimaryKey:   []string{"template_id"},
			CursorFields: []string{"updated_at"},
			Fields:       templateFields(),
		},
		{
			Name:        "team_members",
			Description: "Members of the Dropbox Sign team.",
			PrimaryKey:  []string{"account_id"},
			Fields:      teamMemberFields(),
		},
		{
			Name:        "account",
			Description: "The authenticated Dropbox Sign account.",
			PrimaryKey:  []string{"account_id"},
			Fields:      accountFields(),
		},
	}
}

func signatureRequestFields() []connectors.Field {
	return []connectors.Field{
		{Name: "signature_request_id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "message", Type: "string"},
		{Name: "is_complete", Type: "boolean"},
		{Name: "is_declined", Type: "boolean"},
		{Name: "has_error", Type: "boolean"},
		{Name: "test_mode", Type: "boolean"},
		{Name: "requester_email_address", Type: "string"},
		{Name: "created_at", Type: "integer"},
	}
}

func templateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "template_id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "message", Type: "string"},
		{Name: "is_creator", Type: "boolean"},
		{Name: "is_embedded", Type: "boolean"},
		{Name: "is_locked", Type: "boolean"},
		{Name: "updated_at", Type: "integer"},
	}
}

func teamMemberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "account_id", Type: "string"},
		{Name: "email_address", Type: "string"},
		{Name: "role", Type: "string"},
	}
}

func accountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "account_id", Type: "string"},
		{Name: "email_address", Type: "string"},
		{Name: "is_paid_hs", Type: "boolean"},
		{Name: "is_paid_hf", Type: "boolean"},
		{Name: "role_code", Type: "string"},
		{Name: "locale", Type: "string"},
	}
}

func signatureRequestRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"signature_request_id":    item["signature_request_id"],
		"title":                   item["title"],
		"subject":                 item["subject"],
		"message":                 item["message"],
		"is_complete":             item["is_complete"],
		"is_declined":             item["is_declined"],
		"has_error":               item["has_error"],
		"test_mode":               item["test_mode"],
		"requester_email_address": item["requester_email_address"],
		"created_at":              item["created_at"],
	}
}

func templateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"template_id": item["template_id"],
		"title":       item["title"],
		"message":     item["message"],
		"is_creator":  item["is_creator"],
		"is_embedded": item["is_embedded"],
		"is_locked":   item["is_locked"],
		"updated_at":  item["updated_at"],
	}
}

func teamMemberRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"account_id":    item["account_id"],
		"email_address": item["email_address"],
		"role":          item["role"],
	}
}

func accountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"account_id":    item["account_id"],
		"email_address": item["email_address"],
		"is_paid_hs":    item["is_paid_hs"],
		"is_paid_hf":    item["is_paid_hf"],
		"role_code":     item["role_code"],
		"locale":        item["locale"],
	}
}
