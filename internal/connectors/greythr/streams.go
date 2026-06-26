package greythr

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the greytHR API resource path (relative
// to base_url), the JSON path the records array lives at, the page number the
// API starts counting from, and the record mapper that flattens its objects.
//
// greytHR's two response shapes are captured by recordsPath: employee resources
// wrap rows in {"data":[...]} (recordsPath "data"), while the Users List returns
// a bare top-level array (recordsPath "" selects the root).
type streamEndpoint struct {
	resource    string
	recordsPath string
	startPage   int
	mapRecord   func(map[string]any) connectors.Record
}

// greythrStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in greythrStreams; the read
// path is fully data-driven from this table.
var greythrStreamEndpoints = map[string]streamEndpoint{
	"employees": {
		resource:    "employee/v2/employees",
		recordsPath: "data",
		startPage:   0,
		mapRecord:   employeeRecord,
	},
	"employees_profile": {
		resource:    "employee/v2/employees/profile",
		recordsPath: "data",
		startPage:   1,
		mapRecord:   employeeProfileRecord,
	},
	"employees_work": {
		resource:    "employee/v2/employees/work",
		recordsPath: "data",
		startPage:   1,
		mapRecord:   employeeWorkRecord,
	},
	"employees_bank": {
		resource:    "employee/v2/employees/bank",
		recordsPath: "data",
		startPage:   1,
		mapRecord:   employeeBankRecord,
	},
	"users": {
		resource:    "user/v2/users",
		recordsPath: "",
		startPage:   1,
		mapRecord:   userRecord,
	},
}

// greythrStreams returns the connector's published stream catalog. greytHR is
// full-refresh only (no incremental cursor), so CursorFields are empty.
func greythrStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "employees",
			Description: "greytHR employees directory.",
			PrimaryKey:  []string{"employeeId"},
			Fields:      employeeFields(),
		},
		{
			Name:        "employees_profile",
			Description: "greytHR employee social/profile details.",
			PrimaryKey:  []string{"employeeId"},
			Fields:      employeeProfileFields(),
		},
		{
			Name:        "employees_work",
			Description: "greytHR employee work details (confirmation, notice period).",
			PrimaryKey:  []string{"employeeId"},
			Fields:      employeeWorkFields(),
		},
		{
			Name:        "employees_bank",
			Description: "greytHR employee bank account details.",
			PrimaryKey:  []string{"employeeId"},
			Fields:      employeeBankFields(),
		},
		{
			Name:        "users",
			Description: "greytHR user accounts list.",
			PrimaryKey:  []string{"id"},
			Fields:      userFields(),
		},
	}
}

func employeeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "employeeId", Type: "integer"},
		{Name: "employeeNo", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "personalEmail", Type: "string"},
		{Name: "mobile", Type: "string"},
		{Name: "gender", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "status", Type: "integer"},
		{Name: "leftorg", Type: "boolean"},
		{Name: "dateOfJoin", Type: "string"},
		{Name: "dateOfBirth", Type: "string"},
		{Name: "leavingDate", Type: "string"},
		{Name: "lastModified", Type: "string"},
		{Name: "probationPeriod", Type: "integer"},
	}
}

func employeeProfileFields() []connectors.Field {
	return []connectors.Field{
		{Name: "employeeId", Type: "integer"},
		{Name: "nickname", Type: "string"},
		{Name: "biography", Type: "string"},
		{Name: "wishDOB", Type: "string"},
		{Name: "linkedIn", Type: "string"},
		{Name: "twitter", Type: "string"},
		{Name: "facebook", Type: "string"},
	}
}

func employeeWorkFields() []connectors.Field {
	return []connectors.Field{
		{Name: "employeeId", Type: "integer"},
		{Name: "confirmDate", Type: "string"},
		{Name: "noticePeriod", Type: "integer"},
		{Name: "onboardingStatus", Type: "string"},
		{Name: "extendedProbationDays", Type: "integer"},
	}
}

func employeeBankFields() []connectors.Field {
	return []connectors.Field{
		{Name: "employeeId", Type: "integer"},
		{Name: "bankName", Type: "integer"},
		{Name: "bankBranch", Type: "integer"},
		{Name: "branchCode", Type: "string"},
		{Name: "bankAccountNumber", Type: "string"},
		{Name: "salaryPaymentMode", Type: "integer"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "type", Type: "integer"},
		{Name: "userName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "admin", Type: "boolean"},
		{Name: "deleted", Type: "boolean"},
		{Name: "systemPassword", Type: "boolean"},
	}
}

func employeeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"employeeId":      item["employeeId"],
		"employeeNo":      item["employeeNo"],
		"name":            item["name"],
		"firstName":       item["firstName"],
		"lastName":        item["lastName"],
		"email":           item["email"],
		"personalEmail":   item["personalEmail"],
		"mobile":          item["mobile"],
		"gender":          item["gender"],
		"title":           item["title"],
		"status":          item["status"],
		"leftorg":         item["leftorg"],
		"dateOfJoin":      item["dateOfJoin"],
		"dateOfBirth":     item["dateOfBirth"],
		"leavingDate":     item["leavingDate"],
		"lastModified":    item["lastModified"],
		"probationPeriod": item["probationPeriod"],
	}
}

func employeeProfileRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"employeeId": item["employeeId"],
		"nickname":   item["nickname"],
		"biography":  item["biography"],
		"wishDOB":    item["wishDOB"],
		"linkedIn":   item["linkedIn"],
		"twitter":    item["twitter"],
		"facebook":   item["facebook"],
	}
}

func employeeWorkRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"employeeId":            item["employeeId"],
		"confirmDate":           item["confirmDate"],
		"noticePeriod":          item["noticePeriod"],
		"onboardingStatus":      item["onboardingStatus"],
		"extendedProbationDays": item["extendedProbationDays"],
	}
}

func employeeBankRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"employeeId":        item["employeeId"],
		"bankName":          item["bankName"],
		"bankBranch":        item["bankBranch"],
		"branchCode":        item["branchCode"],
		"bankAccountNumber": item["bankAccountNumber"],
		"salaryPaymentMode": item["salaryPaymentMode"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"type":           item["type"],
		"userName":       item["userName"],
		"email":          item["email"],
		"admin":          item["admin"],
		"deleted":        item["deleted"],
		"systemPassword": item["systemPassword"],
	}
}
