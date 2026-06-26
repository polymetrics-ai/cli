package elasticemail

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Elastic Email v4 API resource path
// (relative to base_url) it reads from, and the record mapper that projects its
// objects into a connectors.Record. The Elastic Email v4 list endpoints return a
// top-level JSON array (record selector field_path is the root), so there is no
// response wrapper to unwrap.
type streamEndpoint struct {
	// resource is the v4 list endpoint path segment (e.g. "contacts").
	resource string
	// mapRecord projects a raw Elastic Email object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// elasticEmailStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in elasticEmailStreams;
// the read path is fully data-driven from this table.
var elasticEmailStreamEndpoints = map[string]streamEndpoint{
	"contacts":  {resource: "contacts", mapRecord: contactRecord},
	"campaigns": {resource: "campaigns", mapRecord: campaignRecord},
	"lists":     {resource: "lists", mapRecord: listRecord},
	"segments":  {resource: "segments", mapRecord: segmentRecord},
	"templates": {resource: "templates", mapRecord: templateRecord},
}

// elasticEmailStreams returns the connector's published stream catalog. Elastic
// Email v4 objects are keyed by natural identifiers (Email, Name, ListName)
// rather than a synthetic id, so primary keys vary per stream.
func elasticEmailStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "Elastic Email contacts (recipients).",
			PrimaryKey:   []string{"Email"},
			CursorFields: []string{"DateUpdated"},
			Fields:       contactFields(),
		},
		{
			Name:        "campaigns",
			Description: "Elastic Email campaigns.",
			PrimaryKey:  []string{"Name"},
			Fields:      campaignFields(),
		},
		{
			Name:        "lists",
			Description: "Elastic Email contact lists.",
			PrimaryKey:  []string{"ListName"},
			Fields:      listFields(),
		},
		{
			Name:        "segments",
			Description: "Elastic Email contact segments.",
			PrimaryKey:  []string{"Name"},
			Fields:      segmentFields(),
		},
		{
			Name:        "templates",
			Description: "Elastic Email templates.",
			PrimaryKey:  []string{"Name"},
			Fields:      templateFields(),
		},
	}
}

func contactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "Email", Type: "string"},
		{Name: "FirstName", Type: "string"},
		{Name: "LastName", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "Source", Type: "string"},
		{Name: "DateAdded", Type: "timestamp"},
		{Name: "DateUpdated", Type: "timestamp"},
		{Name: "StatusChangeDate", Type: "timestamp"},
	}
}

func campaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "Name", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "Content", Type: "object"},
		{Name: "Recipients", Type: "object"},
		{Name: "Options", Type: "object"},
	}
}

func listFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ListName", Type: "string"},
		{Name: "PublicListID", Type: "string"},
		{Name: "DateAdded", Type: "timestamp"},
		{Name: "AllowUnsubscribe", Type: "boolean"},
	}
}

func segmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "Name", Type: "string"},
		{Name: "Rule", Type: "string"},
	}
}

func templateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "Name", Type: "string"},
		{Name: "Subject", Type: "string"},
		{Name: "TemplateScope", Type: "string"},
		{Name: "DateAdded", Type: "timestamp"},
		{Name: "Body", Type: "object"},
	}
}

func contactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"Email":            item["Email"],
		"FirstName":        item["FirstName"],
		"LastName":         item["LastName"],
		"Status":           item["Status"],
		"Source":           item["Source"],
		"DateAdded":        item["DateAdded"],
		"DateUpdated":      item["DateUpdated"],
		"StatusChangeDate": item["StatusChangeDate"],
		"Activity":         item["Activity"],
		"Consent":          item["Consent"],
		"CustomFields":     item["CustomFields"],
	}
}

func campaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"Name":       item["Name"],
		"Status":     item["Status"],
		"Content":    item["Content"],
		"Recipients": item["Recipients"],
		"Options":    item["Options"],
	}
}

func listRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ListName":         item["ListName"],
		"PublicListID":     item["PublicListID"],
		"DateAdded":        item["DateAdded"],
		"AllowUnsubscribe": item["AllowUnsubscribe"],
	}
}

func segmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"Name": item["Name"],
		"Rule": item["Rule"],
	}
}

func templateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"Name":          item["Name"],
		"Subject":       item["Subject"],
		"TemplateScope": item["TemplateScope"],
		"DateAdded":     item["DateAdded"],
		"Body":          item["Body"],
	}
}
