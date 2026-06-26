package hubplanner

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Hubplanner API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Hubplanner list endpoint path segment (e.g. "resource").
	resource string
	// mapRecord flattens a raw Hubplanner object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// hubplannerStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in hubplannerStreams; the
// read path is fully data-driven from this table.
//
// Stream names are the public, pluralized names (matching the Airbyte source);
// the resource path is the Hubplanner singular endpoint segment.
var hubplannerStreamEndpoints = map[string]streamEndpoint{
	"resources":     {resource: "resource", mapRecord: hubplannerResourceRecord},
	"projects":      {resource: "project", mapRecord: hubplannerProjectRecord},
	"clients":       {resource: "client", mapRecord: hubplannerClientRecord},
	"events":        {resource: "event", mapRecord: hubplannerEventRecord},
	"holidays":      {resource: "holiday", mapRecord: hubplannerHolidayRecord},
	"bookings":      {resource: "booking", mapRecord: hubplannerBookingRecord},
	"billing_rates": {resource: "billingrate", mapRecord: hubplannerBillingRateRecord},
}

// hubplannerStreams returns the connector's published stream catalog. Every
// Hubplanner object exposes a string `_id`, so the primary key is ["_id"]. The
// API only supports full_refresh, so no cursor fields are declared.
func hubplannerStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "resources",
			Description: "Hubplanner resources (people and machines that can be scheduled).",
			PrimaryKey:  []string{"_id"},
			Fields:      hubplannerResourceFields(),
		},
		{
			Name:        "projects",
			Description: "Hubplanner projects.",
			PrimaryKey:  []string{"_id"},
			Fields:      hubplannerProjectFields(),
		},
		{
			Name:        "clients",
			Description: "Hubplanner clients/customers.",
			PrimaryKey:  []string{"_id"},
			Fields:      hubplannerClientFields(),
		},
		{
			Name:        "events",
			Description: "Hubplanner events.",
			PrimaryKey:  []string{"_id"},
			Fields:      hubplannerEventFields(),
		},
		{
			Name:        "holidays",
			Description: "Hubplanner public holidays.",
			PrimaryKey:  []string{"_id"},
			Fields:      hubplannerHolidayFields(),
		},
		{
			Name:        "bookings",
			Description: "Hubplanner bookings (scheduled allocations of resources to projects).",
			PrimaryKey:  []string{"_id"},
			Fields:      hubplannerBookingFields(),
		},
		{
			Name:        "billing_rates",
			Description: "Hubplanner billing rates.",
			PrimaryKey:  []string{"_id"},
			Fields:      hubplannerBillingRateFields(),
		},
	}
}

func hubplannerResourceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "note", Type: "string"},
		{Name: "createdDate", Type: "string"},
	}
}

func hubplannerProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "projectCode", Type: "string"},
		{Name: "budgetHours", Type: "number"},
		{Name: "budgetCashAmount", Type: "number"},
		{Name: "budgetCurrency", Type: "string"},
		{Name: "note", Type: "string"},
		{Name: "createdDate", Type: "string"},
		{Name: "updatedDate", Type: "string"},
	}
}

func hubplannerClientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "note", Type: "string"},
		{Name: "createdDate", Type: "string"},
	}
}

func hubplannerEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "start", Type: "string"},
		{Name: "end", Type: "string"},
		{Name: "note", Type: "string"},
	}
}

func hubplannerHolidayFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "start", Type: "string"},
		{Name: "end", Type: "string"},
		{Name: "holidayGroup", Type: "string"},
	}
}

func hubplannerBookingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "resource", Type: "string"},
		{Name: "project", Type: "string"},
		{Name: "start", Type: "string"},
		{Name: "end", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "note", Type: "string"},
	}
}

func hubplannerBillingRateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "rate", Type: "number"},
		{Name: "currency", Type: "string"},
		{Name: "default", Type: "boolean"},
	}
}

func hubplannerResourceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":         item["_id"],
		"firstName":   item["firstName"],
		"lastName":    item["lastName"],
		"email":       item["email"],
		"status":      item["status"],
		"role":        item["role"],
		"type":        item["type"],
		"note":        item["note"],
		"createdDate": item["createdDate"],
	}
}

func hubplannerProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":              item["_id"],
		"name":             item["name"],
		"status":           item["status"],
		"projectCode":      item["projectCode"],
		"budgetHours":      item["budgetHours"],
		"budgetCashAmount": item["budgetCashAmount"],
		"budgetCurrency":   item["budgetCurrency"],
		"note":             item["note"],
		"createdDate":      item["createdDate"],
		"updatedDate":      item["updatedDate"],
	}
}

func hubplannerClientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":         item["_id"],
		"name":        item["name"],
		"email":       item["email"],
		"phone":       item["phone"],
		"note":        item["note"],
		"createdDate": item["createdDate"],
	}
}

func hubplannerEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":   item["_id"],
		"name":  item["name"],
		"type":  item["type"],
		"start": item["start"],
		"end":   item["end"],
		"note":  item["note"],
	}
}

func hubplannerHolidayRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":          item["_id"],
		"name":         item["name"],
		"date":         item["date"],
		"start":        item["start"],
		"end":          item["end"],
		"holidayGroup": item["holidayGroup"],
	}
}

func hubplannerBookingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":      item["_id"],
		"resource": item["resource"],
		"project":  item["project"],
		"start":    item["start"],
		"end":      item["end"],
		"state":    item["state"],
		"category": item["category"],
		"note":     item["note"],
	}
}

func hubplannerBillingRateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":      item["_id"],
		"name":     item["name"],
		"rate":     item["rate"],
		"currency": item["currency"],
		"default":  item["default"],
	}
}
