package fleetio

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Fleetio API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Fleetio list endpoint path segment (e.g. "vehicles").
	resource string
	// mapRecord flattens a raw Fleetio object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// fleetioStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in fleetioStreams; the read
// path is fully data-driven from this table.
var fleetioStreamEndpoints = map[string]streamEndpoint{
	"vehicles":        {resource: "vehicles", mapRecord: fleetioVehicleRecord},
	"contacts":        {resource: "contacts", mapRecord: fleetioContactRecord},
	"fuel_entries":    {resource: "fuel_entries", mapRecord: fleetioFuelEntryRecord},
	"issues":          {resource: "issues", mapRecord: fleetioIssueRecord},
	"service_entries": {resource: "service_entries", mapRecord: fleetioServiceEntryRecord},
}

// fleetioStreams returns the connector's published stream catalog. Fleetio
// objects carry an integer id and updated_at/created_at timestamps, so the
// primary key is ["id"] and the incremental cursor field is ["updated_at"]
// across the board.
func fleetioStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "vehicles",
			Description:  "Fleetio fleet vehicles.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       fleetioVehicleFields(),
		},
		{
			Name:         "contacts",
			Description:  "Fleetio contacts (people associated with the account).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       fleetioContactFields(),
		},
		{
			Name:         "fuel_entries",
			Description:  "Fleetio fuel entries (fuel consumption records).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       fleetioFuelEntryFields(),
		},
		{
			Name:         "issues",
			Description:  "Fleetio issues (vehicle or fleet issues).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       fleetioIssueFields(),
		},
		{
			Name:         "service_entries",
			Description:  "Fleetio service entries (maintenance/service logs).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       fleetioServiceEntryFields(),
		},
	}
}

func fleetioVehicleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "vin", Type: "string"},
		{Name: "make", Type: "string"},
		{Name: "model", Type: "string"},
		{Name: "year", Type: "integer"},
		{Name: "license_plate", Type: "string"},
		{Name: "vehicle_status_name", Type: "string"},
		{Name: "vehicle_type_name", Type: "string"},
		{Name: "current_meter_value", Type: "string"},
		{Name: "archived_at", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func fleetioContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "group_name", Type: "string"},
		{Name: "technician", Type: "boolean"},
		{Name: "employee", Type: "boolean"},
		{Name: "archived_at", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func fleetioFuelEntryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "vehicle_id", Type: "integer"},
		{Name: "vehicle_name", Type: "string"},
		{Name: "date", Type: "timestamp"},
		{Name: "us_gallons", Type: "string"},
		{Name: "cost", Type: "string"},
		{Name: "total_amount", Type: "string"},
		{Name: "meter_value", Type: "string"},
		{Name: "is_sample", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func fleetioIssueFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "number", Type: "integer"},
		{Name: "vehicle_id", Type: "integer"},
		{Name: "vehicle_name", Type: "string"},
		{Name: "summary", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "due_date", Type: "timestamp"},
		{Name: "resolved_at", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func fleetioServiceEntryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "vehicle_id", Type: "integer"},
		{Name: "vehicle_name", Type: "string"},
		{Name: "started_at", Type: "timestamp"},
		{Name: "completed_at", Type: "timestamp"},
		{Name: "total_amount", Type: "string"},
		{Name: "labor_subtotal", Type: "string"},
		{Name: "parts_subtotal", Type: "string"},
		{Name: "meter_value", Type: "string"},
		{Name: "is_sample", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func fleetioVehicleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"name":                item["name"],
		"vin":                 item["vin"],
		"make":                item["make"],
		"model":               item["model"],
		"year":                item["year"],
		"license_plate":       item["license_plate"],
		"vehicle_status_name": item["vehicle_status_name"],
		"vehicle_type_name":   item["vehicle_type_name"],
		"current_meter_value": item["current_meter_value"],
		"archived_at":         item["archived_at"],
		"created_at":          item["created_at"],
		"updated_at":          item["updated_at"],
	}
}

func fleetioContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"first_name":  item["first_name"],
		"last_name":   item["last_name"],
		"email":       item["email"],
		"group_name":  item["group_name"],
		"technician":  item["technician"],
		"employee":    item["employee"],
		"archived_at": item["archived_at"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func fleetioFuelEntryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"vehicle_id":   item["vehicle_id"],
		"vehicle_name": item["vehicle_name"],
		"date":         item["date"],
		"us_gallons":   item["us_gallons"],
		"cost":         item["cost"],
		"total_amount": item["total_amount"],
		"meter_value":  item["meter_value"],
		"is_sample":    item["is_sample"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func fleetioIssueRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"number":       item["number"],
		"vehicle_id":   item["vehicle_id"],
		"vehicle_name": item["vehicle_name"],
		"summary":      item["summary"],
		"description":  item["description"],
		"state":        item["state"],
		"due_date":     item["due_date"],
		"resolved_at":  item["resolved_at"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func fleetioServiceEntryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"vehicle_id":     item["vehicle_id"],
		"vehicle_name":   item["vehicle_name"],
		"started_at":     item["started_at"],
		"completed_at":   item["completed_at"],
		"total_amount":   item["total_amount"],
		"labor_subtotal": item["labor_subtotal"],
		"parts_subtotal": item["parts_subtotal"],
		"meter_value":    item["meter_value"],
		"is_sample":      item["is_sample"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
	}
}
