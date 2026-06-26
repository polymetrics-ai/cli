package hoorayhr

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a catalog stream name to the HoorayHR API resource path
// (relative to base_url) it reads from, and the record mapper that selects the
// fields the connector publishes. HoorayHR responses are top-level JSON arrays
// (record selector field_path []), so there is no records wrapper key.
type streamEndpoint struct {
	// resource is the HoorayHR API path segment (e.g. "sick-leave").
	resource string
	// mapRecord projects a raw HoorayHR object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// hoorayhrStreamEndpoints is the per-stream routing table. The catalog stream
// names follow the Airbyte source (sick-leaves, time-off, ...) while the API
// paths are the underlying HoorayHR resources (/sick-leave, /time-off, ...).
var hoorayhrStreamEndpoints = map[string]streamEndpoint{
	"users":       {resource: "users", mapRecord: hoorayhrUserRecord},
	"time-off":    {resource: "time-off", mapRecord: hoorayhrTimeOffRecord},
	"leave-types": {resource: "leave-types", mapRecord: hoorayhrLeaveTypeRecord},
	"sick-leaves": {resource: "sick-leave", mapRecord: hoorayhrSickLeaveRecord},
}

// hoorayhrStreams returns the connector's published stream catalog. Every
// HoorayHR resource is keyed by a numeric "id"; the source only supports full
// refresh, so no incremental cursor fields are advertised.
func hoorayhrStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "users",
			Description: "HoorayHR users (employees) in the company.",
			PrimaryKey:  []string{"id"},
			Fields:      hoorayhrUserFields(),
		},
		{
			Name:        "time-off",
			Description: "HoorayHR time-off (leave) requests.",
			PrimaryKey:  []string{"id"},
			Fields:      hoorayhrTimeOffFields(),
		},
		{
			Name:        "leave-types",
			Description: "HoorayHR leave types configured for the company.",
			PrimaryKey:  []string{"id"},
			Fields:      hoorayhrLeaveTypeFields(),
		},
		{
			Name:        "sick-leaves",
			Description: "HoorayHR sick-leave records.",
			PrimaryKey:  []string{"id"},
			Fields:      hoorayhrSickLeaveFields(),
		},
	}
}

func hoorayhrUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "email", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "jobTitle", Type: "string"},
		{Name: "companyId", Type: "integer"},
		{Name: "isAdmin", Type: "boolean"},
		{Name: "companyStartDate", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func hoorayhrTimeOffFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "userId", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "start", Type: "string"},
		{Name: "end", Type: "string"},
		{Name: "leaveTypeId", Type: "integer"},
		{Name: "timeOffType", Type: "string"},
		{Name: "leaveUnit", Type: "string"},
		{Name: "notes", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func hoorayhrLeaveTypeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "icon", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "budget", Type: "number"},
		{Name: "default", Type: "boolean"},
		{Name: "unpaidLeave", Type: "boolean"},
		{Name: "leaveInDays", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func hoorayhrSickLeaveFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "userId", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "percentage", Type: "number"},
		{Name: "notes", Type: "string"},
		{Name: "actualStart", Type: "string"},
		{Name: "actualReturn", Type: "string"},
		{Name: "reportedStart", Type: "string"},
		{Name: "reportedReturn", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func hoorayhrUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"email":            item["email"],
		"firstName":        item["firstName"],
		"lastName":         item["lastName"],
		"status":           item["status"],
		"jobTitle":         item["jobTitle"],
		"companyId":        item["companyId"],
		"isAdmin":          item["isAdmin"],
		"companyStartDate": item["companyStartDate"],
		"createdAt":        item["createdAt"],
		"updatedAt":        item["updatedAt"],
	}
}

func hoorayhrTimeOffRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"userId":      item["userId"],
		"status":      item["status"],
		"start":       item["start"],
		"end":         item["end"],
		"leaveTypeId": item["leaveTypeId"],
		"timeOffType": item["timeOffType"],
		"leaveUnit":   item["leaveUnit"],
		"notes":       item["notes"],
		"createdAt":   item["createdAt"],
		"updatedAt":   item["updatedAt"],
	}
}

func hoorayhrLeaveTypeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"icon":        item["icon"],
		"color":       item["color"],
		"budget":      item["budget"],
		"default":     item["default"],
		"unpaidLeave": item["unpaidLeave"],
		"leaveInDays": item["leaveInDays"],
		"createdAt":   item["createdAt"],
		"updatedAt":   item["updatedAt"],
	}
}

func hoorayhrSickLeaveRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"userId":         item["userId"],
		"status":         item["status"],
		"percentage":     item["percentage"],
		"notes":          item["notes"],
		"actualStart":    item["actualStart"],
		"actualReturn":   item["actualReturn"],
		"reportedStart":  item["reportedStart"],
		"reportedReturn": item["reportedReturn"],
		"createdAt":      item["createdAt"],
		"updatedAt":      item["updatedAt"],
	}
}
