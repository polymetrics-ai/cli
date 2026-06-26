package onesignal

import "polymetrics.ai/internal/connectors"

// authScope selects which secret authenticates a stream's requests. The OneSignal
// API splits credentials: account-level endpoints (apps) use the organization /
// user auth key, while app-scoped endpoints (players, notifications, outcomes)
// use the per-app REST API key.
type authScope int

const (
	// scopeApp uses the per-app REST API key (app_api_key).
	scopeApp authScope = iota
	// scopeUser uses the account-level user auth key (user_auth_key).
	scopeUser
)

// streamEndpoint maps a stream name to its OneSignal API endpoint and shape.
type streamEndpoint struct {
	// resource is the path relative to base_url. It may contain "{app_id}".
	resource string
	// recordsPath is the dotted JSON path to the records array. "" means the
	// response body is itself a top-level array.
	recordsPath string
	// scope selects which credential authenticates the request.
	scope authScope
	// needsAppID is true when the request must carry an app_id query param.
	needsAppID bool
	// paginated is true when the endpoint supports offset/limit pagination with
	// a total_count in the body.
	paginated bool
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// onesignalStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in onesignalStreams; the read
// path is fully data-driven from this table.
var onesignalStreamEndpoints = map[string]streamEndpoint{
	"apps": {
		resource:    "apps",
		recordsPath: "",
		scope:       scopeUser,
		mapRecord:   onesignalAppRecord,
	},
	"devices": {
		resource:    "players",
		recordsPath: "players",
		scope:       scopeApp,
		needsAppID:  true,
		paginated:   true,
		mapRecord:   onesignalDeviceRecord,
	},
	"notifications": {
		resource:    "notifications",
		recordsPath: "notifications",
		scope:       scopeApp,
		needsAppID:  true,
		paginated:   true,
		mapRecord:   onesignalNotificationRecord,
	},
	"outcomes": {
		resource:    "apps/{app_id}/outcomes",
		recordsPath: "outcomes",
		scope:       scopeApp,
		mapRecord:   onesignalOutcomeRecord,
	},
}

// onesignalStreams returns the connector's published stream catalog.
func onesignalStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "apps",
			Description: "OneSignal applications visible to the user auth key.",
			PrimaryKey:  []string{"id"},
			Fields:      onesignalAppFields(),
		},
		{
			Name:         "devices",
			Description:  "OneSignal devices (subscriptions / players) for the configured app.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       onesignalDeviceFields(),
		},
		{
			Name:         "notifications",
			Description:  "OneSignal notifications (messages) for the configured app.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"queued_at"},
			Fields:       onesignalNotificationFields(),
		},
		{
			Name:        "outcomes",
			Description: "OneSignal outcome metrics for the configured app.",
			PrimaryKey:  []string{"id"},
			Fields:      onesignalOutcomeFields(),
		},
	}
}

func onesignalAppFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "players", Type: "integer"},
		{Name: "messageable_players", Type: "integer"},
		{Name: "organization_id", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func onesignalDeviceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "identifier", Type: "string"},
		{Name: "device_type", Type: "integer"},
		{Name: "device_os", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "timezone", Type: "integer"},
		{Name: "game_version", Type: "string"},
		{Name: "session_count", Type: "integer"},
		{Name: "external_user_id", Type: "string"},
		{Name: "invalid_identifier", Type: "boolean"},
		{Name: "last_active", Type: "integer"},
		{Name: "created_at", Type: "integer"},
	}
}

func onesignalNotificationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "app_id", Type: "string"},
		{Name: "successful", Type: "integer"},
		{Name: "failed", Type: "integer"},
		{Name: "errored", Type: "integer"},
		{Name: "converted", Type: "integer"},
		{Name: "remaining", Type: "integer"},
		{Name: "queued_at", Type: "integer"},
		{Name: "send_after", Type: "integer"},
		{Name: "completed_at", Type: "integer"},
		{Name: "canceled", Type: "boolean"},
	}
}

func onesignalOutcomeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "value", Type: "number"},
		{Name: "aggregation", Type: "string"},
	}
}

func onesignalAppRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"name":                item["name"],
		"players":             item["players"],
		"messageable_players": item["messageable_players"],
		"organization_id":     item["organization_id"],
		"created_at":          item["created_at"],
		"updated_at":          item["updated_at"],
	}
}

func onesignalDeviceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"identifier":         item["identifier"],
		"device_type":        item["device_type"],
		"device_os":          item["device_os"],
		"language":           item["language"],
		"timezone":           item["timezone"],
		"game_version":       item["game_version"],
		"session_count":      item["session_count"],
		"external_user_id":   item["external_user_id"],
		"invalid_identifier": item["invalid_identifier"],
		"last_active":        item["last_active"],
		"created_at":         item["created_at"],
	}
}

func onesignalNotificationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"app_id":       item["app_id"],
		"successful":   item["successful"],
		"failed":       item["failed"],
		"errored":      item["errored"],
		"converted":    item["converted"],
		"remaining":    item["remaining"],
		"queued_at":    item["queued_at"],
		"send_after":   item["send_after"],
		"completed_at": item["completed_at"],
		"canceled":     item["canceled"],
	}
}

func onesignalOutcomeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"value":       item["value"],
		"aggregation": item["aggregation"],
	}
}
