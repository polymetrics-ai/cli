package hubspot

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the HubSpot CRM v3 object segment it reads
// from (relative to /crm/v3/objects) and the record mapper that flattens its
// objects. Adding a stream means adding one entry here plus a Stream definition
// in hubspotStreams; the read path is fully data-driven from this table.
type streamEndpoint struct {
	// object is the CRM v3 object name (e.g. "contacts"); the resource path is
	// crm/v3/objects/<object>.
	object string
	// properties is the set of object properties requested via the `properties`
	// query param so HubSpot returns them in each record's properties map.
	properties []string
	// mapRecord flattens a raw CRM object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// hubspotStreamEndpoints is the per-stream routing table for the core CRM
// objects. Each CRM v3 object list endpoint shares the same response shape
// ({results:[...], paging:{next:{after}}}) so the read loop is identical across
// streams; only the object segment and the requested properties differ.
var hubspotStreamEndpoints = map[string]streamEndpoint{
	"contacts": {
		object:     "contacts",
		properties: []string{"email", "firstname", "lastname", "phone", "company", "lifecyclestage", "createdate", "lastmodifieddate"},
		mapRecord:  hubspotContactRecord,
	},
	"companies": {
		object:     "companies",
		properties: []string{"name", "domain", "industry", "city", "country", "numberofemployees", "createdate", "hs_lastmodifieddate"},
		mapRecord:  hubspotCompanyRecord,
	},
	"deals": {
		object:     "deals",
		properties: []string{"dealname", "amount", "dealstage", "pipeline", "closedate", "createdate", "hs_lastmodifieddate"},
		mapRecord:  hubspotDealRecord,
	},
	"tickets": {
		object:     "tickets",
		properties: []string{"subject", "content", "hs_pipeline", "hs_pipeline_stage", "hs_ticket_priority", "createdate", "hs_lastmodifieddate"},
		mapRecord:  hubspotTicketRecord,
	},
}

// hubspotStreams returns the connector's published stream catalog. Every CRM v3
// object exposes a string id and createdAt/updatedAt timestamps, so the primary
// key is ["id"] and the incremental cursor field is ["updatedAt"] across the
// board.
func hubspotStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "HubSpot CRM contacts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       hubspotContactFields(),
		},
		{
			Name:         "companies",
			Description:  "HubSpot CRM companies.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       hubspotCompanyFields(),
		},
		{
			Name:         "deals",
			Description:  "HubSpot CRM deals.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       hubspotDealFields(),
		},
		{
			Name:         "tickets",
			Description:  "HubSpot CRM tickets.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       hubspotTicketFields(),
		},
	}
}

func hubspotContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "firstname", Type: "string"},
		{Name: "lastname", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "lifecyclestage", Type: "string"},
		{Name: "createdate", Type: "string"},
		{Name: "lastmodifieddate", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
		{Name: "archived", Type: "boolean"},
	}
}

func hubspotCompanyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "domain", Type: "string"},
		{Name: "industry", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "numberofemployees", Type: "string"},
		{Name: "createdate", Type: "string"},
		{Name: "hs_lastmodifieddate", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
		{Name: "archived", Type: "boolean"},
	}
}

func hubspotDealFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "dealname", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "dealstage", Type: "string"},
		{Name: "pipeline", Type: "string"},
		{Name: "closedate", Type: "string"},
		{Name: "createdate", Type: "string"},
		{Name: "hs_lastmodifieddate", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
		{Name: "archived", Type: "boolean"},
	}
}

func hubspotTicketFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "content", Type: "string"},
		{Name: "hs_pipeline", Type: "string"},
		{Name: "hs_pipeline_stage", Type: "string"},
		{Name: "hs_ticket_priority", Type: "string"},
		{Name: "createdate", Type: "string"},
		{Name: "hs_lastmodifieddate", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
		{Name: "archived", Type: "boolean"},
	}
}

// flattenObject builds a Record from a raw CRM v3 object. HubSpot nests the
// user-facing fields under "properties"; we lift the listed property keys to the
// top level alongside the envelope fields (id, createdAt, updatedAt, archived)
// so downstream consumers see a flat record. The id and the raw properties map
// are always preserved.
func flattenObject(item map[string]any, propertyKeys []string) connectors.Record {
	record := connectors.Record{
		"id":        item["id"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
		"archived":  item["archived"],
	}
	props, _ := item["properties"].(map[string]any)
	for _, key := range propertyKeys {
		if props != nil {
			record[key] = props[key]
		}
	}
	return record
}

func hubspotContactRecord(item map[string]any) connectors.Record {
	return flattenObject(item, []string{"email", "firstname", "lastname", "phone", "company", "lifecyclestage", "createdate", "lastmodifieddate"})
}

func hubspotCompanyRecord(item map[string]any) connectors.Record {
	return flattenObject(item, []string{"name", "domain", "industry", "city", "country", "numberofemployees", "createdate", "hs_lastmodifieddate"})
}

func hubspotDealRecord(item map[string]any) connectors.Record {
	return flattenObject(item, []string{"dealname", "amount", "dealstage", "pipeline", "closedate", "createdate", "hs_lastmodifieddate"})
}

func hubspotTicketRecord(item map[string]any) connectors.Record {
	return flattenObject(item, []string{"subject", "content", "hs_pipeline", "hs_pipeline_stage", "hs_ticket_priority", "createdate", "hs_lastmodifieddate"})
}
