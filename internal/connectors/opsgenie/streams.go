package opsgenie

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Opsgenie API resource path and the
// mapper that flattens raw API objects into connector records.
type streamEndpoint struct {
	resource  string
	mapRecord func(map[string]any) connectors.Record
}

var opsgenieStreamEndpoints = map[string]streamEndpoint{
	"alerts":    {resource: "alerts", mapRecord: opsgenieAlertRecord},
	"incidents": {resource: "incidents", mapRecord: opsgenieIncidentRecord},
	"users":     {resource: "users", mapRecord: opsgenieUserRecord},
	"teams":     {resource: "teams", mapRecord: opsgenieTeamRecord},
	"services":  {resource: "services", mapRecord: opsgenieServiceRecord},
}

func opsgenieStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "alerts",
			Description:  "Opsgenie alerts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       opsgenieAlertFields(),
		},
		{
			Name:         "incidents",
			Description:  "Opsgenie incidents.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       opsgenieIncidentFields(),
		},
		{
			Name:        "users",
			Description: "Opsgenie users.",
			PrimaryKey:  []string{"id"},
			Fields:      opsgenieUserFields(),
		},
		{
			Name:        "teams",
			Description: "Opsgenie teams.",
			PrimaryKey:  []string{"id"},
			Fields:      opsgenieTeamFields(),
		},
		{
			Name:        "services",
			Description: "Opsgenie services.",
			PrimaryKey:  []string{"id"},
			Fields:      opsgenieServiceFields(),
		},
	}
}

func opsgenieAlertFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "tiny_id", Type: "string"},
		{Name: "alias", Type: "string"},
		{Name: "message", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "priority", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "last_occurred_at", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "owner", Type: "string"},
		{Name: "tags", Type: "array"},
		{Name: "responders", Type: "array"},
		{Name: "details", Type: "object"},
	}
}

func opsgenieIncidentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "tiny_id", Type: "string"},
		{Name: "message", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "priority", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "owner_team", Type: "object"},
		{Name: "impacted_services", Type: "array"},
		{Name: "tags", Type: "array"},
		{Name: "responders", Type: "array"},
	}
}

func opsgenieUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "username", Type: "string"},
		{Name: "full_name", Type: "string"},
		{Name: "role", Type: "object"},
		{Name: "time_zone", Type: "string"},
		{Name: "locale", Type: "string"},
		{Name: "blocked", Type: "boolean"},
		{Name: "verified", Type: "boolean"},
	}
}

func opsgenieTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "members", Type: "array"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func opsgenieServiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "team_id", Type: "string"},
		{Name: "visibility", Type: "string"},
		{Name: "tags", Type: "array"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func opsgenieAlertRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"tiny_id":          item["tinyId"],
		"alias":            item["alias"],
		"message":          item["message"],
		"status":           item["status"],
		"priority":         item["priority"],
		"created_at":       item["createdAt"],
		"updated_at":       item["updatedAt"],
		"last_occurred_at": item["lastOccurredAt"],
		"source":           item["source"],
		"owner":            item["owner"],
		"tags":             item["tags"],
		"responders":       item["responders"],
		"details":          item["details"],
	}
}

func opsgenieIncidentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"tiny_id":           item["tinyId"],
		"message":           item["message"],
		"description":       item["description"],
		"status":            item["status"],
		"priority":          item["priority"],
		"created_at":        item["createdAt"],
		"updated_at":        item["updatedAt"],
		"owner_team":        item["ownerTeam"],
		"impacted_services": item["impactedServices"],
		"tags":              item["tags"],
		"responders":        item["responders"],
	}
}

func opsgenieUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"username":  item["username"],
		"full_name": item["fullName"],
		"role":      item["role"],
		"time_zone": item["timeZone"],
		"locale":    item["locale"],
		"blocked":   item["blocked"],
		"verified":  item["verified"],
	}
}

func opsgenieTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"members":     item["members"],
		"created_at":  item["createdAt"],
		"updated_at":  item["updatedAt"],
	}
}

func opsgenieServiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"team_id":     item["teamId"],
		"visibility":  item["visibility"],
		"tags":        item["tags"],
		"created_at":  item["createdAt"],
		"updated_at":  item["updatedAt"],
	}
}
