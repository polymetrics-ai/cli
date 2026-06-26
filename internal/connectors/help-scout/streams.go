package helpscout

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Help Scout Mailbox API resource path
// (relative to base_url), the HAL+JSON _embedded key the records array lives
// under, and the record mapper that flattens the resource objects.
type streamEndpoint struct {
	// resource is the list endpoint path segment (e.g. "conversations").
	resource string
	// embeddedKey is the key under _embedded that holds the records array.
	embeddedKey string
	// mapRecord flattens a raw Help Scout object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// helpScoutStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in helpScoutStreams; the read
// path is fully data-driven from this table.
var helpScoutStreamEndpoints = map[string]streamEndpoint{
	"conversations": {resource: "conversations", embeddedKey: "conversations", mapRecord: conversationRecord},
	"customers":     {resource: "customers", embeddedKey: "customers", mapRecord: customerRecord},
	"mailboxes":     {resource: "mailboxes", embeddedKey: "mailboxes", mapRecord: mailboxRecord},
	"users":         {resource: "users", embeddedKey: "users", mapRecord: userRecord},
}

// helpScoutStreams returns the connector's published stream catalog. Every Help
// Scout object exposes a numeric id; objects that change over time carry an
// updatedAt timestamp used as the incremental cursor.
func helpScoutStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "conversations",
			Description:  "Help Scout conversations (support tickets) across mailboxes.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"userUpdatedAt"},
			Fields:       conversationFields(),
		},
		{
			Name:         "customers",
			Description:  "Help Scout customers (end users who contact support).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       customerFields(),
		},
		{
			Name:         "mailboxes",
			Description:  "Help Scout mailboxes (inboxes that receive conversations).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       mailboxFields(),
		},
		{
			Name:         "users",
			Description:  "Help Scout users (agents and admins on the account).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       userFields(),
		},
	}
}

func conversationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "number", Type: "integer"},
		{Name: "type", Type: "string"},
		{Name: "folderId", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "mailboxId", Type: "integer"},
		{Name: "assigneeId", Type: "integer"},
		{Name: "preview", Type: "string"},
		{Name: "threads", Type: "integer"},
		{Name: "createdAt", Type: "string"},
		{Name: "closedAt", Type: "string"},
		{Name: "userUpdatedAt", Type: "string"},
	}
}

func customerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "gender", Type: "string"},
		{Name: "jobTitle", Type: "string"},
		{Name: "organization", Type: "string"},
		{Name: "photoUrl", Type: "string"},
		{Name: "age", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func mailboxFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "jobTitle", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func conversationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"number":        item["number"],
		"type":          item["type"],
		"folderId":      item["folderId"],
		"status":        item["status"],
		"state":         item["state"],
		"subject":       item["subject"],
		"mailboxId":     item["mailboxId"],
		"assigneeId":    item["assigneeId"],
		"preview":       item["preview"],
		"threads":       item["threads"],
		"createdAt":     item["createdAt"],
		"closedAt":      item["closedAt"],
		"userUpdatedAt": item["userUpdatedAt"],
	}
}

func customerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"firstName":    item["firstName"],
		"lastName":     item["lastName"],
		"gender":       item["gender"],
		"jobTitle":     item["jobTitle"],
		"organization": item["organization"],
		"photoUrl":     item["photoUrl"],
		"age":          item["age"],
		"createdAt":    item["createdAt"],
		"updatedAt":    item["updatedAt"],
	}
}

func mailboxRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"name":      item["name"],
		"slug":      item["slug"],
		"email":     item["email"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"firstName": item["firstName"],
		"lastName":  item["lastName"],
		"email":     item["email"],
		"role":      item["role"],
		"timezone":  item["timezone"],
		"type":      item["type"],
		"jobTitle":  item["jobTitle"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
	}
}
