package salesloft

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Salesloft API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Salesloft list endpoint path segment (e.g. "people").
	resource string
	// mapRecord flattens a raw Salesloft object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// salesloftStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in salesloftStreams; the read
// path is fully data-driven from this table.
var salesloftStreamEndpoints = map[string]streamEndpoint{
	"people":   {resource: "people", mapRecord: peopleRecord},
	"accounts": {resource: "accounts", mapRecord: accountRecord},
	"cadences": {resource: "cadences", mapRecord: cadenceRecord},
	"users":    {resource: "users", mapRecord: userRecord},
	"emails":   {resource: "emails", mapRecord: emailRecord},
}

// salesloftStreams returns the connector's published stream catalog. Every
// Salesloft object exposes an integer id and most expose an updated_at RFC3339
// timestamp used as the incremental cursor.
func salesloftStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "people",
			Description:  "Salesloft people (contacts/prospects).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       peopleFields(),
		},
		{
			Name:         "accounts",
			Description:  "Salesloft accounts (companies).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       accountFields(),
		},
		{
			Name:         "cadences",
			Description:  "Salesloft cadences (sequences).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       cadenceFields(),
		},
		{
			Name:         "users",
			Description:  "Salesloft users (team members).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       userFields(),
		},
		{
			Name:         "emails",
			Description:  "Salesloft emails sent through cadences.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       emailFields(),
		},
	}
}

func peopleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "email_address", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "person_company_name", Type: "string"},
		{Name: "account_id", Type: "integer"},
		{Name: "owner_id", Type: "integer"},
		{Name: "do_not_contact", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func accountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "domain", Type: "string"},
		{Name: "website", Type: "string"},
		{Name: "industry", Type: "string"},
		{Name: "company_type", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "owner_id", Type: "integer"},
		{Name: "archived_at", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func cadenceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "team_cadence", Type: "boolean"},
		{Name: "shared", Type: "boolean"},
		{Name: "remove_bounces_enabled", Type: "boolean"},
		{Name: "remove_replies_enabled", Type: "boolean"},
		{Name: "archived_at", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "guid", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "time_zone", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func emailFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "subject", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "bounced", Type: "boolean"},
		{Name: "view_tracking", Type: "boolean"},
		{Name: "click_tracking", Type: "boolean"},
		{Name: "sent_at", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func peopleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"email_address":       item["email_address"],
		"first_name":          item["first_name"],
		"last_name":           item["last_name"],
		"display_name":        item["display_name"],
		"title":               item["title"],
		"phone":               item["phone"],
		"person_company_name": item["person_company_name"],
		"account_id":          relationID(item["account"]),
		"owner_id":            relationID(item["owner"]),
		"do_not_contact":      item["do_not_contact"],
		"created_at":          item["created_at"],
		"updated_at":          item["updated_at"],
	}
}

func accountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"domain":       item["domain"],
		"website":      item["website"],
		"industry":     item["industry"],
		"company_type": item["company_type"],
		"phone":        item["phone"],
		"city":         item["city"],
		"country":      item["country"],
		"owner_id":     relationID(item["owner"]),
		"archived_at":  item["archived_at"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func cadenceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                     item["id"],
		"name":                   item["name"],
		"team_cadence":           item["team_cadence"],
		"shared":                 item["shared"],
		"remove_bounces_enabled": item["remove_bounces_enabled"],
		"remove_replies_enabled": item["remove_replies_enabled"],
		"archived_at":            item["archived_at"],
		"created_at":             item["created_at"],
		"updated_at":             item["updated_at"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"guid":       item["guid"],
		"name":       item["name"],
		"first_name": item["first_name"],
		"last_name":  item["last_name"],
		"email":      item["email"],
		"active":     item["active"],
		"time_zone":  item["time_zone"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func emailRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"subject":        item["subject"],
		"status":         item["status"],
		"bounced":        item["bounced"],
		"view_tracking":  item["view_tracking"],
		"click_tracking": item["click_tracking"],
		"sent_at":        item["sent_at"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
	}
}

// relationID extracts the "id" from an embedded Salesloft relationship object
// such as {"id":123,"_href":"..."}. Salesloft nests foreign keys this way; this
// flattens them to a scalar id. Returns nil when absent.
func relationID(v any) any {
	obj, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	return obj["id"]
}
