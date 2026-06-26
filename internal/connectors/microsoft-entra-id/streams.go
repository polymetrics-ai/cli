package microsoftentraid

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Microsoft Graph resource path
// (relative to base_url) it reads from, and the record mapper that flattens its
// objects into a connectors.Record.
type streamEndpoint struct {
	// resource is the Graph collection path segment (e.g. "users").
	resource string
	// mapRecord flattens a raw Graph directory object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streams(); the read path is fully
// data-driven from this table.
var streamEndpoints = map[string]streamEndpoint{
	"users":             {resource: "users", mapRecord: userRecord},
	"groups":            {resource: "groups", mapRecord: groupRecord},
	"applications":      {resource: "applications", mapRecord: applicationRecord},
	"serviceprincipals": {resource: "servicePrincipals", mapRecord: servicePrincipalRecord},
	"directoryroles":    {resource: "directoryRoles", mapRecord: directoryRoleRecord},
}

// streams returns the connector's published stream catalog. Every Microsoft
// Graph directory object exposes a string "id", so the primary key is ["id"]
// across the board. Graph directory collections are full-refresh only, so there
// are no incremental cursor fields.
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "users",
			Description: "Microsoft Entra ID users (directory members).",
			PrimaryKey:  []string{"id"},
			Fields:      userFields(),
		},
		{
			Name:        "groups",
			Description: "Microsoft Entra ID groups.",
			PrimaryKey:  []string{"id"},
			Fields:      groupFields(),
		},
		{
			Name:        "applications",
			Description: "Microsoft Entra ID application registrations.",
			PrimaryKey:  []string{"id"},
			Fields:      applicationFields(),
		},
		{
			Name:        "serviceprincipals",
			Description: "Microsoft Entra ID service principals.",
			PrimaryKey:  []string{"id"},
			Fields:      servicePrincipalFields(),
		},
		{
			Name:        "directoryroles",
			Description: "Microsoft Entra ID activated directory roles.",
			PrimaryKey:  []string{"id"},
			Fields:      directoryRoleFields(),
		},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "user_principal_name", Type: "string"},
		{Name: "given_name", Type: "string"},
		{Name: "surname", Type: "string"},
		{Name: "mail", Type: "string"},
		{Name: "job_title", Type: "string"},
		{Name: "department", Type: "string"},
		{Name: "office_location", Type: "string"},
		{Name: "mobile_phone", Type: "string"},
		{Name: "account_enabled", Type: "boolean"},
	}
}

func groupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "mail", Type: "string"},
		{Name: "mail_nickname", Type: "string"},
		{Name: "mail_enabled", Type: "boolean"},
		{Name: "security_enabled", Type: "boolean"},
		{Name: "visibility", Type: "string"},
		{Name: "created_date_time", Type: "string"},
	}
}

func applicationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "app_id", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "sign_in_audience", Type: "string"},
		{Name: "publisher_domain", Type: "string"},
		{Name: "created_date_time", Type: "string"},
	}
}

func servicePrincipalFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "app_id", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "service_principal_type", Type: "string"},
		{Name: "account_enabled", Type: "boolean"},
		{Name: "app_owner_organization_id", Type: "string"},
		{Name: "sign_in_audience", Type: "string"},
	}
}

func directoryRoleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "role_template_id", Type: "string"},
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"display_name":        item["displayName"],
		"user_principal_name": item["userPrincipalName"],
		"given_name":          item["givenName"],
		"surname":             item["surname"],
		"mail":                item["mail"],
		"job_title":           item["jobTitle"],
		"department":          item["department"],
		"office_location":     item["officeLocation"],
		"mobile_phone":        item["mobilePhone"],
		"account_enabled":     item["accountEnabled"],
	}
}

func groupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"display_name":      item["displayName"],
		"description":       item["description"],
		"mail":              item["mail"],
		"mail_nickname":     item["mailNickname"],
		"mail_enabled":      item["mailEnabled"],
		"security_enabled":  item["securityEnabled"],
		"visibility":        item["visibility"],
		"created_date_time": item["createdDateTime"],
	}
}

func applicationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"app_id":            item["appId"],
		"display_name":      item["displayName"],
		"description":       item["description"],
		"sign_in_audience":  item["signInAudience"],
		"publisher_domain":  item["publisherDomain"],
		"created_date_time": item["createdDateTime"],
	}
}

func servicePrincipalRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                        item["id"],
		"app_id":                    item["appId"],
		"display_name":              item["displayName"],
		"service_principal_type":    item["servicePrincipalType"],
		"account_enabled":           item["accountEnabled"],
		"app_owner_organization_id": item["appOwnerOrganizationId"],
		"sign_in_audience":          item["signInAudience"],
	}
}

func directoryRoleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"display_name":     item["displayName"],
		"description":      item["description"],
		"role_template_id": item["roleTemplateId"],
	}
}
