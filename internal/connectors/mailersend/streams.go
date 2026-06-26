package mailersend

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the MailerSend API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
//
// requiresDomain marks the activity stream, whose path is templated with the
// configured domain_id (activity/{domain_id}). requiresDateRange marks streams
// that mandate a date_from/date_to window (also activity).
type streamEndpoint struct {
	// resource is the list endpoint path segment (e.g. "domains"). For the
	// activity stream it is the prefix; the domain id is appended at read time.
	resource string
	// requiresDomain templates resource as "<resource>/<domain_id>".
	requiresDomain bool
	// requiresDateRange requires date_from/date_to to be supplied.
	requiresDateRange bool
	// mapRecord flattens a raw MailerSend object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// mailersendStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in mailersendStreams; the
// read path is fully data-driven from this table.
var mailersendStreamEndpoints = map[string]streamEndpoint{
	"activity":   {resource: "activity", requiresDomain: true, requiresDateRange: true, mapRecord: mailersendActivityRecord},
	"domains":    {resource: "domains", mapRecord: mailersendDomainRecord},
	"messages":   {resource: "messages", mapRecord: mailersendMessageRecord},
	"recipients": {resource: "recipients", mapRecord: mailersendRecipientRecord},
}

// mailersendStreams returns the connector's published stream catalog. Every
// MailerSend object exposes a string id and RFC3339-style created_at/updated_at
// timestamps, so the primary key is ["id"] and the cursor field is
// ["updated_at"] (or ["created_at"] for activity, which has no update).
func mailersendStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "activity",
			Description:  "MailerSend email activity events for a domain (sent, delivered, opened, clicked, ...).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       mailersendActivityFields(),
		},
		{
			Name:         "domains",
			Description:  "MailerSend sending domains and their verification state.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       mailersendDomainFields(),
		},
		{
			Name:         "messages",
			Description:  "MailerSend messages (API send requests).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       mailersendMessageFields(),
		},
		{
			Name:         "recipients",
			Description:  "MailerSend recipients.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       mailersendRecipientFields(),
		},
	}
}

func mailersendActivityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "email", Type: "object"},
	}
}

func mailersendDomainFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "dkim", Type: "boolean"},
		{Name: "spf", Type: "boolean"},
		{Name: "tracking", Type: "boolean"},
		{Name: "is_verified", Type: "boolean"},
		{Name: "is_dns_active", Type: "boolean"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func mailersendMessageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func mailersendRecipientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "deleted_at", Type: "string"},
	}
}

func mailersendActivityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"type":       item["type"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
		"email":      item["email"],
	}
}

func mailersendDomainRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"dkim":          item["dkim"],
		"spf":           item["spf"],
		"tracking":      item["tracking"],
		"is_verified":   item["is_verified"],
		"is_dns_active": item["is_dns_active"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
	}
}

func mailersendMessageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func mailersendRecipientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"email":      item["email"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
		"deleted_at": item["deleted_at"],
	}
}
