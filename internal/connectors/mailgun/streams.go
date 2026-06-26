package mailgun

import "polymetrics.ai/internal/connectors"

// pagination identifies how a Mailgun stream pages its list endpoint.
type pagination int

const (
	// paginationOffset uses skip/limit and a {items,total_count} envelope
	// (used by /v3/domains).
	paginationOffset pagination = iota
	// paginationPagingNext follows the absolute paging.next URL in the body
	// envelope {items,paging:{next}} (used by events and most v3 sub-resources).
	paginationPagingNext
)

// streamEndpoint maps a stream name to the Mailgun API resource path it reads
// from, how it paginates, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the path relative to base_url. When it contains the token
	// "{domain}" it is substituted with the configured domain_name at read time.
	resource string
	// pagination selects the read loop driving this stream.
	pagination pagination
	// needsDomain marks streams whose resource path requires a domain_name.
	needsDomain bool
	// mapRecord flattens a raw Mailgun object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// mailgunStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in mailgunStreams; the read
// path is fully data-driven from this table.
var mailgunStreamEndpoints = map[string]streamEndpoint{
	"domains":       {resource: "v3/domains", pagination: paginationOffset, mapRecord: mailgunDomainRecord},
	"events":        {resource: "v3/{domain}/events", pagination: paginationPagingNext, needsDomain: true, mapRecord: mailgunEventRecord},
	"mailing_lists": {resource: "v3/lists/pages", pagination: paginationPagingNext, mapRecord: mailgunMailingListRecord},
	"tags":          {resource: "v3/{domain}/tags", pagination: paginationPagingNext, needsDomain: true, mapRecord: mailgunTagRecord},
}

// mailgunStreams returns the connector's published stream catalog.
func mailgunStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "domains",
			Description:  "Mailgun sending domains for the account.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       mailgunDomainFields(),
		},
		{
			Name:         "events",
			Description:  "Mailgun email events (accepted, delivered, opened, failed, etc.) for a domain.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"timestamp"},
			Fields:       mailgunEventFields(),
		},
		{
			Name:         "mailing_lists",
			Description:  "Mailgun mailing lists for the account.",
			PrimaryKey:   []string{"address"},
			CursorFields: []string{"created_at"},
			Fields:       mailgunMailingListFields(),
		},
		{
			Name:         "tags",
			Description:  "Mailgun analytics tags for a domain.",
			PrimaryKey:   []string{"tag"},
			CursorFields: nil,
			Fields:       mailgunTagFields(),
		},
	}
}

func mailgunDomainFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "is_disabled", Type: "boolean"},
		{Name: "smtp_login", Type: "string"},
		{Name: "spam_action", Type: "string"},
		{Name: "wildcard", Type: "boolean"},
		{Name: "created_at", Type: "string"},
	}
}

func mailgunEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "event", Type: "string"},
		{Name: "timestamp", Type: "number"},
		{Name: "recipient", Type: "string"},
		{Name: "message_id", Type: "string"},
		{Name: "log_level", Type: "string"},
		{Name: "reason", Type: "string"},
	}
}

func mailgunMailingListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "address", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "access_level", Type: "string"},
		{Name: "members_count", Type: "integer"},
		{Name: "created_at", Type: "string"},
	}
}

func mailgunTagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "tag", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "first_seen", Type: "string"},
		{Name: "last_seen", Type: "string"},
	}
}

func mailgunDomainRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          firstNonEmpty(item, "id", "name"),
		"name":        item["name"],
		"state":       item["state"],
		"type":        item["type"],
		"is_disabled": item["is_disabled"],
		"smtp_login":  item["smtp_login"],
		"spam_action": item["spam_action"],
		"wildcard":    item["wildcard"],
		"created_at":  item["created_at"],
	}
}

func mailgunEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"event":      item["event"],
		"timestamp":  item["timestamp"],
		"recipient":  item["recipient"],
		"message_id": item["message-id"],
		"log_level":  item["log-level"],
		"reason":     item["reason"],
	}
}

func mailgunMailingListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"address":       item["address"],
		"name":          item["name"],
		"description":   item["description"],
		"access_level":  item["access_level"],
		"members_count": item["members_count"],
		"created_at":    item["created_at"],
	}
}

func mailgunTagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"tag":         firstNonEmpty(item, "tag", "name"),
		"description": item["description"],
		"first_seen":  item["first-seen"],
		"last_seen":   item["last-seen"],
	}
}

// firstNonEmpty returns the first non-empty string value among keys.
func firstNonEmpty(item map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := item[k]; ok {
			if s, isStr := v.(string); isStr {
				if s != "" {
					return v
				}
				continue
			}
			if v != nil {
				return v
			}
		}
	}
	return nil
}
