package myhours

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the My Hours API resource path (relative
// to base_url) and the record mapper that flattens its objects. timeRanged marks
// the time_logs stream, which is fetched in date windows rather than a single
// request.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "Clients", "Users/getAll").
	resource string
	// mapRecord flattens a raw My Hours object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// timeRanged is true for streams driven by DateFrom/DateTo windowing.
	timeRanged bool
}

// myHoursStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in myHoursStreams. Every list
// endpoint returns a top-level JSON array, so the record selector path is the
// document root ("").
var myHoursStreamEndpoints = map[string]streamEndpoint{
	"clients":   {resource: "Clients", mapRecord: clientRecord},
	"projects":  {resource: "Projects/getAll", mapRecord: projectRecord},
	"users":     {resource: "Users/getAll", mapRecord: userRecord},
	"tags":      {resource: "Tags", mapRecord: tagRecord},
	"time_logs": {resource: "Reports/activity", mapRecord: timeLogRecord, timeRanged: true},
}

// myHoursStreams returns the connector's published stream catalog.
func myHoursStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "clients",
			Description: "My Hours clients.",
			PrimaryKey:  []string{"id"},
			Fields:      clientFields(),
		},
		{
			Name:        "projects",
			Description: "My Hours projects.",
			PrimaryKey:  []string{"id"},
			Fields:      projectFields(),
		},
		{
			Name:        "users",
			Description: "My Hours team members.",
			PrimaryKey:  []string{"id"},
			Fields:      userFields(),
		},
		{
			Name:        "tags",
			Description: "My Hours tags.",
			PrimaryKey:  []string{"id"},
			Fields:      tagFields(),
		},
		{
			Name:         "time_logs",
			Description:  "My Hours time log activity entries.",
			PrimaryKey:   []string{"logId"},
			CursorFields: []string{"date"},
			Fields:       timeLogFields(),
		},
	}
}

func clientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "date_archived", Type: "string"},
		{Name: "custom_id", Type: "string"},
	}
}

func projectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "client_id", Type: "integer"},
		{Name: "client_name", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "billable", Type: "boolean"},
		{Name: "date_created", Type: "string"},
		{Name: "date_archived", Type: "string"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "archived", Type: "boolean"},
		{Name: "account_owner", Type: "boolean"},
		{Name: "admin", Type: "boolean"},
		{Name: "rate", Type: "number"},
		{Name: "billable_rate", Type: "number"},
	}
}

func tagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "date_archived", Type: "string"},
	}
}

func timeLogFields() []connectors.Field {
	return []connectors.Field{
		{Name: "logId", Type: "integer"},
		{Name: "user_id", Type: "integer"},
		{Name: "user_name", Type: "string"},
		{Name: "date", Type: "string"},
		{Name: "client_id", Type: "integer"},
		{Name: "client_name", Type: "string"},
		{Name: "project_id", Type: "integer"},
		{Name: "project_name", Type: "string"},
		{Name: "task_id", Type: "integer"},
		{Name: "task_name", Type: "string"},
		{Name: "tags", Type: "string"},
		{Name: "note", Type: "string"},
		{Name: "billable", Type: "boolean"},
		{Name: "invoiced", Type: "boolean"},
		{Name: "rate", Type: "number"},
		{Name: "amount", Type: "number"},
		{Name: "billable_amount", Type: "number"},
		{Name: "log_duration", Type: "number"},
		{Name: "labor_hours", Type: "number"},
		{Name: "billable_hours", Type: "number"},
	}
}

func clientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"archived":      item["archived"],
		"date_archived": item["dateArchived"],
		"custom_id":     item["customId"],
	}
}

func projectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"client_id":     item["clientId"],
		"client_name":   item["clientName"],
		"archived":      item["archived"],
		"billable":      item["billable"],
		"date_created":  item["dateCreated"],
		"date_archived": item["dateArchived"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"email":         item["email"],
		"active":        item["active"],
		"archived":      item["archived"],
		"account_owner": item["accountOwner"],
		"admin":         item["admin"],
		"rate":          item["rate"],
		"billable_rate": item["billableRate"],
	}
}

func tagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"archived":      item["archived"],
		"date_archived": item["dateArchived"],
	}
}

func timeLogRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"logId":           item["logId"],
		"user_id":         item["userId"],
		"user_name":       item["userName"],
		"date":            item["date"],
		"client_id":       item["clientId"],
		"client_name":     item["clientName"],
		"project_id":      item["projectId"],
		"project_name":    item["projectName"],
		"task_id":         item["taskId"],
		"task_name":       item["taskName"],
		"tags":            item["tags"],
		"note":            item["note"],
		"billable":        item["billable"],
		"invoiced":        item["invoiced"],
		"rate":            item["rate"],
		"amount":          item["amount"],
		"billable_amount": item["billableAmount"],
		"log_duration":    item["logDuration"],
		"labor_hours":     item["laborHours"],
		"billable_hours":  item["billableHours"],
	}
}
