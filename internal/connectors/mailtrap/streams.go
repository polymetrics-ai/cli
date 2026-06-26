package mailtrap

import "polymetrics.ai/internal/connectors"

// scope describes how a stream's endpoint is addressed.
type scope int

const (
	// scopeRoot is a top-level resource (no account_id needed), e.g. accounts.
	scopeRoot scope = iota
	// scopeAccount is an account-scoped resource under accounts/{account_id}/...,
	// e.g. inboxes, projects, sending_domains.
	scopeAccount
)

// streamEndpoint maps a stream name to the Mailtrap API resource it reads from,
// the scope (whether it requires an account_id), the JSON path the records live
// at, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the path segment under the account scope (e.g. "inboxes") or
	// the top-level path for root-scoped streams (e.g. "accounts").
	resource string
	scope    scope
	// recordsPath is the dotted path to the records array in the response body.
	// "" selects the root (most Mailtrap list endpoints return a bare array);
	// "data" selects the {"data":[...]} envelope used by sending_domains.
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

// mailtrapStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in mailtrapStreams; the read
// path is fully data-driven from this table.
var mailtrapStreamEndpoints = map[string]streamEndpoint{
	"accounts":        {resource: "accounts", scope: scopeRoot, recordsPath: "", mapRecord: accountRecord},
	"inboxes":         {resource: "inboxes", scope: scopeAccount, recordsPath: "", mapRecord: inboxRecord},
	"projects":        {resource: "projects", scope: scopeAccount, recordsPath: "", mapRecord: projectRecord},
	"sending_domains": {resource: "sending_domains", scope: scopeAccount, recordsPath: "data", mapRecord: sendingDomainRecord},
}

// mailtrapStreams returns the connector's published stream catalog. Mailtrap is
// full-refresh only (no incremental cursors), so CursorFields is empty.
func mailtrapStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "accounts",
			Description: "Mailtrap accounts the API token can access.",
			PrimaryKey:  []string{"id"},
			Fields:      accountFields(),
		},
		{
			Name:        "inboxes",
			Description: "Email-testing inboxes within an account.",
			PrimaryKey:  []string{"id"},
			Fields:      inboxFields(),
		},
		{
			Name:        "projects",
			Description: "Projects (inbox groupings) within an account.",
			PrimaryKey:  []string{"id"},
			Fields:      projectFields(),
		},
		{
			Name:        "sending_domains",
			Description: "Sending domains configured for an account.",
			PrimaryKey:  []string{"id"},
			Fields:      sendingDomainFields(),
		},
	}
}

func accountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "access_levels", Type: "array"},
	}
}

func inboxFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "account_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "domain", Type: "string"},
		{Name: "email_username", Type: "string"},
		{Name: "emails_count", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "max_size", Type: "integer"},
		{Name: "used_size", Type: "integer"},
	}
}

func projectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "account_id", Type: "string"},
		{Name: "name", Type: "string"},
	}
}

func sendingDomainFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "account_id", Type: "string"},
		{Name: "domain_name", Type: "string"},
		{Name: "demo", Type: "boolean"},
		{Name: "status", Type: "string"},
	}
}

func accountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"access_levels": item["access_levels"],
	}
}

func inboxRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"domain":         item["domain"],
		"email_username": item["email_username"],
		"emails_count":   item["emails_count"],
		"status":         item["status"],
		"max_size":       item["max_size"],
		"used_size":      item["used_size"],
	}
}

func projectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   item["id"],
		"name": item["name"],
	}
}

func sendingDomainRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"domain_name": item["domain_name"],
		"demo":        item["demo"],
		"status":      item["status"],
	}
}
