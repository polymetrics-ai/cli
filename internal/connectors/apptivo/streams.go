package apptivo

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Apptivo DAO resource path (relative
// to base_url) it reads from, the JSON path holding the records array, and the
// record mapper that projects each object. Adding a stream means adding one
// entry here plus a Stream definition in apptivoStreams; the read path is fully
// data-driven from this table.
type streamEndpoint struct {
	// resource is the DAO path segment (e.g. "customers"), appended to the
	// shared /app/dao/v6/ prefix.
	resource string
	// recordsPath is the dotted path to the records array (Apptivo nests its
	// list results under "data").
	recordsPath string
	// primaryKey is the field that uniquely identifies a record; also used by
	// fixture mode to synthesize deterministic ids.
	primaryKey string
	// mapRecord projects a raw Apptivo object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// apptivoStreamEndpoints is the per-stream routing table.
var apptivoStreamEndpoints = map[string]streamEndpoint{
	"customers":     {resource: "customers", recordsPath: "data", primaryKey: "customerId", mapRecord: apptivoCustomerRecord},
	"contacts":      {resource: "contacts", recordsPath: "data", primaryKey: "contactId", mapRecord: apptivoContactRecord},
	"leads":         {resource: "leads", recordsPath: "data", primaryKey: "id", mapRecord: apptivoLeadRecord},
	"opportunities": {resource: "opportunities", recordsPath: "data", primaryKey: "opportunityId", mapRecord: apptivoOpportunityRecord},
}

// apptivoStreams returns the connector's published stream catalog. Apptivo's
// CRM is full-refresh only (no incremental cursor), so CursorFields are empty.
func apptivoStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "customers",
			Description: "Apptivo customers (accounts).",
			PrimaryKey:  []string{"customerId"},
			Fields:      apptivoCustomerFields(),
		},
		{
			Name:        "contacts",
			Description: "Apptivo contacts.",
			PrimaryKey:  []string{"contactId"},
			Fields:      apptivoContactFields(),
		},
		{
			Name:        "leads",
			Description: "Apptivo leads.",
			PrimaryKey:  []string{"id"},
			Fields:      apptivoLeadFields(),
		},
		{
			Name:        "opportunities",
			Description: "Apptivo opportunities.",
			PrimaryKey:  []string{"opportunityId"},
			Fields:      apptivoOpportunityFields(),
		},
	}
}

func apptivoCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "customerId", Type: "string"},
		{Name: "customerName", Type: "string"},
		{Name: "customerNumber", Type: "string"},
		{Name: "emailAddress", Type: "string"},
		{Name: "phoneNumber", Type: "string"},
		{Name: "website", Type: "string"},
		{Name: "currencyCode", Type: "string"},
		{Name: "statusName", Type: "string"},
		{Name: "creationDate", Type: "string"},
		{Name: "lastUpdateDate", Type: "string"},
	}
}

func apptivoContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "contactId", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "fullName", Type: "string"},
		{Name: "emailAddress", Type: "string"},
		{Name: "phoneNumber", Type: "string"},
		{Name: "companyName", Type: "string"},
		{Name: "creationDate", Type: "string"},
		{Name: "lastUpdateDate", Type: "string"},
	}
}

func apptivoLeadFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "leadId", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "companyName", Type: "string"},
		{Name: "emailAddress", Type: "string"},
		{Name: "phoneNumber", Type: "string"},
		{Name: "leadSource", Type: "string"},
		{Name: "statusName", Type: "string"},
		{Name: "creationDate", Type: "string"},
	}
}

func apptivoOpportunityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "opportunityId", Type: "string"},
		{Name: "opportunityName", Type: "string"},
		{Name: "customerName", Type: "string"},
		{Name: "salesStageName", Type: "string"},
		{Name: "opportunityAmount", Type: "string"},
		{Name: "currencyCode", Type: "string"},
		{Name: "closingDate", Type: "string"},
		{Name: "creationDate", Type: "string"},
		{Name: "lastUpdateDate", Type: "string"},
	}
}

func apptivoCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"customerId":     item["customerId"],
		"customerName":   item["customerName"],
		"customerNumber": item["customerNumber"],
		"emailAddress":   item["emailAddress"],
		"phoneNumber":    item["phoneNumber"],
		"website":        item["website"],
		"currencyCode":   item["currencyCode"],
		"statusName":     item["statusName"],
		"creationDate":   item["creationDate"],
		"lastUpdateDate": item["lastUpdateDate"],
	}
}

func apptivoContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"contactId":      item["contactId"],
		"firstName":      item["firstName"],
		"lastName":       item["lastName"],
		"fullName":       item["fullName"],
		"emailAddress":   item["emailAddress"],
		"phoneNumber":    item["phoneNumber"],
		"companyName":    item["companyName"],
		"creationDate":   item["creationDate"],
		"lastUpdateDate": item["lastUpdateDate"],
	}
}

func apptivoLeadRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"leadId":       item["leadId"],
		"firstName":    item["firstName"],
		"lastName":     item["lastName"],
		"companyName":  item["companyName"],
		"emailAddress": item["emailAddress"],
		"phoneNumber":  item["phoneNumber"],
		"leadSource":   item["leadSource"],
		"statusName":   item["statusName"],
		"creationDate": item["creationDate"],
	}
}

func apptivoOpportunityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"opportunityId":     item["opportunityId"],
		"opportunityName":   item["opportunityName"],
		"customerName":      item["customerName"],
		"salesStageName":    item["salesStageName"],
		"opportunityAmount": item["opportunityAmount"],
		"currencyCode":      item["currencyCode"],
		"closingDate":       item["closingDate"],
		"creationDate":      item["creationDate"],
		"lastUpdateDate":    item["lastUpdateDate"],
	}
}
