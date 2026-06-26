package opinionstage

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream to an Opinion Stage endpoint and record mapper.
// responses and questions are substreams under each item id.
type streamEndpoint struct {
	path        string
	subresource string
	mapRecord   func(map[string]any, string) connectors.Record
}

func (e streamEndpoint) isSubstream() bool { return e.subresource != "" }

var opinionStageStreamEndpoints = map[string]streamEndpoint{
	"items":     {path: "/api/v2/items", mapRecord: opinionStageItemRecord},
	"responses": {subresource: "responses", mapRecord: opinionStageResponseRecord},
	"questions": {subresource: "questions", mapRecord: opinionStageQuestionRecord},
}

func opinionStageStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "items",
			Description: "Opinion Stage items such as polls, quizzes, and forms.",
			PrimaryKey:  []string{"id"},
			Fields:      opinionStageItemFields(),
		},
		{
			Name:        "responses",
			Description: "Opinion Stage responses, read per item.",
			PrimaryKey:  []string{"id"},
			Fields:      opinionStageResponseFields(),
		},
		{
			Name:        "questions",
			Description: "Opinion Stage questions, read per item.",
			PrimaryKey:  []string{"id"},
			Fields:      opinionStageQuestionFields(),
		},
	}
}

func opinionStageItemFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "embed", Type: "object"},
		{Name: "created", Type: "timestamp"},
		{Name: "modified", Type: "timestamp"},
		{Name: "links", Type: "object"},
		{Name: "relationships", Type: "object"},
	}
}

func opinionStageResponseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "item_id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "answers", Type: "array"},
		{Name: "result", Type: "object"},
		{Name: "result_title", Type: "string"},
		{Name: "result_text", Type: "string"},
		{Name: "created", Type: "timestamp"},
		{Name: "duration", Type: "number"},
		{Name: "utm", Type: "object"},
		{Name: "links", Type: "object"},
	}
}

func opinionStageQuestionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "item_id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "kind", Type: "string"},
		{Name: "lead", Type: "boolean"},
		{Name: "created", Type: "timestamp"},
		{Name: "modified", Type: "timestamp"},
	}
}

func opinionStageItemRecord(item map[string]any, _ string) connectors.Record {
	rec := jsonAPIRecord(item)
	return connectors.Record{
		"id":            rec["id"],
		"type":          rec["type"],
		"title":         rec["title"],
		"status":        rec["status"],
		"embed":         rec["embed"],
		"created":       rec["created"],
		"modified":      rec["modified"],
		"links":         rec["links"],
		"relationships": rec["relationships"],
	}
}

func opinionStageItemRecordForParent(item map[string]any) connectors.Record {
	return opinionStageItemRecord(item, "")
}

func opinionStageResponseRecord(item map[string]any, itemID string) connectors.Record {
	rec := jsonAPIRecord(item)
	result, _ := rec["result"].(map[string]any)
	timings, _ := rec["timings"].(map[string]any)
	return connectors.Record{
		"id":           rec["id"],
		"item_id":      itemID,
		"type":         rec["type"],
		"answers":      rec["answers"],
		"result":       rec["result"],
		"result_title": result["title"],
		"result_text":  result["text"],
		"created":      rec["created"],
		"duration":     timings["duration"],
		"utm":          rec["utm"],
		"links":        rec["links"],
	}
}

func opinionStageQuestionRecord(item map[string]any, itemID string) connectors.Record {
	rec := jsonAPIRecord(item)
	return connectors.Record{
		"id":       rec["id"],
		"item_id":  itemID,
		"type":     rec["type"],
		"title":    rec["title"],
		"kind":     rec["kind"],
		"lead":     rec["lead"],
		"created":  rec["created"],
		"modified": rec["modified"],
	}
}

// jsonAPIRecord flattens {id,type,attributes:{...}} and copies JSON:API links
// and relationships. attributes.timestamps.created/modified are surfaced as
// created/modified for stable field names across Opinion Stage streams.
func jsonAPIRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":            item["id"],
		"type":          item["type"],
		"links":         item["links"],
		"relationships": item["relationships"],
	}
	attrs := attributesOf(item)
	for k, v := range attrs {
		rec[k] = v
	}
	if timestamps, ok := attrs["timestamps"].(map[string]any); ok {
		rec["created"] = timestamps["created"]
		rec["modified"] = timestamps["modified"]
	}
	return rec
}

func attributesOf(item map[string]any) map[string]any {
	if attrs, ok := item["attributes"].(map[string]any); ok {
		return attrs
	}
	return map[string]any{}
}
