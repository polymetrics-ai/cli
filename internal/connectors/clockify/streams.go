package clockify

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Clockify API resource path and the
// record mapper that flattens its objects. workspaceScoped marks endpoints that
// live under /v1/workspaces/{workspaceId}/... versus the top-level
// /v1/workspaces list.
type streamEndpoint struct {
	// resource is the path segment appended after the workspace prefix
	// (e.g. "clients"). Empty for the top-level workspaces stream.
	resource string
	// workspaceScoped is true when the path is
	// /v1/workspaces/{workspaceId}/<resource>; false for the bare
	// /v1/workspaces list.
	workspaceScoped bool
	// mapRecord flattens a raw Clockify object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// clockifyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in clockifyStreams; the read
// path is fully data-driven from this table.
var clockifyStreamEndpoints = map[string]streamEndpoint{
	"workspaces": {resource: "", workspaceScoped: false, mapRecord: clockifyWorkspaceRecord},
	"clients":    {resource: "clients", workspaceScoped: true, mapRecord: clockifyClientRecord},
	"projects":   {resource: "projects", workspaceScoped: true, mapRecord: clockifyProjectRecord},
	"tags":       {resource: "tags", workspaceScoped: true, mapRecord: clockifyTagRecord},
	"users":      {resource: "users", workspaceScoped: true, mapRecord: clockifyUserRecord},
}

// clockifyStreams returns the connector's published stream catalog. Clockify
// objects are identified by a string "id"; Clockify list endpoints do not expose
// an updated-at cursor field, so these streams are full-refresh (no cursor).
func clockifyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "workspaces",
			Description: "Clockify workspaces the API key can access.",
			PrimaryKey:  []string{"id"},
			Fields:      clockifyWorkspaceFields(),
		},
		{
			Name:        "clients",
			Description: "Clients within the configured workspace.",
			PrimaryKey:  []string{"id"},
			Fields:      clockifyClientFields(),
		},
		{
			Name:        "projects",
			Description: "Projects within the configured workspace.",
			PrimaryKey:  []string{"id"},
			Fields:      clockifyProjectFields(),
		},
		{
			Name:        "tags",
			Description: "Tags within the configured workspace.",
			PrimaryKey:  []string{"id"},
			Fields:      clockifyTagFields(),
		},
		{
			Name:        "users",
			Description: "Users within the configured workspace.",
			PrimaryKey:  []string{"id"},
			Fields:      clockifyUserFields(),
		},
	}
}

func clockifyWorkspaceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "hourlyRate", Type: "object"},
		{Name: "memberships", Type: "array"},
		{Name: "workspaceSettings", Type: "object"},
		{Name: "imageUrl", Type: "string"},
		{Name: "featureSubscriptionType", Type: "string"},
	}
}

func clockifyClientFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "workspaceId", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "address", Type: "string"},
		{Name: "note", Type: "string"},
		{Name: "archived", Type: "boolean"},
	}
}

func clockifyProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "workspaceId", Type: "string"},
		{Name: "clientId", Type: "string"},
		{Name: "clientName", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "billable", Type: "boolean"},
		{Name: "public", Type: "boolean"},
		{Name: "archived", Type: "boolean"},
		{Name: "duration", Type: "string"},
		{Name: "note", Type: "string"},
	}
}

func clockifyTagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "workspaceId", Type: "string"},
		{Name: "archived", Type: "boolean"},
	}
}

func clockifyUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "activeWorkspace", Type: "string"},
		{Name: "defaultWorkspace", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "profilePicture", Type: "string"},
	}
}

func clockifyWorkspaceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                      item["id"],
		"name":                    item["name"],
		"hourlyRate":              item["hourlyRate"],
		"memberships":             item["memberships"],
		"workspaceSettings":       item["workspaceSettings"],
		"imageUrl":                item["imageUrl"],
		"featureSubscriptionType": item["featureSubscriptionType"],
	}
}

func clockifyClientRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"workspaceId": item["workspaceId"],
		"email":       item["email"],
		"address":     item["address"],
		"note":        item["note"],
		"archived":    item["archived"],
	}
}

func clockifyProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"workspaceId": item["workspaceId"],
		"clientId":    item["clientId"],
		"clientName":  item["clientName"],
		"color":       item["color"],
		"billable":    item["billable"],
		"public":      item["public"],
		"archived":    item["archived"],
		"duration":    item["duration"],
		"note":        item["note"],
	}
}

func clockifyTagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"workspaceId": item["workspaceId"],
		"archived":    item["archived"],
	}
}

func clockifyUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"email":            item["email"],
		"activeWorkspace":  item["activeWorkspace"],
		"defaultWorkspace": item["defaultWorkspace"],
		"status":           item["status"],
		"profilePicture":   item["profilePicture"],
	}
}
