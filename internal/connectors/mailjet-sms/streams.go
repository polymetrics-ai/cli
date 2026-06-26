package mailjetsms

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Mailjet SMS API resource path
// (relative to base_url), the JSON path to its record array, whether it is a
// paginated list, and the mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "sms").
	resource string
	// recordsPath is the dotted JSON path to the records (e.g. "Data"). An empty
	// path means the response is a single object emitted as one record.
	recordsPath string
	// paginated reports whether the endpoint supports Limit/Offset pagination.
	paginated bool
	// mapRecord flattens a raw API object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streams(); the read path is fully
// data-driven from this table.
var streamEndpoints = map[string]streamEndpoint{
	"sms":       {resource: "sms", recordsPath: "Data", paginated: true, mapRecord: smsRecord},
	"sms_count": {resource: "sms/count", recordsPath: "", paginated: false, mapRecord: smsCountRecord},
}

// streams returns the connector's published stream catalog. The Mailjet SMS API
// exposes a single SMS resource plus a count endpoint; both are full-refresh
// only. The SMS object carries a string ID primary key and a CreationTS cursor.
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "sms",
			Description:  "Mailjet outbound SMS messages.",
			PrimaryKey:   []string{"ID"},
			CursorFields: []string{"CreationTS"},
			Fields:       smsFields(),
		},
		{
			Name:        "sms_count",
			Description: "Mailjet SMS message count for the configured date window.",
			PrimaryKey:  nil,
			Fields:      smsCountFields(),
		},
	}
}

func smsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ID", Type: "string"},
		{Name: "From", Type: "string"},
		{Name: "To", Type: "string"},
		{Name: "MessageId", Type: "string"},
		{Name: "CreationTS", Type: "integer"},
		{Name: "SentTS", Type: "integer"},
		{Name: "SMSCount", Type: "integer"},
		{Name: "status_code", Type: "integer"},
		{Name: "status_name", Type: "string"},
		{Name: "status_description", Type: "string"},
		{Name: "cost_value", Type: "number"},
		{Name: "cost_currency", Type: "string"},
	}
}

func smsCountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "Count", Type: "integer"},
	}
}

// smsRecord flattens a Mailjet SMS message. The API nests Status and Cost as
// sub-objects; we surface their scalar members as flat columns so downstream
// warehouses get a tabular shape, while preserving the top-level fields.
func smsRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"ID":         item["ID"],
		"From":       item["From"],
		"To":         item["To"],
		"MessageId":  item["MessageId"],
		"CreationTS": item["CreationTS"],
		"SentTS":     item["SentTS"],
		"SMSCount":   item["SMSCount"],
	}
	if status, ok := item["Status"].(map[string]any); ok {
		rec["status_code"] = status["Code"]
		rec["status_name"] = status["Name"]
		rec["status_description"] = status["Description"]
	}
	if cost, ok := item["Cost"].(map[string]any); ok {
		rec["cost_value"] = cost["Value"]
		rec["cost_currency"] = cost["Currency"]
	}
	return rec
}

func smsCountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"Count": item["Count"],
	}
}
