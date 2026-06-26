package dwolla

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Dwolla API resource path (relative to
// base_url) it reads from, the HAL _embedded key its records live under, and the
// record mapper that flattens its objects.
//
// Dwolla is a HAL+JSON API: list responses nest their records under
// _embedded.<embedKey> and advertise the next page as an absolute URL at
// _links.next.href. The embed key is not always the stream name (e.g. the
// exchange_partners stream embeds under "exchange-partners").
type streamEndpoint struct {
	// resource is the Dwolla list endpoint path segment (e.g. "customers").
	resource string
	// embedKey is the key under _embedded holding the records array.
	embedKey string
	// mapRecord flattens a raw Dwolla object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// dwollaStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in dwollaStreams; the read path
// is fully data-driven from this table. The core set is the top-level list
// endpoints that read cleanly without a parent partition (Dwolla's
// funding-sources require a per-customer path, so they are intentionally out of
// this first cut).
var dwollaStreamEndpoints = map[string]streamEndpoint{
	"customers":                {resource: "customers", embedKey: "customers", mapRecord: dwollaCustomerRecord},
	"events":                   {resource: "events", embedKey: "events", mapRecord: dwollaEventRecord},
	"exchange_partners":        {resource: "exchange-partners", embedKey: "exchange-partners", mapRecord: dwollaExchangePartnerRecord},
	"business_classifications": {resource: "business-classifications", embedKey: "business-classifications", mapRecord: dwollaBusinessClassificationRecord},
}

// dwollaStreams returns the connector's published stream catalog. Every Dwolla
// resource exposes a string id and an RFC3339 `created` timestamp, so the
// primary key is ["id"] and the incremental cursor field is ["created"].
func dwollaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "customers",
			Description:  "Dwolla customers (account holders).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       dwollaCustomerFields(),
		},
		{
			Name:         "events",
			Description:  "Dwolla account activity events.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       dwollaEventFields(),
		},
		{
			Name:         "exchange_partners",
			Description:  "Dwolla exchange partners.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created"},
			Fields:       dwollaExchangePartnerFields(),
		},
		{
			Name:        "business_classifications",
			Description: "Dwolla business classifications reference data.",
			PrimaryKey:  []string{"id"},
			Fields:      dwollaBusinessClassificationFields(),
		},
	}
}

func dwollaCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "businessName", Type: "string"},
		{Name: "created", Type: "string"},
	}
}

func dwollaEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "topic", Type: "string"},
		{Name: "resourceId", Type: "string"},
		{Name: "created", Type: "string"},
	}
}

func dwollaExchangePartnerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created", Type: "string"},
	}
}

func dwollaBusinessClassificationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
	}
}

func dwollaCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"firstName":    item["firstName"],
		"lastName":     item["lastName"],
		"email":        item["email"],
		"type":         item["type"],
		"status":       item["status"],
		"businessName": item["businessName"],
		"created":      item["created"],
	}
}

func dwollaEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"topic":      item["topic"],
		"resourceId": item["resourceId"],
		"created":    item["created"],
	}
}

func dwollaExchangePartnerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      item["id"],
		"name":    item["name"],
		"status":  item["status"],
		"created": item["created"],
	}
}

func dwollaBusinessClassificationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   item["id"],
		"name": item["name"],
	}
}
