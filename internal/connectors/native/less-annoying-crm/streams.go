package lessannoyingcrm

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Less Annoying CRM RPC Function it
// reads from, the JSON path to its records, and the mapper that flattens its
// objects into a connectors.Record.
type streamEndpoint struct {
	// function is the Less Annoying CRM v2 Function name (e.g. "GetContacts").
	function string
	// recordsPath is the dotted path to the records array. Empty means the
	// package default ("Results"); GetUsers returns a root array, so it uses
	// "" with a recordsPath of "." resolved at the call site -> here we set it
	// explicitly per stream.
	recordsPath string
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// lacrmStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in lacrmStreams; the read path
// is fully data-driven from this table.
//
// GetUsers returns a top-level array (record path "."); the list Functions
// return their rows under "Results".
var lacrmStreamEndpoints = map[string]streamEndpoint{
	"users":    {function: "GetUsers", recordsPath: ".", mapRecord: lacrmUserRecord},
	"contacts": {function: "GetContacts", recordsPath: lacrmRecordsPath, mapRecord: lacrmContactRecord},
	"tasks":    {function: "GetTasks", recordsPath: lacrmRecordsPath, mapRecord: lacrmTaskRecord},
	"notes":    {function: "GetNotes", recordsPath: lacrmRecordsPath, mapRecord: lacrmNoteRecord},
	"events":   {function: "GetEvents", recordsPath: lacrmRecordsPath, mapRecord: lacrmEventRecord},
}

// lacrmStreams returns the connector's published stream catalog. tasks and
// events support incremental sync upstream (DateCreated / DateUpdated cursors);
// the others are full-refresh.
func lacrmStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "users",
			Description: "Less Annoying CRM users on the account.",
			PrimaryKey:  []string{"UserId"},
			Fields:      lacrmUserFields(),
		},
		{
			Name:        "contacts",
			Description: "Less Annoying CRM contacts and companies.",
			PrimaryKey:  []string{"ContactId"},
			Fields:      lacrmContactFields(),
		},
		{
			Name:         "tasks",
			Description:  "Less Annoying CRM tasks.",
			PrimaryKey:   []string{"TaskId"},
			CursorFields: []string{"DateCreated"},
			Fields:       lacrmTaskFields(),
		},
		{
			Name:        "notes",
			Description: "Less Annoying CRM notes attached to contacts.",
			PrimaryKey:  []string{"NoteId"},
			Fields:      lacrmNoteFields(),
		},
		{
			Name:         "events",
			Description:  "Less Annoying CRM calendar events.",
			PrimaryKey:   []string{"EventId"},
			CursorFields: []string{"DateUpdated"},
			Fields:       lacrmEventFields(),
		},
	}
}

func lacrmUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "UserId", Type: "string"},
		{Name: "FirstName", Type: "string"},
		{Name: "LastName", Type: "string"},
		{Name: "Timezone", Type: "string"},
	}
}

func lacrmContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "ContactId", Type: "string"},
		{Name: "CompanyId", Type: "string"},
		{Name: "Name", Type: "string"},
		{Name: "Company Name", Type: "string"},
		{Name: "Job Title", Type: "string"},
		{Name: "Email", Type: "array"},
		{Name: "Phone", Type: "array"},
		{Name: "Address", Type: "array"},
		{Name: "Website", Type: "string"},
		{Name: "AssignedTo", Type: "number"},
		{Name: "IsCompany", Type: "boolean"},
		{Name: "DateCreated", Type: "timestamp"},
		{Name: "LastUpdate", Type: "timestamp"},
	}
}

func lacrmTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "TaskId", Type: "string"},
		{Name: "ContactId", Type: "string"},
		{Name: "Name", Type: "string"},
		{Name: "Description", Type: "string"},
		{Name: "DueDate", Type: "string"},
		{Name: "AssignedTo", Type: "string"},
		{Name: "CalendarId", Type: "string"},
		{Name: "IsCompleted", Type: "boolean"},
		{Name: "DateCreated", Type: "timestamp"},
	}
}

func lacrmNoteFields() []connectors.Field {
	return []connectors.Field{
		{Name: "NoteId", Type: "string"},
		{Name: "ContactId", Type: "string"},
		{Name: "UserId", Type: "string"},
		{Name: "Note", Type: "string"},
		{Name: "IsRichText", Type: "boolean"},
		{Name: "DateCreated", Type: "timestamp"},
		{Name: "DateDisplayedInHistory", Type: "timestamp"},
	}
}

func lacrmEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "EventId", Type: "string"},
		{Name: "Name", Type: "string"},
		{Name: "Description", Type: "string"},
		{Name: "Location", Type: "string"},
		{Name: "StartDate", Type: "timestamp"},
		{Name: "EndDate", Type: "timestamp"},
		{Name: "IsAllDay", Type: "boolean"},
		{Name: "IsRecurring", Type: "boolean"},
		{Name: "ContactIds", Type: "array"},
		{Name: "UserIds", Type: "array"},
		{Name: "DateCreated", Type: "timestamp"},
		{Name: "DateUpdated", Type: "timestamp"},
	}
}

func lacrmUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"UserId":    item["UserId"],
		"FirstName": item["FirstName"],
		"LastName":  item["LastName"],
		"Timezone":  item["Timezone"],
	}
}

func lacrmContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ContactId":    item["ContactId"],
		"CompanyId":    item["CompanyId"],
		"Name":         item["Name"],
		"Company Name": item["Company Name"],
		"Job Title":    item["Job Title"],
		"Email":        item["Email"],
		"Phone":        item["Phone"],
		"Address":      item["Address"],
		"Website":      item["Website"],
		"AssignedTo":   item["AssignedTo"],
		"IsCompany":    item["IsCompany"],
		"DateCreated":  item["DateCreated"],
		"LastUpdate":   item["LastUpdate"],
	}
}

func lacrmTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"TaskId":      item["TaskId"],
		"ContactId":   item["ContactId"],
		"Name":        item["Name"],
		"Description": item["Description"],
		"DueDate":     item["DueDate"],
		"AssignedTo":  item["AssignedTo"],
		"CalendarId":  item["CalendarId"],
		"IsCompleted": item["IsCompleted"],
		"DateCreated": item["DateCreated"],
	}
}

func lacrmNoteRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"NoteId":                 item["NoteId"],
		"ContactId":              item["ContactId"],
		"UserId":                 item["UserId"],
		"Note":                   item["Note"],
		"IsRichText":             item["IsRichText"],
		"DateCreated":            item["DateCreated"],
		"DateDisplayedInHistory": item["DateDisplayedInHistory"],
	}
}

func lacrmEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"EventId":     item["EventId"],
		"Name":        item["Name"],
		"Description": item["Description"],
		"Location":    item["Location"],
		"StartDate":   item["StartDate"],
		"EndDate":     item["EndDate"],
		"IsAllDay":    item["IsAllDay"],
		"IsRecurring": item["IsRecurring"],
		"ContactIds":  item["ContactIds"],
		"UserIds":     item["UserIds"],
		"DateCreated": item["DateCreated"],
		"DateUpdated": item["DateUpdated"],
	}
}
