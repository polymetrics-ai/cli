package datascope

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the DataScope API resource path (relative
// to base_url) it reads from, the record mapper that flattens its objects, and
// whether it accepts the start/end datetime window (DataScope's date filter on
// the time-series streams).
type streamEndpoint struct {
	// resource is the DataScope endpoint path (e.g. "locations", "v2/answers").
	resource string
	// mapRecord flattens a raw DataScope object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// windowed is true when the endpoint accepts start/end request parameters
	// derived from the start_date config (answers, notifications).
	windowed bool
}

// datascopeStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in datascopeStreams; the read
// path is fully data-driven from this table.
var datascopeStreamEndpoints = map[string]streamEndpoint{
	"locations":     {resource: "locations", mapRecord: datascopeLocationRecord},
	"answers":       {resource: "v2/answers", mapRecord: datascopeAnswerRecord, windowed: true},
	"lists":         {resource: "metadata_objects", mapRecord: datascopeListRecord},
	"notifications": {resource: "notifications", mapRecord: datascopeNotificationRecord, windowed: true},
}

// datascopeStreams returns the connector's published stream catalog, mirroring
// the upstream DataScope source manifest (locations, answers, lists,
// notifications).
func datascopeStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "locations",
			Description:  "DataScope locations (sites) configured for the account.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       datascopeLocationFields(),
		},
		{
			Name:         "answers",
			Description:  "DataScope form answers (submitted form responses).",
			PrimaryKey:   []string{"form_answer_id"},
			CursorFields: []string{"created_at"},
			Fields:       datascopeAnswerFields(),
		},
		{
			Name:         "lists",
			Description:  "DataScope metadata list objects.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       datascopeListFields(),
		},
		{
			Name:         "notifications",
			Description:  "DataScope notifications.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"created_at"},
			Fields:       datascopeNotificationFields(),
		},
	}
}

func datascopeLocationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "address", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "latitude", Type: "number"},
		{Name: "longitude", Type: "number"},
		{Name: "region", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "company_code", Type: "string"},
		{Name: "company_name", Type: "string"},
	}
}

func datascopeAnswerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "form_answer_id", Type: "integer"},
		{Name: "form_id", Type: "integer"},
		{Name: "form_name", Type: "string"},
		{Name: "form_state", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "user_name", Type: "string"},
		{Name: "user_identifier", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "latitude", Type: "number"},
		{Name: "longitude", Type: "number"},
	}
}

func datascopeListFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "attribute1", Type: "string"},
		{Name: "attribute2", Type: "string"},
		{Name: "list_id", Type: "integer"},
		{Name: "account_id", Type: "integer"},
		{Name: "code", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func datascopeNotificationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "type", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "form_name", Type: "string"},
		{Name: "form_code", Type: "string"},
		{Name: "user", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func datascopeLocationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"description":  item["description"],
		"code":         item["code"],
		"address":      item["address"],
		"city":         item["city"],
		"country":      item["country"],
		"latitude":     item["latitude"],
		"longitude":    item["longitude"],
		"region":       item["region"],
		"phone":        item["phone"],
		"company_code": item["company_code"],
		"company_name": item["company_name"],
	}
}

func datascopeAnswerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"form_answer_id":  item["form_answer_id"],
		"form_id":         item["form_id"],
		"form_name":       item["form_name"],
		"form_state":      item["form_state"],
		"code":            item["code"],
		"user_name":       item["user_name"],
		"user_identifier": item["user_identifier"],
		"created_at":      item["created_at"],
		"latitude":        item["latitude"],
		"longitude":       item["longitude"],
	}
}

func datascopeListRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"description": item["description"],
		"attribute1":  item["attribute1"],
		"attribute2":  item["attribute2"],
		"list_id":     item["list_id"],
		"account_id":  item["account_id"],
		"code":        item["code"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}

func datascopeNotificationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"type":       item["type"],
		"url":        item["url"],
		"form_name":  item["form_name"],
		"form_code":  item["form_code"],
		"user":       item["user"],
		"created_at": item["created_at"],
	}
}
