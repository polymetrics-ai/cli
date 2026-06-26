package mailjetmail

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Mailjet REST resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Mailjet REST list endpoint path segment (e.g. "contact").
	resource string
	// mapRecord flattens a raw Mailjet object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// mailjetStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in mailjetStreams; the read
// path is fully data-driven from this table. The published stream names mirror
// the Airbyte source-mailjet-mail streams (pluralized) while the resource is the
// underlying Mailjet v3 REST resource.
var mailjetStreamEndpoints = map[string]streamEndpoint{
	"contacts":      {resource: "contact", mapRecord: mailjetContactRecord},
	"contactslists": {resource: "contactslist", mapRecord: mailjetContactsListRecord},
	"messages":      {resource: "message", mapRecord: mailjetMessageRecord},
	"campaigns":     {resource: "campaign", mapRecord: mailjetCampaignRecord},
	"stats":         {resource: "statcounters", mapRecord: mailjetStatRecord},
}

// mailjetStreams returns the connector's published stream catalog. Mailjet REST
// objects expose a numeric ID; the primary key is ["ID"] across the board. The
// API supports only full-refresh syncs, so no cursor fields are published.
func mailjetStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "contacts",
			Description: "Mailjet contacts.",
			PrimaryKey:  []string{"ID"},
			Fields:      mailjetContactFields(),
		},
		{
			Name:        "contactslists",
			Description: "Mailjet contact lists.",
			PrimaryKey:  []string{"ID"},
			Fields:      mailjetContactsListFields(),
		},
		{
			Name:        "messages",
			Description: "Mailjet messages (sent email events).",
			PrimaryKey:  []string{"ID"},
			Fields:      mailjetMessageFields(),
		},
		{
			Name:        "campaigns",
			Description: "Mailjet campaigns.",
			PrimaryKey:  []string{"ID"},
			Fields:      mailjetCampaignFields(),
		},
		{
			Name:        "stats",
			Description: "Mailjet aggregated statistics counters.",
			PrimaryKey:  []string{"ID"},
			Fields:      mailjetStatFields(),
		},
	}
}

func mailjetContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ID", Type: "integer"},
		{Name: "Email", Type: "string"},
		{Name: "Name", Type: "string"},
		{Name: "CreatedAt", Type: "timestamp"},
		{Name: "DeliveredCount", Type: "integer"},
		{Name: "IsExcludedFromCampaigns", Type: "boolean"},
		{Name: "IsOptInPending", Type: "boolean"},
		{Name: "IsSpamComplaining", Type: "boolean"},
		{Name: "LastActivityAt", Type: "timestamp"},
		{Name: "LastUpdateAt", Type: "timestamp"},
	}
}

func mailjetContactsListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ID", Type: "integer"},
		{Name: "Name", Type: "string"},
		{Name: "Address", Type: "string"},
		{Name: "CreatedAt", Type: "timestamp"},
		{Name: "IsDeleted", Type: "boolean"},
		{Name: "SubscriberCount", Type: "integer"},
	}
}

func mailjetMessageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ID", Type: "integer"},
		{Name: "ContactID", Type: "integer"},
		{Name: "CampaignID", Type: "integer"},
		{Name: "Status", Type: "string"},
		{Name: "ArrivedAt", Type: "timestamp"},
		{Name: "AttemptCount", Type: "integer"},
		{Name: "IsClickTracked", Type: "boolean"},
		{Name: "IsOpenTracked", Type: "boolean"},
		{Name: "MessageSize", Type: "integer"},
	}
}

func mailjetCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ID", Type: "integer"},
		{Name: "Subject", Type: "string"},
		{Name: "FromEmail", Type: "string"},
		{Name: "FromName", Type: "string"},
		{Name: "CreatedAt", Type: "timestamp"},
		{Name: "SendStartAt", Type: "timestamp"},
		{Name: "Status", Type: "integer"},
		{Name: "IsDeleted", Type: "boolean"},
		{Name: "IsStarred", Type: "boolean"},
	}
}

func mailjetStatFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ID", Type: "integer"},
		{Name: "MessageSentCount", Type: "integer"},
		{Name: "MessageOpenedCount", Type: "integer"},
		{Name: "MessageClickedCount", Type: "integer"},
		{Name: "MessageDeliveredCount", Type: "integer"},
		{Name: "MessageBouncedCount", Type: "integer"},
		{Name: "MessageSpamCount", Type: "integer"},
		{Name: "MessageUnsubscribedCount", Type: "integer"},
	}
}

func mailjetContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ID":                      item["ID"],
		"Email":                   item["Email"],
		"Name":                    item["Name"],
		"CreatedAt":               item["CreatedAt"],
		"DeliveredCount":          item["DeliveredCount"],
		"IsExcludedFromCampaigns": item["IsExcludedFromCampaigns"],
		"IsOptInPending":          item["IsOptInPending"],
		"IsSpamComplaining":       item["IsSpamComplaining"],
		"LastActivityAt":          item["LastActivityAt"],
		"LastUpdateAt":            item["LastUpdateAt"],
	}
}

func mailjetContactsListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ID":              item["ID"],
		"Name":            item["Name"],
		"Address":         item["Address"],
		"CreatedAt":       item["CreatedAt"],
		"IsDeleted":       item["IsDeleted"],
		"SubscriberCount": item["SubscriberCount"],
	}
}

func mailjetMessageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ID":             item["ID"],
		"ContactID":      item["ContactID"],
		"CampaignID":     item["CampaignID"],
		"Status":         item["Status"],
		"ArrivedAt":      item["ArrivedAt"],
		"AttemptCount":   item["AttemptCount"],
		"IsClickTracked": item["IsClickTracked"],
		"IsOpenTracked":  item["IsOpenTracked"],
		"MessageSize":    item["MessageSize"],
	}
}

func mailjetCampaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ID":          item["ID"],
		"Subject":     item["Subject"],
		"FromEmail":   item["FromEmail"],
		"FromName":    item["FromName"],
		"CreatedAt":   item["CreatedAt"],
		"SendStartAt": item["SendStartAt"],
		"Status":      item["Status"],
		"IsDeleted":   item["IsDeleted"],
		"IsStarred":   item["IsStarred"],
	}
}

func mailjetStatRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ID":                       item["ID"],
		"MessageSentCount":         item["MessageSentCount"],
		"MessageOpenedCount":       item["MessageOpenedCount"],
		"MessageClickedCount":      item["MessageClickedCount"],
		"MessageDeliveredCount":    item["MessageDeliveredCount"],
		"MessageBouncedCount":      item["MessageBouncedCount"],
		"MessageSpamCount":         item["MessageSpamCount"],
		"MessageUnsubscribedCount": item["MessageUnsubscribedCount"],
	}
}
