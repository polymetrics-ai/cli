package easypost

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the EasyPost API resource path (relative
// to base_url) it reads from, the JSON key that holds the records array in a
// list response (EasyPost names it after the resource, e.g. "shipments"), and
// the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the EasyPost list endpoint path segment (e.g. "shipments").
	resource string
	// arrayKey is the JSON object key holding the records array in the list
	// response. For EasyPost this equals the resource name.
	arrayKey string
	// mapRecord flattens a raw EasyPost object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// easypostStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in easypostStreams; the read
// path is fully data-driven from this table.
var easypostStreamEndpoints = map[string]streamEndpoint{
	"shipments":  {resource: "shipments", arrayKey: "shipments", mapRecord: easypostShipmentRecord},
	"trackers":   {resource: "trackers", arrayKey: "trackers", mapRecord: easypostTrackerRecord},
	"addresses":  {resource: "addresses", arrayKey: "addresses", mapRecord: easypostAddressRecord},
	"parcels":    {resource: "parcels", arrayKey: "parcels", mapRecord: easypostParcelRecord},
	"insurances": {resource: "insurances", arrayKey: "insurances", mapRecord: easypostInsuranceRecord},
}

// easypostStreams returns the connector's published stream catalog. Every
// EasyPost object exposes a string id and an RFC3339 created_at/updated_at
// timestamp, so the primary key is ["id"] and the incremental cursor field is
// ["created_at"] across the board.
func easypostStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "shipments",
			Description:  "EasyPost shipments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       easypostShipmentFields(),
		},
		{
			Name:         "trackers",
			Description:  "EasyPost trackers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       easypostTrackerFields(),
		},
		{
			Name:         "addresses",
			Description:  "EasyPost addresses.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       easypostAddressFields(),
		},
		{
			Name:         "parcels",
			Description:  "EasyPost parcels.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       easypostParcelFields(),
		},
		{
			Name:         "insurances",
			Description:  "EasyPost insurances.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       easypostInsuranceFields(),
		},
	}
}

func easypostShipmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "mode", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "status", Type: "string"},
		{Name: "tracking_code", Type: "string"},
		{Name: "reference", Type: "string"},
		{Name: "batch_id", Type: "string"},
		{Name: "batch_status", Type: "string"},
		{Name: "is_return", Type: "boolean"},
	}
}

func easypostTrackerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "mode", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "tracking_code", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "status_detail", Type: "string"},
		{Name: "carrier", Type: "string"},
		{Name: "signed_by", Type: "string"},
		{Name: "shipment_id", Type: "string"},
		{Name: "est_delivery_date", Type: "timestamp"},
	}
}

func easypostAddressFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "mode", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "name", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "street1", Type: "string"},
		{Name: "street2", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "zip", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "residential", Type: "boolean"},
	}
}

func easypostParcelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "mode", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "length", Type: "number"},
		{Name: "width", Type: "number"},
		{Name: "height", Type: "number"},
		{Name: "weight", Type: "number"},
		{Name: "predefined_package", Type: "string"},
	}
}

func easypostInsuranceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "object", Type: "string"},
		{Name: "mode", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "reference", Type: "string"},
		{Name: "amount", Type: "string"},
		{Name: "tracking_code", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "provider", Type: "string"},
		{Name: "shipment_id", Type: "string"},
	}
}

func easypostShipmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"object":        item["object"],
		"mode":          item["mode"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
		"status":        item["status"],
		"tracking_code": item["tracking_code"],
		"reference":     item["reference"],
		"batch_id":      item["batch_id"],
		"batch_status":  item["batch_status"],
		"is_return":     item["is_return"],
	}
}

func easypostTrackerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"object":            item["object"],
		"mode":              item["mode"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
		"tracking_code":     item["tracking_code"],
		"status":            item["status"],
		"status_detail":     item["status_detail"],
		"carrier":           item["carrier"],
		"signed_by":         item["signed_by"],
		"shipment_id":       item["shipment_id"],
		"est_delivery_date": item["est_delivery_date"],
	}
}

func easypostAddressRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"object":      item["object"],
		"mode":        item["mode"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
		"name":        item["name"],
		"company":     item["company"],
		"street1":     item["street1"],
		"street2":     item["street2"],
		"city":        item["city"],
		"state":       item["state"],
		"zip":         item["zip"],
		"country":     item["country"],
		"phone":       item["phone"],
		"email":       item["email"],
		"residential": item["residential"],
	}
}

func easypostParcelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"object":             item["object"],
		"mode":               item["mode"],
		"created_at":         item["created_at"],
		"updated_at":         item["updated_at"],
		"length":             item["length"],
		"width":              item["width"],
		"height":             item["height"],
		"weight":             item["weight"],
		"predefined_package": item["predefined_package"],
	}
}

func easypostInsuranceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"object":        item["object"],
		"mode":          item["mode"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
		"reference":     item["reference"],
		"amount":        item["amount"],
		"tracking_code": item["tracking_code"],
		"status":        item["status"],
		"provider":      item["provider"],
		"shipment_id":   item["shipment_id"],
	}
}
