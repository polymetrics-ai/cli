package zendesktalk

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to its Zendesk Talk API resource path
// (relative to base_url), the JSON key holding its record array, and the mapper
// that flattens its objects into a connectors.Record.
type streamEndpoint struct {
	// resource is the Talk endpoint path segment (e.g. "phone_numbers").
	resource string
	// recordsKey is the top-level JSON key holding the array of objects
	// (e.g. "phone_numbers"). The Talk API wraps lists in a named key.
	recordsKey string
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// zendeskTalkStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in zendeskTalkStreams;
// the read path is fully data-driven from this table.
var zendeskTalkStreamEndpoints = map[string]streamEndpoint{
	"phone_numbers":       {resource: "phone_numbers", recordsKey: "phone_numbers", mapRecord: phoneNumberRecord},
	"greetings":           {resource: "greetings", recordsKey: "greetings", mapRecord: greetingRecord},
	"greeting_categories": {resource: "greeting_categories", recordsKey: "greeting_categories", mapRecord: greetingCategoryRecord},
	"ivrs":                {resource: "ivrs", recordsKey: "ivrs", mapRecord: ivrRecord},
	"agents_activity":     {resource: "stats/agents_activity", recordsKey: "agents_activity", mapRecord: agentActivityRecord},
}

// zendeskTalkStreams returns the connector's published stream catalog.
func zendeskTalkStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "phone_numbers",
			Description:  "Zendesk Talk phone numbers provisioned for the account.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       phoneNumberFields(),
		},
		{
			Name:        "greetings",
			Description: "Zendesk Talk greetings (recorded audio prompts).",
			PrimaryKey:  []string{"id"},
			Fields:      greetingFields(),
		},
		{
			Name:        "greeting_categories",
			Description: "Categories that group Zendesk Talk greetings.",
			PrimaryKey:  []string{"id"},
			Fields:      greetingCategoryFields(),
		},
		{
			Name:        "ivrs",
			Description: "Zendesk Talk interactive voice response (IVR) trees.",
			PrimaryKey:  []string{"id"},
			Fields:      ivrFields(),
		},
		{
			Name:        "agents_activity",
			Description: "Per-agent Zendesk Talk activity statistics.",
			PrimaryKey:  []string{"agent_id"},
			Fields:      agentActivityFields(),
		},
	}
}

func phoneNumberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "number", Type: "string"},
		{Name: "display_number", Type: "string"},
		{Name: "nickname", Type: "string"},
		{Name: "country_code", Type: "string"},
		{Name: "toll_free", Type: "boolean"},
		{Name: "voice_enabled", Type: "boolean"},
		{Name: "sms_enabled", Type: "boolean"},
		{Name: "recorded", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
	}
}

func greetingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "category_id", Type: "integer"},
		{Name: "default", Type: "boolean"},
		{Name: "active", Type: "boolean"},
		{Name: "audio_name", Type: "string"},
		{Name: "audio_url", Type: "string"},
		{Name: "has_sub_settings", Type: "boolean"},
	}
}

func greetingCategoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
	}
}

func ivrFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "phone_number_ids", Type: "array"},
		{Name: "phone_number_names", Type: "array"},
	}
}

func agentActivityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "agent_id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "avatar_url", Type: "string"},
		{Name: "agent_state", Type: "string"},
		{Name: "call_status", Type: "string"},
		{Name: "available_time", Type: "integer"},
		{Name: "away_time", Type: "integer"},
		{Name: "online_time", Type: "integer"},
		{Name: "calls_accepted", Type: "integer"},
		{Name: "calls_denied", Type: "integer"},
		{Name: "calls_missed", Type: "integer"},
		{Name: "forwarding_number", Type: "string"},
		{Name: "total_call_duration", Type: "integer"},
		{Name: "total_talk_time", Type: "integer"},
		{Name: "total_wrap_up_time", Type: "integer"},
		{Name: "via", Type: "string"},
	}
}

func phoneNumberRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"number":         item["number"],
		"display_number": item["display_number"],
		"nickname":       item["nickname"],
		"country_code":   item["country_code"],
		"toll_free":      item["toll_free"],
		"voice_enabled":  item["voice_enabled"],
		"sms_enabled":    item["sms_enabled"],
		"recorded":       item["recorded"],
		"created_at":     item["created_at"],
	}
}

func greetingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"category_id":      item["category_id"],
		"default":          item["default"],
		"active":           item["active"],
		"audio_name":       item["audio_name"],
		"audio_url":        item["audio_url"],
		"has_sub_settings": item["has_sub_settings"],
	}
}

func greetingCategoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   item["id"],
		"name": item["name"],
	}
}

func ivrRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"name":               item["name"],
		"phone_number_ids":   item["phone_number_ids"],
		"phone_number_names": item["phone_number_names"],
	}
}

func agentActivityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"agent_id":            item["agent_id"],
		"name":                item["name"],
		"avatar_url":          item["avatar_url"],
		"agent_state":         item["agent_state"],
		"call_status":         item["call_status"],
		"available_time":      item["available_time"],
		"away_time":           item["away_time"],
		"online_time":         item["online_time"],
		"calls_accepted":      item["calls_accepted"],
		"calls_denied":        item["calls_denied"],
		"calls_missed":        item["calls_missed"],
		"forwarding_number":   item["forwarding_number"],
		"total_call_duration": item["total_call_duration"],
		"total_talk_time":     item["total_talk_time"],
		"total_wrap_up_time":  item["total_wrap_up_time"],
		"via":                 item["via"],
	}
}
