package campayn

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Campayn API resource path (relative to
// base_url) it reads from, and the record mapper that shapes its objects. Campayn
// returns each collection as a bare top-level JSON array (record_selector
// field_path is empty), so the records path is the root "".
//
// forms and contacts are substreams of lists: their path is templated per parent
// list id ("/lists/{id}/...") and they are read by fanning out over every list.
type streamEndpoint struct {
	// resource is the path for top-level streams (e.g. "lists.json"). Empty for
	// substreams, which build their path from listResource.
	resource string
	// listResource is the per-list path suffix for substreams (e.g.
	// "contacts.json" -> "/lists/<id>/contacts.json"). Empty for top-level streams.
	listResource string
	// mapRecord shapes a raw Campayn object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// isSubstream reports whether the stream is read per parent list partition.
func (e streamEndpoint) isSubstream() bool { return e.listResource != "" }

// campaignStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in campaignStreams.
var campaignStreamEndpoints = map[string]streamEndpoint{
	"lists":    {resource: "lists.json", mapRecord: listRecord},
	"emails":   {resource: "emails.json", mapRecord: emailRecord},
	"reports":  {resource: "reports/calendar.json", mapRecord: reportRecord},
	"forms":    {listResource: "forms.json", mapRecord: formRecord},
	"contacts": {listResource: "contacts.json", mapRecord: contactRecord},
}

// campaignStreams returns the connector's published stream catalog. Every Campayn
// object exposes a string id as primary key; the API only supports full refresh,
// so there are no incremental cursor fields.
func campaignStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "lists",
			Description: "Campayn subscriber lists.",
			PrimaryKey:  []string{"id"},
			Fields:      listFields(),
		},
		{
			Name:        "forms",
			Description: "Campayn signup forms, read per subscriber list.",
			PrimaryKey:  []string{"id"},
			Fields:      formFields(),
		},
		{
			Name:        "contacts",
			Description: "Campayn contacts, read per subscriber list.",
			PrimaryKey:  []string{"id"},
			Fields:      contactFields(),
		},
		{
			Name:        "emails",
			Description: "Campayn email campaigns.",
			PrimaryKey:  []string{"id"},
			Fields:      emailFields(),
		},
		{
			Name:        "reports",
			Description: "Campayn campaign calendar reports.",
			PrimaryKey:  []string{"id"},
			Fields:      reportFields(),
		},
	}
}

func listFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "list_name", Type: "string"},
		{Name: "contact_count", Type: "number"},
		{Name: "tags", Type: "string"},
	}
}

func formFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "list_id", Type: "string"},
		{Name: "contact_list_id", Type: "string"},
		{Name: "form_title", Type: "string"},
		{Name: "form_type", Type: "string"},
		{Name: "form_html", Type: "string"},
		{Name: "signup_count", Type: "string"},
	}
}

func contactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "list_id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "confirmed", Type: "string"},
		{Name: "image_url", Type: "string"},
	}
}

func emailFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "send_count", Type: "string"},
		{Name: "send_now", Type: "boolean"},
		{Name: "unique_views", Type: "number"},
		{Name: "percent_views", Type: "number"},
		{Name: "unique_responses", Type: "number"},
		{Name: "percent_responses", Type: "number"},
		{Name: "preview_url", Type: "string"},
		{Name: "preview_thumb", Type: "string"},
	}
}

func reportFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "scheduled_date", Type: "string"},
		{Name: "preview_url", Type: "string"},
	}
}

func listRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"list_name":     item["list_name"],
		"contact_count": item["contact_count"],
		"tags":          item["tags"],
	}
}

func formRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"contact_list_id": item["contact_list_id"],
		"form_title":      item["form_title"],
		"form_type":       item["form_type"],
		"form_html":       item["form_html"],
		"signup_count":    item["signup_count"],
	}
}

func contactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"email":      item["email"],
		"first_name": item["first_name"],
		"last_name":  item["last_name"],
		"confirmed":  item["confirmed"],
		"image_url":  item["image_url"],
	}
}

func emailRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"status":            item["status"],
		"send_count":        item["send_count"],
		"send_now":          item["send_now"],
		"unique_views":      item["unique_views"],
		"percent_views":     item["percent_views"],
		"unique_responses":  item["unique_responses"],
		"percent_responses": item["percent_responses"],
		"preview_url":       item["preview_url"],
		"preview_thumb":     item["preview_thumb"],
	}
}

func reportRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"status":         item["status"],
		"scheduled_date": item["scheduled_date"],
		"preview_url":    item["preview_url"],
	}
}
