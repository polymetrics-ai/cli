package mailchimp

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Mailchimp Marketing API resource path
// (relative to base_url), the JSON key under which the records array lives in
// the response body, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Mailchimp list endpoint path segment (e.g. "lists").
	resource string
	// recordsKey is the top-level JSON array key in the response (e.g. "lists").
	recordsKey string
	// mapRecord flattens a raw Mailchimp object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// mailchimpStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in mailchimpStreams; the
// read path is fully data-driven from this table.
var mailchimpStreamEndpoints = map[string]streamEndpoint{
	"lists":       {resource: "lists", recordsKey: "lists", mapRecord: mailchimpListRecord},
	"campaigns":   {resource: "campaigns", recordsKey: "campaigns", mapRecord: mailchimpCampaignRecord},
	"reports":     {resource: "reports", recordsKey: "reports", mapRecord: mailchimpReportRecord},
	"automations": {resource: "automations", recordsKey: "automations", mapRecord: mailchimpAutomationRecord},
}

// mailchimpStreams returns the connector's published stream catalog. Mailchimp
// objects carry a string id and (for lists/campaigns) a creation timestamp that
// serves as the incremental cursor.
func mailchimpStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "lists",
			Description:  "Mailchimp audiences (lists).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_created"},
			Fields:       mailchimpListFields(),
		},
		{
			Name:         "campaigns",
			Description:  "Mailchimp campaigns.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"create_time"},
			Fields:       mailchimpCampaignFields(),
		},
		{
			Name:         "reports",
			Description:  "Mailchimp campaign reports (aggregate stats).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"send_time"},
			Fields:       mailchimpReportFields(),
		},
		{
			Name:         "automations",
			Description:  "Mailchimp classic automations (workflows).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"create_time"},
			Fields:       mailchimpAutomationFields(),
		},
	}
}

func mailchimpListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "web_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "date_created", Type: "string"},
		{Name: "list_rating", Type: "integer"},
		{Name: "email_type_option", Type: "boolean"},
		{Name: "visibility", Type: "string"},
		{Name: "subscribe_url_short", Type: "string"},
	}
}

func mailchimpCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "web_id", Type: "integer"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "create_time", Type: "string"},
		{Name: "send_time", Type: "string"},
		{Name: "emails_sent", Type: "integer"},
		{Name: "archive_url", Type: "string"},
	}
}

func mailchimpReportFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "campaign_title", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "list_id", Type: "string"},
		{Name: "emails_sent", Type: "integer"},
		{Name: "abuse_reports", Type: "integer"},
		{Name: "unsubscribed", Type: "integer"},
		{Name: "send_time", Type: "string"},
	}
}

func mailchimpAutomationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "create_time", Type: "string"},
		{Name: "start_time", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "emails_sent", Type: "integer"},
	}
}

func mailchimpListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"web_id":              item["web_id"],
		"name":                item["name"],
		"date_created":        item["date_created"],
		"list_rating":         item["list_rating"],
		"email_type_option":   item["email_type_option"],
		"visibility":          item["visibility"],
		"subscribe_url_short": item["subscribe_url_short"],
	}
}

func mailchimpCampaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"web_id":      item["web_id"],
		"type":        item["type"],
		"status":      item["status"],
		"create_time": item["create_time"],
		"send_time":   item["send_time"],
		"emails_sent": item["emails_sent"],
		"archive_url": item["archive_url"],
	}
}

func mailchimpReportRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"campaign_title": item["campaign_title"],
		"type":           item["type"],
		"list_id":        item["list_id"],
		"emails_sent":    item["emails_sent"],
		"abuse_reports":  item["abuse_reports"],
		"unsubscribed":   item["unsubscribed"],
		"send_time":      item["send_time"],
	}
}

func mailchimpAutomationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"create_time": item["create_time"],
		"start_time":  item["start_time"],
		"status":      item["status"],
		"emails_sent": item["emails_sent"],
	}
}
