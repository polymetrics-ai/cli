package campaignmonitor

import "polymetrics/internal/connectors"

// streamEndpoint describes how to read one Campaign Monitor stream: the API
// resource path (relative to base_url), whether it is per-client (requiring a
// client_id), whether the payload is paged (a {Results:[...]} envelope) or a
// bare top-level JSON array, and the mapper that flattens each object.
type streamEndpoint struct {
	// resource is the path template. The %s placeholder, when scoped is true, is
	// filled with the configured client_id.
	resource string
	// scoped marks an endpoint nested under /clients/{clientid}/.
	scoped bool
	// paged marks an endpoint that returns the page envelope (Results +
	// PageNumber + NumberOfPages); false means a bare top-level array.
	paged bool
	// mapRecord flattens a raw Campaign Monitor object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in campaignMonitorStreams.
var streamEndpoints = map[string]streamEndpoint{
	"clients":         {resource: "clients.json", scoped: false, paged: false, mapRecord: clientRecord},
	"campaigns":       {resource: "clients/%s/campaigns.json", scoped: true, paged: true, mapRecord: campaignRecord},
	"lists":           {resource: "clients/%s/lists.json", scoped: true, paged: false, mapRecord: listRecord},
	"suppressionlist": {resource: "clients/%s/suppressionlist.json", scoped: true, paged: true, mapRecord: suppressionRecord},
}

// campaignMonitorStreams returns the connector's published stream catalog.
func campaignMonitorStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "clients",
			Description:  "Campaign Monitor clients in the account.",
			PrimaryKey:   []string{"ClientID"},
			CursorFields: nil,
			Fields:       clientFields(),
		},
		{
			Name:         "campaigns",
			Description:  "Sent campaigns for the configured client.",
			PrimaryKey:   []string{"CampaignID"},
			CursorFields: []string{"SentDate"},
			Fields:       campaignFields(),
		},
		{
			Name:         "lists",
			Description:  "Subscriber lists for the configured client.",
			PrimaryKey:   []string{"ListID"},
			CursorFields: nil,
			Fields:       listFields(),
		},
		{
			Name:         "suppressionlist",
			Description:  "Suppressed email addresses for the configured client.",
			PrimaryKey:   []string{"EmailAddress"},
			CursorFields: []string{"Date"},
			Fields:       suppressionFields(),
		},
	}
}

func clientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ClientID", Type: "string"},
		{Name: "Name", Type: "string"},
	}
}

func campaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "CampaignID", Type: "string"},
		{Name: "Name", Type: "string"},
		{Name: "Subject", Type: "string"},
		{Name: "FromName", Type: "string"},
		{Name: "FromEmail", Type: "string"},
		{Name: "ReplyTo", Type: "string"},
		{Name: "WebVersionURL", Type: "string"},
		{Name: "WebVersionTextURL", Type: "string"},
		{Name: "SentDate", Type: "string"},
		{Name: "TotalRecipients", Type: "integer"},
	}
}

func listFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ListID", Type: "string"},
		{Name: "Name", Type: "string"},
	}
}

func suppressionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "EmailAddress", Type: "string"},
		{Name: "Date", Type: "string"},
		{Name: "State", Type: "string"},
		{Name: "SuppressionType", Type: "string"},
	}
}

func clientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ClientID": item["ClientID"],
		"Name":     item["Name"],
	}
}

func campaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"CampaignID":        item["CampaignID"],
		"Name":              item["Name"],
		"Subject":           item["Subject"],
		"FromName":          item["FromName"],
		"FromEmail":         item["FromEmail"],
		"ReplyTo":           item["ReplyTo"],
		"WebVersionURL":     item["WebVersionURL"],
		"WebVersionTextURL": item["WebVersionTextURL"],
		"SentDate":          item["SentDate"],
		"TotalRecipients":   item["TotalRecipients"],
	}
}

func listRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ListID": item["ListID"],
		"Name":   item["Name"],
	}
}

func suppressionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"EmailAddress":    item["EmailAddress"],
		"Date":            item["Date"],
		"State":           item["State"],
		"SuppressionType": item["SuppressionType"],
	}
}
