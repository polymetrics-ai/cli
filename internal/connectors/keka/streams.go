package keka

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Keka API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Keka list endpoint path segment (e.g. "hris/employees").
	resource string
	// mapRecord flattens a raw Keka object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// kekaStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in kekaStreams; the read path is
// fully data-driven from this table. Paths follow the public Keka API reference
// (HRIS, Time, and PSA modules).
var kekaStreamEndpoints = map[string]streamEndpoint{
	"employees":      {resource: "hris/employees", mapRecord: kekaEmployeeRecord},
	"attendance":     {resource: "time/attendance", mapRecord: kekaAttendanceRecord},
	"leave_types":    {resource: "time/leavetypes", mapRecord: kekaLeaveTypeRecord},
	"leave_requests": {resource: "time/leaverequests", mapRecord: kekaLeaveRequestRecord},
	"clients":        {resource: "psa/clients", mapRecord: kekaClientRecord},
	"projects":       {resource: "psa/projects", mapRecord: kekaProjectRecord},
}

// kekaStreams returns the connector's published stream catalog. Keka objects are
// identified by a string id; the API exposes only full-refresh sync, so no cursor
// fields are declared.
func kekaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "employees",
			Description: "Keka HRIS employees.",
			PrimaryKey:  []string{"id"},
			Fields:      kekaEmployeeFields(),
		},
		{
			Name:        "attendance",
			Description: "Keka time and attendance records.",
			PrimaryKey:  []string{"id"},
			Fields:      kekaAttendanceFields(),
		},
		{
			Name:        "leave_types",
			Description: "Keka leave types.",
			PrimaryKey:  []string{"id"},
			Fields:      kekaLeaveTypeFields(),
		},
		{
			Name:        "leave_requests",
			Description: "Keka leave requests.",
			PrimaryKey:  []string{"id"},
			Fields:      kekaLeaveRequestFields(),
		},
		{
			Name:        "clients",
			Description: "Keka PSA clients.",
			PrimaryKey:  []string{"id"},
			Fields:      kekaClientFields(),
		},
		{
			Name:        "projects",
			Description: "Keka PSA projects.",
			PrimaryKey:  []string{"id"},
			Fields:      kekaProjectFields(),
		},
	}
}

func kekaEmployeeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "employeeNumber", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "displayName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "jobTitle", Type: "string"},
		{Name: "department", Type: "string"},
		{Name: "employmentStatus", Type: "string"},
	}
}

func kekaAttendanceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "employeeId", Type: "string"},
		{Name: "attendanceDate", Type: "string"},
		{Name: "shiftStartTime", Type: "string"},
		{Name: "shiftEndTime", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "totalGrossHours", Type: "number"},
	}
}

func kekaLeaveTypeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "identifier", Type: "string"},
		{Name: "leaveTypeUnit", Type: "string"},
		{Name: "isActive", Type: "boolean"},
	}
}

func kekaLeaveRequestFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "employeeId", Type: "string"},
		{Name: "leaveTypeId", Type: "string"},
		{Name: "fromDate", Type: "string"},
		{Name: "toDate", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "dayCount", Type: "number"},
	}
}

func kekaClientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "isActive", Type: "boolean"},
	}
}

func kekaProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "clientId", Type: "string"},
		{Name: "billingType", Type: "string"},
		{Name: "status", Type: "string"},
	}
}

func kekaEmployeeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"employeeNumber":   item["employeeNumber"],
		"firstName":        item["firstName"],
		"lastName":         item["lastName"],
		"displayName":      item["displayName"],
		"email":            item["email"],
		"jobTitle":         item["jobTitle"],
		"department":       item["department"],
		"employmentStatus": item["employmentStatus"],
	}
}

func kekaAttendanceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"employeeId":      item["employeeId"],
		"attendanceDate":  item["attendanceDate"],
		"shiftStartTime":  item["shiftStartTime"],
		"shiftEndTime":    item["shiftEndTime"],
		"status":          item["status"],
		"totalGrossHours": item["totalGrossHours"],
	}
}

func kekaLeaveTypeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"identifier":    item["identifier"],
		"leaveTypeUnit": item["leaveTypeUnit"],
		"isActive":      item["isActive"],
	}
}

func kekaLeaveRequestRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"employeeId":  item["employeeId"],
		"leaveTypeId": item["leaveTypeId"],
		"fromDate":    item["fromDate"],
		"toDate":      item["toDate"],
		"status":      item["status"],
		"dayCount":    item["dayCount"],
	}
}

func kekaClientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"name":     item["name"],
		"code":     item["code"],
		"isActive": item["isActive"],
	}
}

func kekaProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"code":        item["code"],
		"clientId":    item["clientId"],
		"billingType": item["billingType"],
		"status":      item["status"],
	}
}
