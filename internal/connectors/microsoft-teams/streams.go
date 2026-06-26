package microsoftteams

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Microsoft Graph resource path
// (relative to base_url) it reads from, plus the record mapper that flattens its
// objects. The read path is fully data-driven from this table.
type streamEndpoint struct {
	// resource is the Graph collection path segment (e.g. "users").
	resource string
	// mapRecord flattens a raw Graph object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// graphStreamEndpoints is the per-stream routing table. Every entry here is a
// top-level Microsoft Graph collection that returns {value:[...],
// "@odata.nextLink":...}. Adding a stream means adding one entry here plus a
// Stream definition in graphStreams.
var graphStreamEndpoints = map[string]streamEndpoint{
	"users":                    {resource: "users", mapRecord: graphUserRecord},
	"groups":                   {resource: "groups", mapRecord: graphGroupRecord},
	"channels":                 {resource: "teams/getAllChannels", mapRecord: graphChannelRecord},
	"team_device_usage_report": {resource: "reports/getTeamsDeviceUsageUserDetail", mapRecord: graphDeviceUsageRecord},
}

// graphStreams returns the connector's published stream catalog. Graph objects
// expose a string id; collections that carry a change marker use it as the
// incremental cursor where one exists.
func graphStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "users",
			Description:  "Microsoft 365 / Entra ID users in the tenant.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       graphUserFields(),
		},
		{
			Name:         "groups",
			Description:  "Microsoft 365 groups (including Teams-backed groups).",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       graphGroupFields(),
		},
		{
			Name:         "channels",
			Description:  "Channels across all teams in the tenant.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       graphChannelFields(),
		},
		{
			Name:         "team_device_usage_report",
			Description:  "Teams device usage per user over the configured aggregation period.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       graphDeviceUsageFields(),
		},
	}
}

func graphUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "user_principal_name", Type: "string"},
		{Name: "mail", Type: "string"},
		{Name: "job_title", Type: "string"},
		{Name: "office_location", Type: "string"},
		{Name: "mobile_phone", Type: "string"},
		{Name: "account_enabled", Type: "boolean"},
	}
}

func graphGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "mail", Type: "string"},
		{Name: "mail_nickname", Type: "string"},
		{Name: "visibility", Type: "string"},
		{Name: "created_date_time", Type: "string"},
		{Name: "security_enabled", Type: "boolean"},
		{Name: "mail_enabled", Type: "boolean"},
	}
}

func graphChannelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "display_name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "membership_type", Type: "string"},
		{Name: "web_url", Type: "string"},
		{Name: "created_date_time", Type: "string"},
	}
}

func graphDeviceUsageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "user_principal_name", Type: "string"},
		{Name: "last_activity_date", Type: "string"},
		{Name: "is_deleted", Type: "boolean"},
		{Name: "used_web", Type: "boolean"},
		{Name: "used_windows_phone", Type: "boolean"},
		{Name: "used_android_phone", Type: "boolean"},
		{Name: "used_i_os", Type: "boolean"},
		{Name: "used_mac", Type: "boolean"},
		{Name: "report_period", Type: "string"},
	}
}

func graphUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"display_name":        item["displayName"],
		"user_principal_name": item["userPrincipalName"],
		"mail":                item["mail"],
		"job_title":           item["jobTitle"],
		"office_location":     item["officeLocation"],
		"mobile_phone":        item["mobilePhone"],
		"account_enabled":     item["accountEnabled"],
	}
}

func graphGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"display_name":      item["displayName"],
		"description":       item["description"],
		"mail":              item["mail"],
		"mail_nickname":     item["mailNickname"],
		"visibility":        item["visibility"],
		"created_date_time": item["createdDateTime"],
		"security_enabled":  item["securityEnabled"],
		"mail_enabled":      item["mailEnabled"],
	}
}

func graphChannelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"display_name":      item["displayName"],
		"description":       item["description"],
		"email":             item["email"],
		"membership_type":   item["membershipType"],
		"web_url":           item["webUrl"],
		"created_date_time": item["createdDateTime"],
	}
}

func graphDeviceUsageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"user_principal_name": item["userPrincipalName"],
		"last_activity_date":  item["lastActivityDate"],
		"is_deleted":          item["isDeleted"],
		"used_web":            item["usedWeb"],
		"used_windows_phone":  item["usedWindowsPhone"],
		"used_android_phone":  item["usedAndroidPhone"],
		"used_i_os":           item["usedIOS"],
		"used_mac":            item["usedMac"],
		"report_period":       item["reportPeriod"],
	}
}
