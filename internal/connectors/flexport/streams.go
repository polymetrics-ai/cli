package flexport

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Flexport API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Flexport list endpoint path segment (e.g. "products").
	resource string
	// mapRecord flattens a raw Flexport object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// flexportStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in flexportStreams; the read
// path is fully data-driven from this table.
var flexportStreamEndpoints = map[string]streamEndpoint{
	"companies": {resource: "companies", mapRecord: flexportCompanyRecord},
	"locations": {resource: "locations", mapRecord: flexportLocationRecord},
	"products":  {resource: "products", mapRecord: flexportProductRecord},
	"invoices":  {resource: "invoices", mapRecord: flexportInvoiceRecord},
	"shipments": {resource: "shipments", mapRecord: flexportShipmentRecord},
}

// flexportStreams returns the connector's published stream catalog. Every
// Flexport object exposes a string id and most carry an updated_at timestamp,
// which serves as the incremental cursor where present.
func flexportStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "companies",
			Description:  "Flexport companies (business entities you transact with).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       flexportCompanyFields(),
		},
		{
			Name:         "locations",
			Description:  "Flexport locations (addresses and facilities).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       flexportLocationFields(),
		},
		{
			Name:         "products",
			Description:  "Flexport products (SKU-level catalog entries).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       flexportProductFields(),
		},
		{
			Name:         "invoices",
			Description:  "Flexport invoices and their financial state.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       flexportInvoiceFields(),
		},
		{
			Name:         "shipments",
			Description:  "Flexport shipments and their logistics milestones.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       flexportShipmentFields(),
		},
	}
}

func flexportCompanyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "_object", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "legal_name", Type: "string"},
		{Name: "dba_name", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func flexportLocationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "_object", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "street_address", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "country_code", Type: "string"},
		{Name: "zip", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func flexportProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "_object", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "sku", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "hts_code", Type: "string"},
		{Name: "country_of_origin", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func flexportInvoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "_object", Type: "string"},
		{Name: "invoice_number", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "total", Type: "string"},
		{Name: "issued_date", Type: "string"},
		{Name: "due_date", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func flexportShipmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "_object", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "transportation_mode", Type: "string"},
		{Name: "freight_type", Type: "string"},
		{Name: "origin_port", Type: "string"},
		{Name: "destination_port", Type: "string"},
		{Name: "estimated_departure_date", Type: "string"},
		{Name: "estimated_arrival_date", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func flexportCompanyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"_object":    item["_object"],
		"name":       item["name"],
		"legal_name": item["legal_name"],
		"dba_name":   item["dba_name"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func flexportLocationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"_object":        item["_object"],
		"name":           item["name"],
		"street_address": item["street_address"],
		"city":           item["city"],
		"state":          item["state"],
		"country_code":   item["country_code"],
		"zip":            item["zip"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
	}
}

func flexportProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"_object":           item["_object"],
		"name":              item["name"],
		"sku":               item["sku"],
		"description":       item["description"],
		"hts_code":          item["hts_code"],
		"country_of_origin": item["country_of_origin"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
	}
}

func flexportInvoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"_object":        item["_object"],
		"invoice_number": item["invoice_number"],
		"status":         item["status"],
		"currency":       item["currency"],
		"total":          item["total"],
		"issued_date":    item["issued_date"],
		"due_date":       item["due_date"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
	}
}

func flexportShipmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                       item["id"],
		"_object":                  item["_object"],
		"status":                   item["status"],
		"transportation_mode":      item["transportation_mode"],
		"freight_type":             item["freight_type"],
		"origin_port":              item["origin_port"],
		"destination_port":         item["destination_port"],
		"estimated_departure_date": item["estimated_departure_date"],
		"estimated_arrival_date":   item["estimated_arrival_date"],
		"created_at":               item["created_at"],
		"updated_at":               item["updated_at"],
	}
}
