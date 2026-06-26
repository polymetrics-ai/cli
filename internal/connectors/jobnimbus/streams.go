package jobnimbus

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the JobNimbus API resource path (relative
// to base_url), the JSON path the records array lives under (which differs per
// stream: results / activity / files), and the record mapper that flattens
// objects into connectors.Record values.
type streamEndpoint struct {
	// resource is the path segment under /api1 (e.g. "contacts").
	resource string
	// recordsPath is the dotted JSON path to the records array in the list
	// response. JobNimbus is inconsistent: most streams use "results", but
	// activities nest under "activity" and files under "files".
	recordsPath string
	// mapRecord flattens a raw JobNimbus object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// jobnimbusStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in jobnimbusStreams; the
// read path is fully data-driven from this table.
var jobnimbusStreamEndpoints = map[string]streamEndpoint{
	"contacts":   {resource: "contacts", recordsPath: "results", mapRecord: jobnimbusContactRecord},
	"jobs":       {resource: "jobs", recordsPath: "results", mapRecord: jobnimbusJobRecord},
	"tasks":      {resource: "tasks", recordsPath: "results", mapRecord: jobnimbusTaskRecord},
	"activities": {resource: "activities", recordsPath: "activity", mapRecord: jobnimbusActivityRecord},
	"files":      {resource: "files", recordsPath: "files", mapRecord: jobnimbusFileRecord},
}

// jobnimbusStreams returns the connector's published stream catalog. Every
// JobNimbus object exposes a string id `jnid` (the primary key) and a numeric
// `date_updated` epoch timestamp used as the incremental cursor field.
func jobnimbusStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "JobNimbus contacts (people and companies in the CRM).",
			PrimaryKey:   []string{"jnid"},
			CursorFields: []string{"date_updated"},
			Fields:       jobnimbusContactFields(),
		},
		{
			Name:         "jobs",
			Description:  "JobNimbus jobs (projects and work records).",
			PrimaryKey:   []string{"jnid"},
			CursorFields: []string{"date_updated"},
			Fields:       jobnimbusJobFields(),
		},
		{
			Name:         "tasks",
			Description:  "JobNimbus tasks and scheduled activities.",
			PrimaryKey:   []string{"jnid"},
			CursorFields: []string{"date_updated"},
			Fields:       jobnimbusTaskFields(),
		},
		{
			Name:         "activities",
			Description:  "JobNimbus activity log entries (notes, status changes).",
			PrimaryKey:   []string{"jnid"},
			CursorFields: []string{"date_updated"},
			Fields:       jobnimbusActivityFields(),
		},
		{
			Name:         "files",
			Description:  "JobNimbus file attachments metadata.",
			PrimaryKey:   []string{"jnid"},
			CursorFields: []string{"date_updated"},
			Fields:       jobnimbusFileFields(),
		},
	}
}

func jobnimbusContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "jnid", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "mobile_phone", Type: "string"},
		{Name: "home_phone", Type: "string"},
		{Name: "work_phone", Type: "string"},
		{Name: "address_line1", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "state_text", Type: "string"},
		{Name: "zip", Type: "string"},
		{Name: "country_name", Type: "string"},
		{Name: "status_name", Type: "string"},
		{Name: "record_type_name", Type: "string"},
		{Name: "sales_rep_name", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "is_archived", Type: "boolean"},
		{Name: "date_created", Type: "integer"},
		{Name: "date_updated", Type: "integer"},
	}
}

func jobnimbusJobFields() []connectors.Field {
	return []connectors.Field{
		{Name: "jnid", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "status_name", Type: "string"},
		{Name: "record_type_name", Type: "string"},
		{Name: "customer", Type: "string"},
		{Name: "address_line1", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "state_text", Type: "string"},
		{Name: "zip", Type: "string"},
		{Name: "sales_rep_name", Type: "string"},
		{Name: "approved_estimate_total", Type: "number"},
		{Name: "approved_invoice_total", Type: "number"},
		{Name: "is_active", Type: "boolean"},
		{Name: "is_archived", Type: "boolean"},
		{Name: "date_created", Type: "integer"},
		{Name: "date_updated", Type: "integer"},
		{Name: "date_status_change", Type: "integer"},
	}
}

func jobnimbusTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "jnid", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "record_type_name", Type: "string"},
		{Name: "priority", Type: "integer"},
		{Name: "customer", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "is_archived", Type: "boolean"},
		{Name: "is_completed", Type: "boolean"},
		{Name: "date_start", Type: "integer"},
		{Name: "date_end", Type: "integer"},
		{Name: "date_created", Type: "integer"},
		{Name: "date_updated", Type: "integer"},
	}
}

func jobnimbusActivityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "jnid", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "note", Type: "string"},
		{Name: "record_type_name", Type: "string"},
		{Name: "customer", Type: "string"},
		{Name: "created_by_name", Type: "string"},
		{Name: "sales_rep_name", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "is_archived", Type: "boolean"},
		{Name: "is_status_change", Type: "boolean"},
		{Name: "date_created", Type: "integer"},
		{Name: "date_updated", Type: "integer"},
	}
}

func jobnimbusFileFields() []connectors.Field {
	return []connectors.Field{
		{Name: "jnid", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "filename", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "content_type", Type: "string"},
		{Name: "record_type_name", Type: "string"},
		{Name: "customer", Type: "string"},
		{Name: "created_by_name", Type: "string"},
		{Name: "size", Type: "number"},
		{Name: "md5", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "date_file_created", Type: "integer"},
		{Name: "date_created", Type: "integer"},
		{Name: "date_updated", Type: "integer"},
	}
}

func jobnimbusContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"jnid":             item["jnid"],
		"type":             item["type"],
		"display_name":     item["display_name"],
		"first_name":       item["first_name"],
		"last_name":        item["last_name"],
		"company":          item["company"],
		"email":            item["email"],
		"mobile_phone":     item["mobile_phone"],
		"home_phone":       item["home_phone"],
		"work_phone":       item["work_phone"],
		"address_line1":    item["address_line1"],
		"city":             item["city"],
		"state_text":       item["state_text"],
		"zip":              item["zip"],
		"country_name":     item["country_name"],
		"status_name":      item["status_name"],
		"record_type_name": item["record_type_name"],
		"sales_rep_name":   item["sales_rep_name"],
		"is_active":        item["is_active"],
		"is_archived":      item["is_archived"],
		"date_created":     item["date_created"],
		"date_updated":     item["date_updated"],
	}
}

func jobnimbusJobRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"jnid":                    item["jnid"],
		"type":                    item["type"],
		"name":                    item["name"],
		"number":                  item["number"],
		"status_name":             item["status_name"],
		"record_type_name":        item["record_type_name"],
		"customer":                item["customer"],
		"address_line1":           item["address_line1"],
		"city":                    item["city"],
		"state_text":              item["state_text"],
		"zip":                     item["zip"],
		"sales_rep_name":          item["sales_rep_name"],
		"approved_estimate_total": item["approved_estimate_total"],
		"approved_invoice_total":  item["approved_invoice_total"],
		"is_active":               item["is_active"],
		"is_archived":             item["is_archived"],
		"date_created":            item["date_created"],
		"date_updated":            item["date_updated"],
		"date_status_change":      item["date_status_change"],
	}
}

func jobnimbusTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"jnid":             item["jnid"],
		"type":             item["type"],
		"title":            item["title"],
		"description":      item["description"],
		"number":           item["number"],
		"record_type_name": item["record_type_name"],
		"priority":         item["priority"],
		"customer":         item["customer"],
		"is_active":        item["is_active"],
		"is_archived":      item["is_archived"],
		"is_completed":     item["is_completed"],
		"date_start":       item["date_start"],
		"date_end":         item["date_end"],
		"date_created":     item["date_created"],
		"date_updated":     item["date_updated"],
	}
}

func jobnimbusActivityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"jnid":             item["jnid"],
		"type":             item["type"],
		"note":             item["note"],
		"record_type_name": item["record_type_name"],
		"customer":         item["customer"],
		"created_by_name":  item["created_by_name"],
		"sales_rep_name":   item["sales_rep_name"],
		"source":           item["source"],
		"is_active":        item["is_active"],
		"is_archived":      item["is_archived"],
		"is_status_change": item["is_status_change"],
		"date_created":     item["date_created"],
		"date_updated":     item["date_updated"],
	}
}

func jobnimbusFileRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"jnid":              item["jnid"],
		"type":              item["type"],
		"filename":          item["filename"],
		"description":       item["description"],
		"content_type":      item["content_type"],
		"record_type_name":  item["record_type_name"],
		"customer":          item["customer"],
		"created_by_name":   item["created_by_name"],
		"size":              item["size"],
		"md5":               item["md5"],
		"is_active":         item["is_active"],
		"date_file_created": item["date_file_created"],
		"date_created":      item["date_created"],
		"date_updated":      item["date_updated"],
	}
}
