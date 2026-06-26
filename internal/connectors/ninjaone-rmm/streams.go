package ninjaonermm

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the NinjaOne v2 API resource path it reads
// from, the record mapper that flattens its objects, and whether the endpoint
// supports the after-cursor pagination (DefaultPaginator) or returns the full set
// in one request.
type streamEndpoint struct {
	// resource is the path under /v2 (e.g. "organizations").
	resource string
	// mapRecord flattens a raw NinjaOne object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// paginated reports whether the endpoint advances via pageSize + after.
	paginated bool
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streams(); the read path is fully
// data-driven from this table.
var streamEndpoints = map[string]streamEndpoint{
	"organizations": {resource: "organizations", mapRecord: organizationRecord, paginated: true},
	"devices":       {resource: "devices", mapRecord: deviceRecord, paginated: true},
	"locations":     {resource: "locations", mapRecord: locationRecord, paginated: true},
	"activities":    {resource: "activities", mapRecord: activityRecord, paginated: true},
	"policies":      {resource: "policies", mapRecord: policyRecord, paginated: false},
}

// streams returns the connector's published stream catalog. Every NinjaOne entity
// exposes a numeric id, so the primary key is ["id"] across the board.
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "organizations",
			Description: "NinjaOne managed organizations (clients).",
			PrimaryKey:  []string{"id"},
			Fields:      organizationFields(),
		},
		{
			Name:        "devices",
			Description: "NinjaOne managed devices (endpoints).",
			PrimaryKey:  []string{"id"},
			Fields:      deviceFields(),
		},
		{
			Name:        "locations",
			Description: "NinjaOne organization locations.",
			PrimaryKey:  []string{"id"},
			Fields:      locationFields(),
		},
		{
			Name:         "activities",
			Description:  "NinjaOne activity log entries.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"activityTime"},
			Fields:       activityFields(),
		},
		{
			Name:        "policies",
			Description: "NinjaOne automation policies.",
			PrimaryKey:  []string{"id"},
			Fields:      policyFields(),
		},
	}
}

func organizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "node_approval_mode", Type: "string"},
	}
}

func deviceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "organization_id", Type: "integer"},
		{Name: "location_id", Type: "integer"},
		{Name: "system_name", Type: "string"},
		{Name: "dns_name", Type: "string"},
		{Name: "node_class", Type: "string"},
		{Name: "offline", Type: "boolean"},
		{Name: "approval_status", Type: "string"},
	}
}

func locationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "organization_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "address", Type: "string"},
		{Name: "description", Type: "string"},
	}
}

func activityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "activityTime", Type: "number"},
		{Name: "device_id", Type: "integer"},
		{Name: "activity_type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "message", Type: "string"},
	}
}

func policyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "node_class", Type: "string"},
		{Name: "description", Type: "string"},
	}
}

func organizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"name":               item["name"],
		"description":        item["description"],
		"node_approval_mode": item["nodeApprovalMode"],
	}
}

func deviceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"organization_id": item["organizationId"],
		"location_id":     item["locationId"],
		"system_name":     item["systemName"],
		"dns_name":        item["dnsName"],
		"node_class":      item["nodeClass"],
		"offline":         item["offline"],
		"approval_status": item["approvalStatus"],
	}
}

func locationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"organization_id": item["organizationId"],
		"name":            item["name"],
		"address":         item["address"],
		"description":     item["description"],
	}
}

func activityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"activityTime":  item["activityTime"],
		"device_id":     item["deviceId"],
		"activity_type": item["activityType"],
		"status":        item["status"],
		"message":       item["message"],
	}
}

func policyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"node_class":  item["nodeClass"],
		"description": item["description"],
	}
}
