package twilio

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the account-scoped Twilio resource path
// segment (relative to /Accounts/{sid}) it reads from, the JSON key holding the
// records array in the list response, and the record mapper.
type streamEndpoint struct {
	// resource is the path segment after /Accounts/{sid}/, e.g. "Messages.json".
	resource string
	// recordsKey is the JSON object key holding the records array, e.g. "messages".
	recordsKey string
	// mapRecord flattens a raw Twilio object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// twilioStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in twilioStreams; the read path
// is fully data-driven from this table.
var twilioStreamEndpoints = map[string]streamEndpoint{
	"messages":      {resource: "Messages.json", recordsKey: "messages", mapRecord: twilioMessageRecord},
	"calls":         {resource: "Calls.json", recordsKey: "calls", mapRecord: twilioCallRecord},
	"recordings":    {resource: "Recordings.json", recordsKey: "recordings", mapRecord: twilioRecordingRecord},
	"conferences":   {resource: "Conferences.json", recordsKey: "conferences", mapRecord: twilioConferenceRecord},
	"usage_records": {resource: "Usage/Records.json", recordsKey: "usage_records", mapRecord: twilioUsageRecordRecord},
}

// twilioStreams returns the connector's published stream catalog. Every Twilio
// resource exposes a string "sid" primary key; incremental streams cursor on a
// date field (date_sent / date_created / start_date).
func twilioStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "messages",
			Description:  "Twilio SMS/MMS messages.",
			PrimaryKey:   []string{"sid"},
			CursorFields: []string{"date_sent"},
			Fields:       twilioMessageFields(),
		},
		{
			Name:         "calls",
			Description:  "Twilio voice calls.",
			PrimaryKey:   []string{"sid"},
			CursorFields: []string{"start_time"},
			Fields:       twilioCallFields(),
		},
		{
			Name:         "recordings",
			Description:  "Twilio call recordings.",
			PrimaryKey:   []string{"sid"},
			CursorFields: []string{"date_created"},
			Fields:       twilioRecordingFields(),
		},
		{
			Name:         "conferences",
			Description:  "Twilio conferences.",
			PrimaryKey:   []string{"sid"},
			CursorFields: []string{"date_created"},
			Fields:       twilioConferenceFields(),
		},
		{
			Name:         "usage_records",
			Description:  "Twilio account usage records.",
			PrimaryKey:   []string{"category"},
			CursorFields: []string{"start_date"},
			Fields:       twilioUsageRecordFields(),
		},
	}
}

func twilioMessageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "sid", Type: "string"},
		{Name: "account_sid", Type: "string"},
		{Name: "messaging_service_sid", Type: "string"},
		{Name: "date_created", Type: "string"},
		{Name: "date_sent", Type: "string"},
		{Name: "date_updated", Type: "string"},
		{Name: "from", Type: "string"},
		{Name: "to", Type: "string"},
		{Name: "body", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "direction", Type: "string"},
		{Name: "num_segments", Type: "string"},
		{Name: "num_media", Type: "string"},
		{Name: "price", Type: "string"},
		{Name: "price_unit", Type: "string"},
		{Name: "error_code", Type: "string"},
		{Name: "error_message", Type: "string"},
	}
}

func twilioCallFields() []connectors.Field {
	return []connectors.Field{
		{Name: "sid", Type: "string"},
		{Name: "account_sid", Type: "string"},
		{Name: "date_created", Type: "string"},
		{Name: "date_updated", Type: "string"},
		{Name: "start_time", Type: "string"},
		{Name: "end_time", Type: "string"},
		{Name: "from", Type: "string"},
		{Name: "to", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "direction", Type: "string"},
		{Name: "duration", Type: "string"},
		{Name: "price", Type: "string"},
		{Name: "price_unit", Type: "string"},
	}
}

func twilioRecordingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "sid", Type: "string"},
		{Name: "account_sid", Type: "string"},
		{Name: "call_sid", Type: "string"},
		{Name: "date_created", Type: "string"},
		{Name: "date_updated", Type: "string"},
		{Name: "start_time", Type: "string"},
		{Name: "duration", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "channels", Type: "integer"},
		{Name: "source", Type: "string"},
		{Name: "price", Type: "string"},
		{Name: "price_unit", Type: "string"},
	}
}

func twilioConferenceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "sid", Type: "string"},
		{Name: "account_sid", Type: "string"},
		{Name: "friendly_name", Type: "string"},
		{Name: "date_created", Type: "string"},
		{Name: "date_updated", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "region", Type: "string"},
	}
}

func twilioUsageRecordFields() []connectors.Field {
	return []connectors.Field{
		{Name: "account_sid", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "start_date", Type: "string"},
		{Name: "end_date", Type: "string"},
		{Name: "count", Type: "string"},
		{Name: "count_unit", Type: "string"},
		{Name: "usage", Type: "string"},
		{Name: "usage_unit", Type: "string"},
		{Name: "price", Type: "string"},
		{Name: "price_unit", Type: "string"},
	}
}

func twilioMessageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"sid":                   item["sid"],
		"account_sid":           item["account_sid"],
		"messaging_service_sid": item["messaging_service_sid"],
		"date_created":          item["date_created"],
		"date_sent":             item["date_sent"],
		"date_updated":          item["date_updated"],
		"from":                  item["from"],
		"to":                    item["to"],
		"body":                  item["body"],
		"status":                item["status"],
		"direction":             item["direction"],
		"num_segments":          item["num_segments"],
		"num_media":             item["num_media"],
		"price":                 item["price"],
		"price_unit":            item["price_unit"],
		"error_code":            item["error_code"],
		"error_message":         item["error_message"],
	}
}

func twilioCallRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"sid":          item["sid"],
		"account_sid":  item["account_sid"],
		"date_created": item["date_created"],
		"date_updated": item["date_updated"],
		"start_time":   item["start_time"],
		"end_time":     item["end_time"],
		"from":         item["from"],
		"to":           item["to"],
		"status":       item["status"],
		"direction":    item["direction"],
		"duration":     item["duration"],
		"price":        item["price"],
		"price_unit":   item["price_unit"],
	}
}

func twilioRecordingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"sid":          item["sid"],
		"account_sid":  item["account_sid"],
		"call_sid":     item["call_sid"],
		"date_created": item["date_created"],
		"date_updated": item["date_updated"],
		"start_time":   item["start_time"],
		"duration":     item["duration"],
		"status":       item["status"],
		"channels":     item["channels"],
		"source":       item["source"],
		"price":        item["price"],
		"price_unit":   item["price_unit"],
	}
}

func twilioConferenceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"sid":           item["sid"],
		"account_sid":   item["account_sid"],
		"friendly_name": item["friendly_name"],
		"date_created":  item["date_created"],
		"date_updated":  item["date_updated"],
		"status":        item["status"],
		"region":        item["region"],
	}
}

func twilioUsageRecordRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"account_sid": item["account_sid"],
		"category":    item["category"],
		"description": item["description"],
		"start_date":  item["start_date"],
		"end_date":    item["end_date"],
		"count":       item["count"],
		"count_unit":  item["count_unit"],
		"usage":       item["usage"],
		"usage_unit":  item["usage_unit"],
		"price":       item["price"],
		"price_unit":  item["price_unit"],
	}
}
