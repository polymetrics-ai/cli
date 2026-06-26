package highlevel

import "polymetrics.ai/internal/connectors"

// pageStyle describes how a HighLevel stream is paginated.
type pageStyle int

const (
	// styleNone makes a single request and emits whatever the response holds.
	styleNone pageStyle = iota
	// styleCursor follows the absolute meta.nextPageUrl returned in each page.
	styleCursor
)

// streamEndpoint maps a stream name to its HighLevel proxy resource path
// (relative to base_url), the JSON selector for the records array, the
// pagination style, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the proxy list endpoint path segment (e.g. "contacts").
	resource string
	// recordsPath is the dotted JSON path to the records array in the response
	// body (e.g. "contacts", "customFields").
	recordsPath string
	// style controls pagination.
	style pageStyle
	// mapRecord flattens a raw HighLevel object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in highLevelStreams; the read path is
// fully data-driven from this table. Paths mirror the Airbyte source-high-level
// proxy connector (/airbyte/<resource>).
var streamEndpoints = map[string]streamEndpoint{
	"contacts":         {resource: "contacts", recordsPath: "contacts", style: styleCursor, mapRecord: contactRecord},
	"opportunities":    {resource: "opportunities", recordsPath: "opportunities", style: styleCursor, mapRecord: opportunityRecord},
	"pipelines":        {resource: "pipelines", recordsPath: "pipelines", style: styleNone, mapRecord: pipelineRecord},
	"custom_fields":    {resource: "customfields", recordsPath: "customFields", style: styleNone, mapRecord: customFieldRecord},
	"form_submissions": {resource: "form-submissions", recordsPath: "submissions", style: styleCursor, mapRecord: formSubmissionRecord},
}

// highLevelStreams returns the connector's published stream catalog.
func highLevelStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "HighLevel contacts for the configured location.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"dateUpdated"},
			Fields:       contactFields(),
		},
		{
			Name:         "opportunities",
			Description:  "HighLevel opportunities for the configured location.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"dateUpdated"},
			Fields:       opportunityFields(),
		},
		{
			Name:        "pipelines",
			Description: "HighLevel opportunity pipelines for the configured location.",
			PrimaryKey:  []string{"id"},
			Fields:      pipelineFields(),
		},
		{
			Name:        "custom_fields",
			Description: "HighLevel custom fields defined for the configured location.",
			PrimaryKey:  []string{"id"},
			Fields:      customFieldFields(),
		},
		{
			Name:         "form_submissions",
			Description:  "HighLevel form submissions for the configured location.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       formSubmissionFields(),
		},
	}
}

func contactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "locationId", Type: "string"},
		{Name: "contactName", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "dateAdded", Type: "string"},
		{Name: "dateUpdated", Type: "string"},
	}
}

func opportunityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "pipelineId", Type: "string"},
		{Name: "pipelineStageId", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "monetaryValue", Type: "number"},
		{Name: "contactId", Type: "string"},
		{Name: "assignedTo", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "dateAdded", Type: "string"},
		{Name: "dateUpdated", Type: "string"},
	}
}

func pipelineFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "locationId", Type: "string"},
		{Name: "stages", Type: "object"},
		{Name: "dateAdded", Type: "string"},
		{Name: "dateUpdated", Type: "string"},
	}
}

func customFieldFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "fieldKey", Type: "string"},
		{Name: "dataType", Type: "string"},
		{Name: "model", Type: "string"},
		{Name: "position", Type: "integer"},
	}
}

func formSubmissionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "formId", Type: "string"},
		{Name: "contactId", Type: "string"},
		{Name: "locationId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "createdAt", Type: "string"},
	}
}

func contactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"locationId":  item["locationId"],
		"contactName": item["contactName"],
		"firstName":   item["firstName"],
		"lastName":    item["lastName"],
		"email":       item["email"],
		"phone":       item["phone"],
		"type":        item["type"],
		"source":      item["source"],
		"dateAdded":   item["dateAdded"],
		"dateUpdated": item["dateUpdated"],
	}
}

func opportunityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            item["name"],
		"pipelineId":      item["pipelineId"],
		"pipelineStageId": item["pipelineStageId"],
		"status":          item["status"],
		"monetaryValue":   item["monetaryValue"],
		"contactId":       item["contactId"],
		"assignedTo":      item["assignedTo"],
		"source":          item["source"],
		"dateAdded":       item["dateAdded"],
		"dateUpdated":     item["dateUpdated"],
	}
}

func pipelineRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"locationId":  item["locationId"],
		"stages":      item["stages"],
		"dateAdded":   item["dateAdded"],
		"dateUpdated": item["dateUpdated"],
	}
}

func customFieldRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"name":     item["name"],
		"fieldKey": item["fieldKey"],
		"dataType": item["dataType"],
		"model":    item["model"],
		"position": item["position"],
	}
}

func formSubmissionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"formId":     item["formId"],
		"contactId":  item["contactId"],
		"locationId": item["locationId"],
		"name":       item["name"],
		"email":      item["email"],
		"createdAt":  item["createdAt"],
	}
}
