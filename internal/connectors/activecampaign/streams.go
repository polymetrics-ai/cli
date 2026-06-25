package activecampaign

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the ActiveCampaign API v3 resource path
// (relative to base_url) and the record mapper that flattens its objects, plus
// the JSON key under which the list response wraps the records array.
type streamEndpoint struct {
	// resource is the API v3 list endpoint path segment (e.g. "contacts").
	resource string
	// recordsKey is the top-level JSON key holding the records array. For most
	// ActiveCampaign list endpoints this matches the resource name.
	recordsKey string
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streams(); the read path is fully
// data-driven from this table.
var streamEndpoints = map[string]streamEndpoint{
	"contacts":  {resource: "contacts", recordsKey: "contacts", mapRecord: contactRecord},
	"lists":     {resource: "lists", recordsKey: "lists", mapRecord: listRecord},
	"deals":     {resource: "deals", recordsKey: "deals", mapRecord: dealRecord},
	"campaigns": {resource: "campaigns", recordsKey: "campaigns", mapRecord: campaignRecord},
}

// streams returns the connector's published stream catalog. Every ActiveCampaign
// object exposes a string id; cdate (creation date, RFC3339-ish) is the natural
// cursor field where present.
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "ActiveCampaign contacts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"udate"},
			Fields:       contactFields(),
		},
		{
			Name:         "lists",
			Description:  "ActiveCampaign contact lists.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"cdate"},
			Fields:       listFields(),
		},
		{
			Name:         "deals",
			Description:  "ActiveCampaign deals (CRM).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"mdate"},
			Fields:       dealFields(),
		},
		{
			Name:         "campaigns",
			Description:  "ActiveCampaign email campaigns.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"cdate"},
			Fields:       campaignFields(),
		},
	}
}

func contactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "cdate", Type: "string"},
		{Name: "udate", Type: "string"},
		{Name: "orgid", Type: "string"},
		{Name: "deleted", Type: "string"},
	}
}

func listFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "stringid", Type: "string"},
		{Name: "sender_url", Type: "string"},
		{Name: "subscriber_count", Type: "string"},
		{Name: "cdate", Type: "string"},
		{Name: "userid", Type: "string"},
	}
}

func dealFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "contact", Type: "string"},
		{Name: "value", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "stage", Type: "string"},
		{Name: "owner", Type: "string"},
		{Name: "cdate", Type: "string"},
		{Name: "mdate", Type: "string"},
	}
}

func campaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "send_amt", Type: "string"},
		{Name: "opens", Type: "string"},
		{Name: "uniqueopens", Type: "string"},
		{Name: "linkclicks", Type: "string"},
		{Name: "cdate", Type: "string"},
		{Name: "mdate", Type: "string"},
	}
}

func contactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"email":     item["email"],
		"firstName": item["firstName"],
		"lastName":  item["lastName"],
		"phone":     item["phone"],
		"cdate":     item["cdate"],
		"udate":     item["udate"],
		"orgid":     item["orgid"],
		"deleted":   item["deleted"],
	}
}

func listRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"stringid":         item["stringid"],
		"sender_url":       item["sender_url"],
		"subscriber_count": item["subscriber_count"],
		"cdate":            item["cdate"],
		"userid":           item["userid"],
	}
}

func dealRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"title":    item["title"],
		"contact":  item["contact"],
		"value":    item["value"],
		"currency": item["currency"],
		"status":   item["status"],
		"stage":    item["stage"],
		"owner":    item["owner"],
		"cdate":    item["cdate"],
		"mdate":    item["mdate"],
	}
}

func campaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"type":        item["type"],
		"status":      item["status"],
		"subject":     item["subject"],
		"send_amt":    item["send_amt"],
		"opens":       item["opens"],
		"uniqueopens": item["uniqueopens"],
		"linkclicks":  item["linkclicks"],
		"cdate":       item["cdate"],
		"mdate":       item["mdate"],
	}
}
