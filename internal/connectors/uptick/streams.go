package uptick

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Uptick API resource path segment
// (relative to {base_url}/api/<apiVersion>/) and the record mapper that projects
// its objects. The read path is fully data-driven from this table.
type streamEndpoint struct {
	// resource is the Uptick list endpoint path segment (e.g. "clients").
	resource string
	// sparseType is the value of the fields[<Type>] sparse-fieldset request param
	// Uptick uses to bound the response columns (e.g. "Client").
	sparseType string
	// fields is the comma-joined column list requested via fields[<Type>].
	fields string
	// mapRecord projects a raw Uptick object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// uptickStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in uptickStreams.
var uptickStreamEndpoints = map[string]streamEndpoint{
	"tasks": {
		resource:   "tasks",
		sparseType: "Task",
		fields:     "id,created,updated,deleted,ref,description,is_active,name,due,status,client,property,priority",
		mapRecord:  uptickTaskRecord,
	},
	"clients": {
		resource:   "clients",
		sparseType: "Client",
		fields:     "id,created,updated,ref,name,is_active,contact_name,contact_email,contact_phone_bh,address,notes",
		mapRecord:  uptickClientRecord,
	},
	"properties": {
		resource:   "properties",
		sparseType: "Property",
		fields:     "id,created,updated,name,ref,address,timezone,status,coords",
		mapRecord:  uptickPropertyRecord,
	},
	"invoices": {
		resource:   "invoices",
		sparseType: "Invoice",
		fields:     "id,created,updated,number,ref,description,currency,date,due_date,status,subtotal,gst,total,is_overdue,is_sent,property,task",
		mapRecord:  uptickInvoiceRecord,
	},
	"assets": {
		resource:   "assets",
		sparseType: "Asset",
		fields:     "id,created,updated,deleted,is_active,ref,uptick_ref,label,location,status,make,model,size,barcode,serviced_date,property,type,variant",
		mapRecord:  uptickAssetRecord,
	},
}

// uptickStreams returns the connector's published stream catalog. Every Uptick
// object exposes an integer id and an RFC3339 `updated` timestamp, so the primary
// key is ["id"] and the incremental cursor field is ["updated"] across the board.
func uptickStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "tasks",
			Description:  "Uptick tasks (jobs/work orders).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       uptickTaskFields(),
		},
		{
			Name:         "clients",
			Description:  "Uptick clients (customers).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       uptickClientFields(),
		},
		{
			Name:         "properties",
			Description:  "Uptick properties (sites/buildings).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       uptickPropertyFields(),
		},
		{
			Name:         "invoices",
			Description:  "Uptick invoices.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       uptickInvoiceFields(),
		},
		{
			Name:         "assets",
			Description:  "Uptick assets (inspected equipment).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated"},
			Fields:       uptickAssetFields(),
		},
	}
}

func uptickTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "created", Type: "timestamp"},
		{Name: "updated", Type: "timestamp"},
		{Name: "deleted", Type: "timestamp"},
		{Name: "ref", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "name", Type: "string"},
		{Name: "due", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "client", Type: "string"},
		{Name: "property", Type: "string"},
		{Name: "priority", Type: "string"},
	}
}

func uptickClientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "created", Type: "timestamp"},
		{Name: "updated", Type: "timestamp"},
		{Name: "ref", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "contact_name", Type: "string"},
		{Name: "contact_email", Type: "string"},
		{Name: "contact_phone_bh", Type: "string"},
		{Name: "address", Type: "string"},
		{Name: "notes", Type: "string"},
	}
}

func uptickPropertyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "created", Type: "timestamp"},
		{Name: "updated", Type: "timestamp"},
		{Name: "name", Type: "string"},
		{Name: "ref", Type: "string"},
		{Name: "address", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "coords", Type: "string"},
	}
}

func uptickInvoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "created", Type: "timestamp"},
		{Name: "updated", Type: "timestamp"},
		{Name: "number", Type: "string"},
		{Name: "ref", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "due_date", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "subtotal", Type: "string"},
		{Name: "gst", Type: "string"},
		{Name: "total", Type: "string"},
		{Name: "is_overdue", Type: "boolean"},
		{Name: "is_sent", Type: "boolean"},
		{Name: "property", Type: "string"},
		{Name: "task", Type: "string"},
	}
}

func uptickAssetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "created", Type: "timestamp"},
		{Name: "updated", Type: "timestamp"},
		{Name: "deleted", Type: "timestamp"},
		{Name: "is_active", Type: "boolean"},
		{Name: "ref", Type: "string"},
		{Name: "uptick_ref", Type: "string"},
		{Name: "label", Type: "string"},
		{Name: "location", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "make", Type: "string"},
		{Name: "model", Type: "string"},
		{Name: "size", Type: "string"},
		{Name: "barcode", Type: "string"},
		{Name: "serviced_date", Type: "string"},
		{Name: "property", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "variant", Type: "string"},
	}
}

func uptickTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"created":     item["created"],
		"updated":     item["updated"],
		"deleted":     item["deleted"],
		"ref":         item["ref"],
		"description": item["description"],
		"is_active":   item["is_active"],
		"name":        item["name"],
		"due":         item["due"],
		"status":      item["status"],
		"client":      item["client"],
		"property":    item["property"],
		"priority":    item["priority"],
	}
}

func uptickClientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"created":          item["created"],
		"updated":          item["updated"],
		"ref":              item["ref"],
		"name":             item["name"],
		"is_active":        item["is_active"],
		"contact_name":     item["contact_name"],
		"contact_email":    item["contact_email"],
		"contact_phone_bh": item["contact_phone_bh"],
		"address":          item["address"],
		"notes":            item["notes"],
	}
}

func uptickPropertyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"created":  item["created"],
		"updated":  item["updated"],
		"name":     item["name"],
		"ref":      item["ref"],
		"address":  item["address"],
		"timezone": item["timezone"],
		"status":   item["status"],
		"coords":   item["coords"],
	}
}

func uptickInvoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"created":     item["created"],
		"updated":     item["updated"],
		"number":      item["number"],
		"ref":         item["ref"],
		"description": item["description"],
		"currency":    item["currency"],
		"date":        item["date"],
		"due_date":    item["due_date"],
		"status":      item["status"],
		"subtotal":    item["subtotal"],
		"gst":         item["gst"],
		"total":       item["total"],
		"is_overdue":  item["is_overdue"],
		"is_sent":     item["is_sent"],
		"property":    item["property"],
		"task":        item["task"],
	}
}

func uptickAssetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"created":       item["created"],
		"updated":       item["updated"],
		"deleted":       item["deleted"],
		"is_active":     item["is_active"],
		"ref":           item["ref"],
		"uptick_ref":    item["uptick_ref"],
		"label":         item["label"],
		"location":      item["location"],
		"status":        item["status"],
		"make":          item["make"],
		"model":         item["model"],
		"size":          item["size"],
		"barcode":       item["barcode"],
		"serviced_date": item["serviced_date"],
		"property":      item["property"],
		"type":          item["type"],
		"variant":       item["variant"],
	}
}
