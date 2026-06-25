package delighted

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Delighted API resource path (relative
// to base_url) it reads from, the record mapper that flattens its objects, and
// whether the resource is a single object (metrics) rather than a paginated list.
type streamEndpoint struct {
	// resource is the Delighted endpoint path segment, e.g. "survey_responses.json".
	resource string
	// mapRecord flattens a raw Delighted object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// single marks endpoints that return one JSON object instead of an array
	// (metrics). Such streams are read with a single request, no pagination.
	single bool
	// supportsSince marks endpoints that accept the since/updated_since unix
	// timestamp filters (survey_responses, metrics).
	supportsSince bool
}

// delightedStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in delightedStreams; the read
// path is fully data-driven from this table.
var delightedStreamEndpoints = map[string]streamEndpoint{
	"survey_responses": {resource: "survey_responses.json", mapRecord: surveyResponseRecord, supportsSince: true},
	"people":           {resource: "people.json", mapRecord: personRecord},
	"bounces":          {resource: "bounces.json", mapRecord: bounceRecord},
	"unsubscribes":     {resource: "unsubscribes.json", mapRecord: unsubscribeRecord},
	"metrics":          {resource: "metrics.json", mapRecord: metricsRecord, single: true, supportsSince: true},
}

// delightedStreams returns the connector's published stream catalog.
func delightedStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "survey_responses",
			Description:  "Delighted survey responses (NPS/CSAT/CES scores and comments).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       surveyResponseFields(),
		},
		{
			Name:        "people",
			Description: "People (customers) tracked in Delighted.",
			PrimaryKey:  []string{"id"},
			Fields:      personFields(),
		},
		{
			Name:        "bounces",
			Description: "People whose survey emails bounced.",
			PrimaryKey:  []string{"person_id"},
			Fields:      bounceFields(),
		},
		{
			Name:        "unsubscribes",
			Description: "People who unsubscribed from Delighted surveys.",
			PrimaryKey:  []string{"person_id"},
			Fields:      unsubscribeFields(),
		},
		{
			Name:        "metrics",
			Description: "Aggregate Delighted metrics (NPS, promoter/detractor counts).",
			PrimaryKey:  []string{},
			Fields:      metricsFields(),
		},
	}
}

func surveyResponseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "person", Type: "string"},
		{Name: "survey_type", Type: "string"},
		{Name: "score", Type: "integer"},
		{Name: "comment", Type: "string"},
		{Name: "permalink", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "updated_at", Type: "integer"},
		{Name: "person_properties", Type: "object"},
		{Name: "notes", Type: "array"},
		{Name: "tags", Type: "array"},
	}
}

func personFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone_number", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "last_sent_at", Type: "integer"},
		{Name: "last_responded_at", Type: "integer"},
		{Name: "next_survey_scheduled_at", Type: "integer"},
	}
}

func bounceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "person_id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "bounced_at", Type: "integer"},
	}
}

func unsubscribeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "person_id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "unsubscribed_at", Type: "integer"},
	}
}

func metricsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "nps", Type: "integer"},
		{Name: "promoter_count", Type: "integer"},
		{Name: "promoter_percent", Type: "number"},
		{Name: "passive_count", Type: "integer"},
		{Name: "passive_percent", Type: "number"},
		{Name: "detractor_count", Type: "integer"},
		{Name: "detractor_percent", Type: "number"},
		{Name: "response_count", Type: "integer"},
	}
}

func surveyResponseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"person":            item["person"],
		"survey_type":       item["survey_type"],
		"score":             item["score"],
		"comment":           item["comment"],
		"permalink":         item["permalink"],
		"created_at":        item["created_at"],
		"updated_at":        item["updated_at"],
		"person_properties": item["person_properties"],
		"notes":             item["notes"],
		"tags":              item["tags"],
	}
}

func personRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                       item["id"],
		"name":                     item["name"],
		"email":                    item["email"],
		"phone_number":             item["phone_number"],
		"created_at":               item["created_at"],
		"last_sent_at":             item["last_sent_at"],
		"last_responded_at":        item["last_responded_at"],
		"next_survey_scheduled_at": item["next_survey_scheduled_at"],
	}
}

func bounceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"person_id":  item["person_id"],
		"email":      item["email"],
		"name":       item["name"],
		"bounced_at": item["bounced_at"],
	}
}

func unsubscribeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"person_id":       item["person_id"],
		"email":           item["email"],
		"name":            item["name"],
		"unsubscribed_at": item["unsubscribed_at"],
	}
}

func metricsRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"nps":               item["nps"],
		"promoter_count":    item["promoter_count"],
		"promoter_percent":  item["promoter_percent"],
		"passive_count":     item["passive_count"],
		"passive_percent":   item["passive_percent"],
		"detractor_count":   item["detractor_count"],
		"detractor_percent": item["detractor_percent"],
		"response_count":    item["response_count"],
	}
}
