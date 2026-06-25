package auth0

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Auth0 Management API v2 resource path
// (relative to base_url) it reads from, the JSON key under which the
// include_totals envelope returns the array, and the record mapper.
type streamEndpoint struct {
	// resource is the Management API path segment under /api/v2 (e.g. "users").
	resource string
	// arrayKey is the key holding the records array when include_totals=true.
	// Auth0 names it after the resource (users, clients, connections, roles,
	// organizations). Empty means the endpoint returns a bare top-level array.
	arrayKey string
	// mapRecord flattens a raw Auth0 object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// auth0StreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in auth0Streams; the read path
// is fully data-driven from this table.
var auth0StreamEndpoints = map[string]streamEndpoint{
	"users":         {resource: "users", arrayKey: "users", mapRecord: auth0UserRecord},
	"clients":       {resource: "clients", arrayKey: "clients", mapRecord: auth0ClientRecord},
	"connections":   {resource: "connections", arrayKey: "connections", mapRecord: auth0ConnectionRecord},
	"roles":         {resource: "roles", arrayKey: "roles", mapRecord: auth0RoleRecord},
	"organizations": {resource: "organizations", arrayKey: "organizations", mapRecord: auth0OrganizationRecord},
}

// auth0Streams returns the connector's published stream catalog. Auth0 objects
// use resource-specific string ids (user_id, client_id, id) for the primary key;
// users additionally expose updated_at for incremental cursoring.
func auth0Streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "users",
			Description:  "Auth0 user accounts from the Management API.",
			PrimaryKey:   []string{"user_id"},
			CursorFields: []string{"updated_at"},
			Fields:       auth0UserFields(),
		},
		{
			Name:        "clients",
			Description: "Auth0 applications (clients) registered in the tenant.",
			PrimaryKey:  []string{"client_id"},
			Fields:      auth0ClientFields(),
		},
		{
			Name:        "connections",
			Description: "Auth0 identity connections configured in the tenant.",
			PrimaryKey:  []string{"id"},
			Fields:      auth0ConnectionFields(),
		},
		{
			Name:        "roles",
			Description: "Auth0 RBAC roles defined in the tenant.",
			PrimaryKey:  []string{"id"},
			Fields:      auth0RoleFields(),
		},
		{
			Name:        "organizations",
			Description: "Auth0 organizations in the tenant.",
			PrimaryKey:  []string{"id"},
			Fields:      auth0OrganizationFields(),
		},
	}
}

func auth0UserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "user_id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "email_verified", Type: "boolean"},
		{Name: "username", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "nickname", Type: "string"},
		{Name: "given_name", Type: "string"},
		{Name: "family_name", Type: "string"},
		{Name: "picture", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
		{Name: "last_login", Type: "timestamp"},
		{Name: "logins_count", Type: "integer"},
		{Name: "blocked", Type: "boolean"},
	}
}

func auth0ClientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "client_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "app_type", Type: "string"},
		{Name: "is_first_party", Type: "boolean"},
		{Name: "oidc_conformant", Type: "boolean"},
		{Name: "global", Type: "boolean"},
	}
}

func auth0ConnectionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "strategy", Type: "string"},
		{Name: "is_domain_connection", Type: "boolean"},
	}
}

func auth0RoleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
	}
}

func auth0OrganizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "display_name", Type: "string"},
	}
}

func auth0UserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"user_id":        item["user_id"],
		"email":          item["email"],
		"email_verified": item["email_verified"],
		"username":       item["username"],
		"name":           item["name"],
		"nickname":       item["nickname"],
		"given_name":     item["given_name"],
		"family_name":    item["family_name"],
		"picture":        item["picture"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
		"last_login":     item["last_login"],
		"logins_count":   item["logins_count"],
		"blocked":        item["blocked"],
	}
}

func auth0ClientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"client_id":       item["client_id"],
		"name":            item["name"],
		"description":     item["description"],
		"app_type":        item["app_type"],
		"is_first_party":  item["is_first_party"],
		"oidc_conformant": item["oidc_conformant"],
		"global":          item["global"],
	}
}

func auth0ConnectionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                   item["id"],
		"name":                 item["name"],
		"display_name":         item["display_name"],
		"strategy":             item["strategy"],
		"is_domain_connection": item["is_domain_connection"],
	}
}

func auth0RoleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
	}
}

func auth0OrganizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"display_name": item["display_name"],
	}
}
